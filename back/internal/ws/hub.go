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
			} else if m.Type == "UPDATE_CONFIG" {
				h.handleUpdateConfig(m)
			} else if m.Type == "UPDATE_ROOM_MAX_PLAYERS" {
				h.handleUpdateRoomMaxPlayers(m)
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

func (h *Hub) handleUpdateConfig(m *Message) {
	ctx := context.Background()
	r, err := h.roomRepo.GetByCode(ctx, m.Client.roomCode)
	if err != nil || r == nil {
		return
	}

	// Only host can update config
	if m.Client.userID != r.HostID {
		return
	}

	// Broadcast to everyone in the room
	msg, _ := json.Marshal(map[string]any{
		"type": "CONFIG_UPDATED",
		"payload": m.Payload,
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

func (h *Hub) handleUpdateRoomMaxPlayers(m *Message) {
	ctx := context.Background()
	r, err := h.roomRepo.GetByCode(ctx, m.Client.roomCode)
	if err != nil || r == nil {
		return
	}

	// Only host can update
	if m.Client.userID != r.HostID {
		return
	}

	maxPlayersFloat, ok := m.Payload["maxPlayers"].(float64)
	if !ok {
		return
	}
	maxPlayers := int(maxPlayersFloat)

	// Update in DB
	err = h.roomRepo.UpdateMaxPlayers(ctx, r.ID, maxPlayers)
	if err != nil {
		return
	}

	// Also update the local room cache for future reference
	r.MaxPlayers = maxPlayers

	// Broadcast
	msg, _ := json.Marshal(map[string]any{
		"type": "ROOM_UPDATED",
		"payload": map[string]any{
			"room": r,
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
	if err != nil || r == nil {
		fmt.Printf("Room %s not found or deleted, kicking all players\n", roomCode)
		
		msg, _ := json.Marshal(map[string]any{
			"type": "ROOM_CLOSED",
			"payload": map[string]any{
				"reason": "Le salon a été fermé ou supprimé.",
			},
		})

		set := h.clientsByRoom[roomCode]
		for c := range set {
			c.send <- msg
			// We don't removeClient(c) immediately to let the message be sent
			// The writePump will close the connection when the send channel is closed or on error
		}
		
		// Wait a bit or let unregister handle it? 
		// Actually, let's just remove them.
		for c := range set {
			h.removeClient(c)
		}
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

