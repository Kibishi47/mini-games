package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(r chi.Router, uc UseCases, authMw func(http.Handler) http.Handler) {
	authHandler := NewHandler(uc)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)

		r.Group(func(r chi.Router) {
			r.Use(authMw)

			r.Get("/me", authHandler.Me)
			r.Post("/logout", authHandler.Logout)
		})
	})
}
