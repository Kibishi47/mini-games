package http

import (
	"github.com/Kibishi47/mini-games/back/internal/http/auth"
)

type Deps struct {
	Auth auth.Service
}
