package ws

import (
	"net/http"

	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	"github.com/Kibishi47/mini-games/back/internal/http/auth"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub      *Hub
	auth     auth.Service
	userRepo user.Repository
}

func NewHandler(hub *Hub, auth auth.Service, userRepo user.Repository) *Handler {
	return &Handler{
		hub:      hub,
		auth:     auth,
		userRepo: userRepo,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	userID, err := h.auth.Authenticate(r.Context(), rawToken)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	u, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	roomCode := r.URL.Query().Get("room")
	if roomCode == "" {
		http.Error(w, "missing room", http.StatusBadRequest)
		return
	}

	// Verify room exists
	room, err := h.hub.roomRepo.GetByCode(r.Context(), roomCode)
	if err != nil || room == nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	// Verify user is in Redis for this room (source of truth)
	redisKey := "room:" + roomCode + ":players"
	inRedis, _ := h.hub.redis.HExists(r.Context(), redisKey, userID.String()).Result()
	if !inRedis {
		http.Error(w, "you are not a member of this room", http.StatusForbidden)
		return
	}

	// Check max players limit (based on Redis Hash)
	playersCount, _ := h.hub.redis.HLen(r.Context(), redisKey).Result()
	if int(playersCount) > room.MaxPlayers {
		http.Error(w, "room is full", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := NewClient(h.hub, conn, userID, u.Username, roomCode)
	h.hub.register <- client

	go client.WritePump()
	client.ReadPump()
}
