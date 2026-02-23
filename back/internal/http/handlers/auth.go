package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Kibishi47/mini-games/back/internal/domain/auth"
	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	"github.com/Kibishi47/mini-games/back/internal/http/middlewares"
	"github.com/Kibishi47/mini-games/back/internal/http/response"
	"github.com/google/uuid"
)

type AuthUseCase interface {
	LoginLocal(ctx context.Context, username string, password string) (string, time.Time, error)
	RegisterLocal(ctx context.Context, username string, password string) (string, time.Time, error)
	Authenticate(ctx context.Context, rawToken string) (uuid.UUID, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	Logout(ctx context.Context, rawToken string) error
}

type AuthHandler struct {
	auth AuthUseCase
}

func NewAuthHandler(auth AuthUseCase) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		response.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	token, expiresAt, err := h.auth.LoginLocal(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			response.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		response.Error(w, "an error occur", http.StatusInternalServerError)
		return
	}

	response.JSON(w, http.StatusOK, authResponse{
		AccessToken: token,
		ExpiresAt:   expiresAt,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		response.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	token, expiresAt, err := h.auth.RegisterLocal(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUsernameTaken) {
			response.Error(w, "username taken", http.StatusConflict)
			return
		}

		response.Error(w, "an error occur", http.StatusInternalServerError)
		return
	}

	response.JSON(w, http.StatusCreated, authResponse{
		AccessToken: token,
		ExpiresAt:   expiresAt,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := h.auth.GetUserByID(r.Context(), userID)
	if err != nil {
		response.Error(w, "user not found", http.StatusNotFound)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":          u.ID,
			"username":    u.Username,
			"displayName": u.DisplayName,
			"avatarUrl":   u.AvatarURL,
			"createdAt":   u.CreatedAt,
		},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	rawToken, ok := middlewares.TokenFromContext(r.Context())
	if !ok {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.auth.Logout(r.Context(), rawToken)
	if err != nil {
		response.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	response.Success(w, "ok", http.StatusOK)
}
