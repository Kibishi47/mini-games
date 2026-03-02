package ws

import (
	"net/http"

	"github.com/Kibishi47/mini-games/back/internal/http/auth"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub  *Hub
	auth auth.Service
}

func NewHandler(hub *Hub, auth auth.Service) *Handler {
	return &Handler{
		hub:  hub,
		auth: auth,
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

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := NewClient(conn, userID)
	h.hub.register <- client

	go client.writePump()

	for {
		_, _, err := client.conn.ReadMessage()
		if err != nil {
			break
		}
	}

	h.hub.unregister <- client
}
