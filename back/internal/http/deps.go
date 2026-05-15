package http

import (
	"github.com/Kibishi47/mini-games/back/internal/domain/game"
	"github.com/Kibishi47/mini-games/back/internal/domain/room"
	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	"github.com/Kibishi47/mini-games/back/internal/http/auth"
	"github.com/Kibishi47/mini-games/back/internal/ws"
)

type Deps struct {
	Auth auth.Service
	Room room.Service
	User user.Repository
	Game game.Service
	WS   *ws.Hub
}
