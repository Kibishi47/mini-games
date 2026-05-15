package ws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kibishi47/mini-games/back/internal/domain/room"
	"github.com/Kibishi47/mini-games/back/internal/domain/session"
	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	goredis "github.com/redis/go-redis/v9"
	"time"
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
	roomRepo       room.Repository
	userRepo       user.Repository
	sessionService session.Service
}

func NewHub(redis *goredis.Client, roomRepo room.Repository, userRepo user.Repository, sessionService session.Service) *Hub {
	return &Hub{
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		message:        make(chan *Message),
		clients:        make(map[*Client]struct{}),
		clientsByRoom:  make(map[string]map[*Client]struct{}),
		redis:          redis,
		roomRepo:       roomRepo,
		userRepo:       userRepo,
		sessionService: sessionService,
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

			// Send chat history to the newly connected client
			ctx := context.Background()
			chatKey := "room:" + c.roomCode + ":chat"
			chatStrs, _ := h.redis.LRange(ctx, chatKey, 0, -1).Result()
			
			if len(chatStrs) > 0 {
				var history []map[string]any
				for _, str := range chatStrs {
					var msg map[string]any
					if err := json.Unmarshal([]byte(str), &msg); err == nil {
						history = append(history, msg)
					}
				}
				
				historyMsg, _ := json.Marshal(map[string]any{
					"type": "CHAT_HISTORY",
					"payload": history,
				})
				
				c.send <- historyMsg
			}

			// Send session info if game is running
			if r, err := h.roomRepo.GetByCode(ctx, c.roomCode); err == nil && r != nil {
				if r.Status == room.StatusRunning {
					session, err := h.sessionService.GetActiveSessionByRoomID(ctx, r.ID)
					if err == nil && session != nil {
						sessionMsg, _ := json.Marshal(map[string]any{
							"type": "GAME_STARTED",
							"payload": map[string]any{
								"session": session,
								"room":    r,
							},
						})
						c.send <- sessionMsg
					}
				} else {
					// Send current game selection and config if in lobby
					gameKey := "room:" + c.roomCode + ":game"
					configKey := "room:" + c.roomCode + ":config"
					
					selectedGame, _ := h.redis.Get(ctx, gameKey).Result()
					if selectedGame != "" {
						msg, _ := json.Marshal(map[string]any{
							"type": "GAME_SELECTED",
							"payload": map[string]any{
								"gameId": selectedGame,
							},
						})
						c.send <- msg
					}

					configStr, _ := h.redis.Get(ctx, configKey).Result()
					if configStr != "" {
						var config map[string]any
						if err := json.Unmarshal([]byte(configStr), &config); err == nil {
							msg, _ := json.Marshal(map[string]any{
								"type": "CONFIG_UPDATED",
								"payload": config,
							})
							c.send <- msg
						}
					}
				}
			}

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
			} else if m.Type == "START_GAME" {
				h.handleStartGame(m)
			} else if m.Type == "CHAT_MESSAGE" {
				h.handleChatMessage(m)
			} else if m.Type == "STOP_GAME" {
				h.handleStopGame(m)
			} else if m.Type == "KICK_PLAYER" {
				h.handleKickPlayer(m)
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

	// Persist in Redis
	gameKey := "room:" + m.Client.roomCode + ":game"
	h.redis.Set(ctx, gameKey, gameID, 24*time.Hour)

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

	// Persist in Redis
	configKey := "room:" + m.Client.roomCode + ":config"
	configBytes, _ := json.Marshal(m.Payload)
	h.redis.Set(ctx, configKey, configBytes, 24*time.Hour)

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

func (h *Hub) handleStartGame(m *Message) {
	ctx := context.Background()
	r, err := h.roomRepo.GetByCode(ctx, m.Client.roomCode)
	if err != nil || r == nil {
		return
	}

	// Only host can start
	if m.Client.userID != r.HostID {
		return
	}

	gameID, ok := m.Payload["gameId"].(string)
	if !ok {
		return
	}

	config, _ := m.Payload["config"].(map[string]any)

	// Update Room status
	h.roomRepo.UpdateStatus(ctx, r.ID, room.StatusRunning)
	r.Status = room.StatusRunning

	// Create Session
	session, err := h.sessionService.CreateSession(ctx, r.ID, gameID, config)
	if err != nil {
		return
	}

	// Broadcast
	msg, _ := json.Marshal(map[string]any{
		"type": "GAME_STARTED",
		"payload": map[string]any{
			"session": session,
			"room":    r,
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

func (h *Hub) handleChatMessage(m *Message) {
	ctx := context.Background()
	
	// Save to Redis (room:<code/>:chat)
	chatKey := "room:" + m.Client.roomCode + ":chat"
	
	chatMsg := map[string]any{
		"userId":   m.Client.userID.String(),
		"username": m.Client.username,
		"text":     m.Payload["text"],
		"system":   m.Payload["system"],
		"time":     m.Payload["time"],
	}
	
	chatBytes, _ := json.Marshal(chatMsg)
	h.redis.RPush(ctx, chatKey, chatBytes)
	h.redis.Expire(ctx, chatKey, 24*time.Hour) // Keep for 24h

	// Broadcast
	msg, _ := json.Marshal(map[string]any{
		"type": "CHAT_MESSAGE",
		"payload": chatMsg,
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

func (h *Hub) handleStopGame(m *Message) {
	ctx := context.Background()
	r, err := h.roomRepo.GetByCode(ctx, m.Client.roomCode)
	if err != nil || r == nil {
		return
	}

	// Only host can stop
	if m.Client.userID != r.HostID {
		return
	}

	// Get active session
	session, err := h.sessionService.GetActiveSessionByRoomID(ctx, r.ID)
	if err == nil && session != nil {
		// End session
		h.sessionService.EndSession(ctx, session.ID)
	}

	// Update Room status back to Lobby
	h.roomRepo.UpdateStatus(ctx, r.ID, room.StatusLobby)
	r.Status = room.StatusLobby

	// Broadcast
	msg, _ := json.Marshal(map[string]any{
		"type": "GAME_STOPPED",
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

func (h *Hub) handleKickPlayer(m *Message) {
	ctx := context.Background()
	r, err := h.roomRepo.GetByCode(ctx, m.Client.roomCode)
	if err != nil || r == nil {
		return
	}

	// Only host can kick
	if m.Client.userID != r.HostID {
		return
	}

	targetUsername, ok := m.Payload["username"].(string)
	if !ok {
		return
	}

	// Find the client to kick
	set := h.clientsByRoom[m.Client.roomCode]
	var targetClient *Client
	for c := range set {
		if c.username == targetUsername {
			// Cannot kick yourself
			if c.userID == r.HostID {
				return
			}
			targetClient = c
			break
		}
	}

	if targetClient != nil {
		// Send a kick message to the target
		kickMsg, _ := json.Marshal(map[string]any{
			"type": "KICKED",
			"payload": map[string]any{
				"reason": "Vous avez été exclu du salon par l'hôte.",
			},
		})
		targetClient.send <- kickMsg
		
		// REMOVE FROM REDIS FIRST
		redisKey := fmt.Sprintf("room:%s:players", m.Client.roomCode)
		h.redis.HDel(ctx, redisKey, targetClient.userID.String())

		// Then remove from Hub
		h.removeClient(targetClient)
		h.updateRoomPlayers(m.Client.roomCode)
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

		// SOURCE OF TRUTH: Check if user ID is in Redis Hash
		redisKey := fmt.Sprintf("room:%s:players", roomCode)
		exists, _ := h.redis.HExists(ctx, redisKey, c.userID.String()).Result()
		
		if !exists {
			// User was removed from Redis, kick them from WebSocket
			kickMsg, _ := json.Marshal(map[string]any{
				"type": "KICKED",
				"payload": map[string]any{
					"reason": "Vous n'êtes plus dans ce salon.",
				},
			})
			c.send <- kickMsg
			h.removeClient(c)
			continue
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

	// 3. Broadcast to all clients in the room (No more Redis overwrite here)
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

