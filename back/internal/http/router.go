package http

import (
	"net/http"
	"time"

	"github.com/Kibishi47/mini-games/back/internal/http/handlers"
	"github.com/Kibishi47/mini-games/back/internal/http/middlewares"
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

	// Handlers
	healthHandler := handlers.NewHealthHandler(deps.DB)
	authHandler := handlers.NewAuthHandler(deps.Auth)

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler.Health)
		r.Get("/health/db", healthHandler.HealthDB)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/register", authHandler.Register)

			r.Group(func(r chi.Router) {
				r.Use(middlewares.AuthRequired(deps.Auth))

				r.Get("/me", authHandler.Me)
				r.Post("/logout", authHandler.Logout)
			})
		})
	})

	return r
}
