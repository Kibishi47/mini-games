package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"minigames-backend/internal/domain"
	"minigames-backend/internal/service/auth"
	"minigames-backend/internal/service/room"
	"minigames-backend/internal/transport/ws"
	"nhooyr.io/websocket"
)

type RoomHandler struct {
	roomService *room.RoomService
	hub         *ws.Hub
	authService *auth.AuthService
}

func NewRoomHandler(roomService *room.RoomService, hub *ws.Hub, authService *auth.AuthService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		hub:         hub,
		authService: authService,
	}
}

// CreateRoomPOST /api/rooms
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsKey).(*auth.Claims)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "non autorisé")
		return
	}

	var body struct {
		Settings *domain.RoomSettings `json:"settings"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	room, err := h.roomService.CreateRoom(r.Context(), claims.UserID, body.Settings)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, room)
}

// GetRoomGET /api/rooms/{code}
func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	room, err := h.roomService.GetRoom(r.Context(), code)
	if err != nil {
		jsonError(w, http.StatusNotFound, "salle introuvable")
		return
	}

	jsonResponse(w, http.StatusOK, room)
}

// JoinRoomPOST /api/rooms/{code}/join
func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsKey).(*auth.Claims)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "non autorisé")
		return
	}

	code := chi.URLParam(r, "code")
	player, room, err := h.roomService.JoinRoom(r.Context(), code, claims.UserID)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"player": player,
		"room":   room,
	})
}

// HandleWebSocket /ws?token=...&room=...
func (h *RoomHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	roomCode := r.URL.Query().Get("room")

	if tokenStr == "" || roomCode == "" {
		http.Error(w, "Paramètres token ou room manquants", http.StatusBadRequest)
		return
	}

	claims, err := h.authService.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "Token invalide ou expiré", http.StatusUnauthorized)
		return
	}

	// Accepter la connexion WebSocket
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Autoriser CORS pour les clients Web
	})
	if err != nil {
		return
	}

	client := ws.NewClient(h.hub, conn, roomCode, claims.UserID)
	h.hub.Register(client)

	// Lancement des pompes de communication
	go client.WritePump(r.Context())
	client.ReadPump(r.Context())
}
