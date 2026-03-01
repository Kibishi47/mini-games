package health

import "github.com/go-chi/chi/v5"

func Routes(r chi.Router) {
	healthHandler := NewHandler()

	r.Get("/health", healthHandler.Health)
}
