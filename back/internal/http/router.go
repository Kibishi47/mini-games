package http

import (
	"net/http"
	"time"

	"github.com/Kibishi47/mini-games/back/internal/http/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Kibishi47/mini-games/back/internal/config"
)

func NewRouter(cfg config.Config, db *pgxpool.Pool) http.Handler {
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
	healthHandler := handlers.NewHealthHandler(db)

	// Routes
	r.Get("/health", healthHandler.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler.Health)
		r.Get("/health/db", healthHandler.HealthDB)
	})

	return r
}
