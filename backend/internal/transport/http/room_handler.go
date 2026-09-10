package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nhooyr.io/websocket"

	"minigames-backend/internal/domain"
	"minigames-backend/internal/repository/redis"
	"minigames-backend/internal/service/room"
	"minigames-backend/internal/transport/ws"
)

type RoomHandler struct {
	roomService *room.RoomService
	roomRepo    *redis.RoomRepository
	hub         *ws.Hub
}

func NewRoomHandler(roomService *room.RoomService, roomRepo *redis.RoomRepository, hub *ws.Hub) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		roomRepo:    roomRepo,
		hub:         hub,
	}
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	jsonResponse(w, status, map[string]string{"error": msg})
}

// CreateRoom POST /api/rooms
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nickname string               `json:"nickname"`
		Mascot   string               `json:"mascot"`
		Color    string               `json:"color"`
		Settings *domain.RoomSettings `json:"settings"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	room, token, err := h.roomService.CreateRoom(r.Context(), body.Nickname, body.Mascot, body.Color, body.Settings)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"room":          room,
		"session_token": token,
	})
}

// GetRoom GET /api/rooms/{code}
func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	room, err := h.roomService.GetRoom(r.Context(), code)
	if err != nil {
		jsonError(w, http.StatusNotFound, "salle introuvable ou fermée")
		return
	}

	// Masquer le mot secret en cours de jeu
	roomSafe := *room
	if roomSafe.Status == domain.RoomStatusInGame {
		roomSafe.SecretWord = ""
	}

	jsonResponse(w, http.StatusOK, roomSafe)
}

// JoinRoom POST /api/rooms/{code}/join
func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	var body struct {
		Nickname string `json:"nickname"`
		Mascot   string `json:"mascot"`
		Color    string `json:"color"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	player, room, token, err := h.roomService.JoinRoom(r.Context(), code, body.Nickname, body.Mascot, body.Color)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"player":        player,
		"room":          room,
		"session_token": token,
	})
}

// HandleWebSocket /ws?token=...&room=...&nickname=...&mascot=...&color=...
func (h *RoomHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	roomCode := r.URL.Query().Get("room")

	if roomCode == "" {
		http.Error(w, "Paramètre room manquant", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var userID uuid.UUID
	var sessionData *domain.SessionData

	// 1. Tenter de résoudre la session existante via token (reconnexion transparente)
	if tokenStr != "" {
		sessionData, _ = h.roomRepo.GetSession(ctx, tokenStr)
	}

	if sessionData != nil && sessionData.RoomCode == roomCode {
		userID = sessionData.UserID
	} else {
		// 2. Si pas de token ou expiré, créer un nouvel ID pour ce client
		nickname := r.URL.Query().Get("nickname")
		if nickname == "" {
			nickname = "Joueur"
		}
		mascot := r.URL.Query().Get("mascot")
		if mascot == "" {
			mascot = "dice"
		}
		color := r.URL.Query().Get("color")
		if color == "" {
			color = "#FFD300"
		}

		userID = uuid.New()
		tokenStr = h.roomService.GenerateSessionToken()
		sessionData = &domain.SessionData{
			UserID:   userID,
			RoomCode: roomCode,
			Nickname: nickname,
			Mascot:   mascot,
			Color:    color,
		}
		_ = h.roomRepo.CreateSession(ctx, tokenStr, sessionData)

		// Ajouter à la salle si absent
		if _, err := h.roomRepo.GetPlayer(ctx, roomCode, userID); err != nil {
			_ = h.roomRepo.AddPlayer(ctx, roomCode, &domain.RoomPlayer{
				ID:          userID,
				Nickname:    nickname,
				Mascot:      mascot,
				Color:       color,
				Role:        domain.RolePlayer,
				IsMaster:    false,
				IsSpectator: false,
				IsMuted:     false,
				IsConnected: true,
				JoinedAt:    time.Now(),
				LastSeenAt:  time.Now(),
			})
		}
	}

	// Accepter la connexion WebSocket
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}

	client := ws.NewClient(h.hub, conn, roomCode, userID, tokenStr)
	h.hub.Register(client)

	// Lancement des pompes de communication
	go client.WritePump(r.Context())
	client.ReadPump(r.Context())
}
