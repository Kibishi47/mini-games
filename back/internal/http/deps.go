package http

import (
	"github.com/Kibishi47/mini-games/back/internal/http/auth"
	"github.com/Kibishi47/mini-games/back/internal/ws"
)

type Deps struct {
	Auth auth.Service
	WS   *ws.Hub
}
