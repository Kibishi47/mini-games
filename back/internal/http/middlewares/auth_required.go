package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"

	domainauth "github.com/Kibishi47/mini-games/back/internal/domain/auth"
	"github.com/Kibishi47/mini-games/back/internal/http/response"
	"github.com/google/uuid"
)

type Authenticator interface {
	Authenticate(ctx context.Context, rawToken string) (uuid.UUID, error)
}

type ctxKey int

const (
	userIDKey ctxKey = iota
	rawTokenKey
)

func AuthRequired(auth Authenticator) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				response.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
			if rawToken == "" {
				response.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			userID, err := auth.Authenticate(r.Context(), rawToken)
			if err != nil {
				if errors.Is(err, domainauth.ErrInvalidToken) {
					response.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}

				response.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, rawTokenKey, rawToken)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TokenFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(rawTokenKey)
	token, ok := v.(string)
	return token, ok
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(userIDKey)
	id, ok := v.(uuid.UUID)
	return id, ok
}
