package http

import (
	"github.com/Kibishi47/mini-games/back/internal/domain/room"
	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	"github.com/Kibishi47/mini-games/back/internal/http/auth"
	"github.com/Kibishi47/mini-games/back/internal/ws"
)

type Deps struct {
	Auth auth.Service
	Room room.Service
	User user.Repository
	WS   *ws.Hub
}
