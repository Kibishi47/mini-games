package ws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kibishi47/mini-games/back/internal/domain/room"
	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	goredis "github.com/redis/go-redis/v9"
)

type Message struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
	Client  *Client        `json:"-"`
}

type Hub struct {
	register   chan *Client
	unregister chan *Client
	message    chan *Message

	clients       map[*Client]struct{}
	clientsByRoom map[string]map[*Client]struct{}
	redis         *goredis.Client
	roomRepo      room.Repository
	userRepo      user.Repository
}

func NewHub(redis *goredis.Client, roomRepo room.Repository, userRepo user.Repository) *Hub {
	return &Hub{
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		message:       make(chan *Message),
		clients:       make(map[*Client]struct{}),
		clientsByRoom: make(map[string]map[*Client]struct{}),
		redis:         redis,
		roomRepo:      roomRepo,
		userRepo:      userRepo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}
			if h.clientsByRoom[c.roomCode] == nil {
				h.clientsByRoom[c.roomCode] = make(map[*Client]struct{})
			}
			h.clientsByRoom[c.roomCode][c] = struct{}{}
			h.updateRoomPlayers(c.roomCode)

		case c := <-h.unregister:
			roomCode := c.roomCode
			h.removeClient(c)
			h.updateRoomPlayers(roomCode)

		case m := <-h.message:
			if m.Type == "SELECT_GAME" {
				h.handleSelectGame(m)
			}
		}
	}
}

func (h *Hub) handleSelectGame(m *Message) {
	ctx := context.Background()
	r, err := h.roomRepo.GetByCode(ctx, m.Client.roomCode)
	if err != nil || r == nil {
		return
	}

	// Only host can select game
	if m.Client.userID != r.HostID {
		return
	}

	gameID, ok := m.Payload["gameId"].(string)
	if !ok {
		return
	}

	// Broadcast to everyone in the room
	msg, _ := json.Marshal(map[string]any{
		"type": "GAME_SELECTED",
		"payload": map[string]any{
			"gameId": gameID,
		},
	})

	set := h.clientsByRoom[m.Client.roomCode]
	for c := range set {
		select {
		case c.send <- msg:
		default:
			h.removeClient(c)
		}
	}
}

func (h *Hub) removeClient(c *Client) {
	if _, ok := h.clients[c]; !ok {
		return
	}
	delete(h.clients, c)

	if set := h.clientsByRoom[c.roomCode]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clientsByRoom, c.roomCode)
		}
	}

	close(c.send)
}

func (h *Hub) updateRoomPlayers(roomCode string) {
	if roomCode == "" {
		return
	}

	ctx := context.Background()
	r, err := h.roomRepo.GetByCode(ctx, roomCode)
	if err != nil {
		fmt.Printf("Error getting room %s: %v\n", roomCode, err)
		return
	}

	type PlayerInfo struct {
		Username string `json:"username"`
		IsHost   bool   `json:"isHost"`
	}

	// 1. Get all players in this room from our local map
	set := h.clientsByRoom[roomCode]
	
	// Track unique users in room
	playersMap := make(map[string]PlayerInfo)
	usernames := []string{}
	for c := range set {
		// Fetch fresh username/displayName from DB to handle renames
		u, err := h.userRepo.GetByID(ctx, c.userID)
		nameToShow := c.username
		if err == nil && u != nil {
			if u.DisplayName != nil && *u.DisplayName != "" {
				nameToShow = *u.DisplayName
			} else {
				nameToShow = u.Username
			}
			// Update client's cached username too (we'll use this for consistency)
			c.username = nameToShow
		}

		if _, exists := playersMap[nameToShow]; !exists {
			playersMap[nameToShow] = PlayerInfo{
				Username: nameToShow,
				IsHost:   c.userID == r.HostID,
			}
			usernames = append(usernames, nameToShow)
		}
	}

	playersList := []PlayerInfo{}
	for _, name := range usernames {
		playersList = append(playersList, playersMap[name])
	}

	// 2. Store in Redis (list of usernames for now to keep it compatible)
	redisKey := fmt.Sprintf("room:%s:players", roomCode)
	
	h.redis.Del(ctx, redisKey)
	if len(usernames) > 0 {
		h.redis.SAdd(ctx, redisKey, usernames)
	}

	// 3. Broadcast to all clients in the room
	msg, _ := json.Marshal(map[string]any{
		"type": "PLAYER_LIST",
		"payload": map[string]any{
			"players": playersList,
		},
	})

	for c := range set {
		select {
		case c.send <- msg:
		default:
			h.removeClient(c)
		}
	}
}

