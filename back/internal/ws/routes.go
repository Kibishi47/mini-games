package ws

import (
	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	"github.com/Kibishi47/mini-games/back/internal/http/auth"
	"github.com/go-chi/chi/v5"
)

func Routes(r chi.Router, hub *Hub, auth auth.Service, userRepo user.Repository) {
	wsHandler := NewHandler(hub, auth, userRepo)

	r.Group(func(r chi.Router) {
		r.Get("/ws", wsHandler.Handle)
	})
}
