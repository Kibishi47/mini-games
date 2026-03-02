package ws

import (
	"github.com/Kibishi47/mini-games/back/internal/http/auth"
	"github.com/go-chi/chi/v5"
)

func Routes(r chi.Router, hub *Hub, auth auth.Service) {
	wsHandler := NewHandler(hub, auth)

	r.Group(func(r chi.Router) {
		r.Get("/ws", wsHandler.Handle)
	})
}
