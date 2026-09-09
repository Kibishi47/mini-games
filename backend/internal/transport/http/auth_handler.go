package http

import (
	"encoding/json"
	"net/http"

	"minigames-backend/internal/repository/postgres"
	"minigames-backend/internal/service/auth"
)

type AuthHandler struct {
	authService *auth.AuthService
	userRepo    *postgres.UserRepository
}

func NewAuthHandler(authService *auth.AuthService, userRepo *postgres.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
	}
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	jsonResponse(w, status, map[string]string{"error": msg})
}

// GuestLoginPOST crée une session invité 1-clic
func (h *AuthHandler) GuestLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomName string `json:"custom_name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	res, err := h.authService.CreateGuestSession(r.Context(), body.CustomName)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, res)
}

// RegisterLocalPOST enregistre un compte normal
func (h *AuthHandler) RegisterLocal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Display  string `json:"display_username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "données invalides")
		return
	}

	res, err := h.authService.RegisterLocal(r.Context(), body.Username, body.Display, body.Email, body.Password)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, res)
}

// LoginLocalPOST connecte un compte
func (h *AuthHandler) LoginLocal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "données invalides")
		return
	}

	res, err := h.authService.LoginLocal(r.Context(), body.Username, body.Password)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, res)
}

// UpgradeGuestPOST convertit un invité connecté en compte persistant
func (h *AuthHandler) UpgradeGuest(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsKey).(*auth.Claims)
	if !ok || !claims.IsGuest {
		jsonError(w, http.StatusBadRequest, "cette action est réservée aux comptes invités")
		return
	}

	var body struct {
		Username string `json:"username"`
		Display  string `json:"display_username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "données invalides")
		return
	}

	res, err := h.authService.ConvertGuestToLocal(r.Context(), claims.UserID, body.Username, body.Display, body.Email, body.Password)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, res)
}

// DiscordAuthURLGET renvoie l'URL Discord OAuth2
func (h *AuthHandler) DiscordAuthURL(w http.ResponseWriter, r *http.Request) {
	url := h.authService.GetDiscordAuthURL()
	if url == "" {
		jsonError(w, http.StatusNotImplemented, "Discord OAuth non activé")
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"url": url})
}

// DiscordCallbackGET traite le retour Discord
func (h *AuthHandler) DiscordCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		jsonError(w, http.StatusBadRequest, "code OAuth manquant")
		return
	}

	res, err := h.authService.HandleDiscordCallback(r.Context(), code)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, res)
}

// GetMeGET renvoie le profil et les stats de l'utilisateur connecté
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsKey).(*auth.Claims)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "non autorisé")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), claims.UserID)
	if err != nil {
		jsonError(w, http.StatusNotFound, "utilisateur introuvable")
		return
	}

	stats, _ := h.userRepo.GetStats(r.Context(), claims.UserID, "wordle")

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"user":  user,
		"stats": stats,
	})
}

// UpdateProfilePUT met à jour les informations de profil
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsKey).(*auth.Claims)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "non autorisé")
		return
	}

	var body struct {
		DisplayUsername string `json:"display_username"`
		AvatarURL       string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "données invalides")
		return
	}

	if err := h.userRepo.UpdateProfile(r.Context(), claims.UserID, body.DisplayUsername, body.AvatarURL); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, _ := h.userRepo.GetByID(r.Context(), claims.UserID)
	jsonResponse(w, http.StatusOK, user)
}
