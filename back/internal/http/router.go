package http

import (
	"net/http"
	"time"

	"github.com/Kibishi47/mini-games/back/internal/http/auth"
	"github.com/Kibishi47/mini-games/back/internal/http/health"
	"github.com/Kibishi47/mini-games/back/internal/http/middlewares"
	"github.com/Kibishi47/mini-games/back/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Kibishi47/mini-games/back/internal/config"
)

func NewRouter(cfg config.Config, deps *Deps) http.Handler {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))

	// Logger HTTP
	r.Use(middleware.Logger)

	// CORS
	if len(cfg.CorsAllowedOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   cfg.CorsAllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
			ExposedHeaders:   []string{"X-Request-ID"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	// Middlewares
	authMw := middlewares.AuthRequired(deps.Auth)

	// Routes
	r.Route("/api", func(r chi.Router) {
		health.Routes(r)
		auth.Routes(r, deps.Auth, authMw)
	})

	ws.Routes(r, deps.WS, deps.Auth)

	return r
}
