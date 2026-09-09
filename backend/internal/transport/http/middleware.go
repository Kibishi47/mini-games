package http

import (
	"context"
	"net/http"
	"strings"

	"minigames-backend/internal/service/auth"
)

type contextKey string

const UserClaimsKey contextKey = "user_claims"

func AuthMiddleware(authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			tokenStr := ""

			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				// Vérifier aussi le query param token pour WebSocket
				tokenStr = r.URL.Query().Get("token")
			}

			if tokenStr == "" {
				http.Error(w, `{"error":"non autorisé, token manquant"}`, http.StatusUnauthorized)
				return
			}

			claims, err := authService.ValidateToken(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"token expiré ou invalide"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RecoverMiddleware protège les routes HTTP contre les crashs inattendus
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"erreur interne du serveur"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
