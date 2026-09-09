package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"minigames-backend/internal/config"
	"minigames-backend/internal/domain"
	"minigames-backend/internal/repository/postgres"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	IsGuest  bool      `json:"is_guest"`
	jwt.RegisteredClaims
}

type AuthService struct {
	cfg      *config.Config
	userRepo *postgres.UserRepository
}

func NewAuthService(cfg *config.Config, userRepo *postgres.UserRepository) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
	}
}

// GenerateToken crée un JWT signé
func (s *AuthService) GenerateToken(user *domain.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsGuest:  user.IsGuest,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.JWTExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "minigames-auth",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// ValidateToken vérifie et extrait les claims d'un JWT
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("token invalide")
}

// RegisterLocal crée un compte avec mot de passe haché Argon2id
func (s *AuthService) RegisterLocal(ctx context.Context, username, display, email, password string) (*domain.AuthResponse, error) {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 30 {
		return nil, errors.New("le nom d'utilisateur doit comporter entre 3 et 30 caractères")
	}
	if len(password) < 6 {
		return nil, errors.New("le mot de passe doit comporter au moins 6 caractères")
	}

	if display == "" {
		display = username
	}

	// Vérifier l'unicité
	existing, _ := s.userRepo.GetByUsername(ctx, username)
	if existing != nil {
		return nil, errors.New("ce nom d'utilisateur est déjà pris")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du hash: %w", err)
	}

	var mailPtr *string
	if email != "" {
		mailPtr = &email
	}

	avatar := fmt.Sprintf("https://api.dicebear.com/7.x/bottts/svg?seed=%s", username)

	user := &domain.User{
		ID:              uuid.New(),
		Username:        username,
		DisplayUsername: display,
		Email:           mailPtr,
		PasswordHash:    &hash,
		AvatarURL:       avatar,
		IsGuest:         false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}

// LoginLocal authentifie un utilisateur local
func (s *AuthService) LoginLocal(ctx context.Context, username, password string) (*domain.AuthResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("identifiants invalides")
	}

	if user.PasswordHash == nil {
		return nil, errors.New("ce compte ne supporte pas la connexion par mot de passe")
	}

	valid, err := VerifyPassword(password, *user.PasswordHash)
	if err != nil || !valid {
		return nil, errors.New("identifiants invalides")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}

// CreateGuestSession génère instantanément un compte invité
func (s *AuthService) CreateGuestSession(ctx context.Context, customName string) (*domain.AuthResponse, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	randomNum := 1000 + n.Int64()

	username := fmt.Sprintf("guest_%d", randomNum)
	displayName := fmt.Sprintf("Guest_%d", randomNum)
	if customName != "" {
		displayName = strings.TrimSpace(customName)
	}

	avatar := fmt.Sprintf("https://api.dicebear.com/7.x/bottts/svg?seed=%s", username)

	user := &domain.User{
		ID:              uuid.New(),
		Username:        username,
		DisplayUsername: displayName,
		AvatarURL:       avatar,
		IsGuest:         true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}

// ConvertGuestToLocal convertit un invité en compte permanent
func (s *AuthService) ConvertGuestToLocal(ctx context.Context, guestID uuid.UUID, username, display, email, password string) (*domain.AuthResponse, error) {
	if len(username) < 3 || len(password) < 6 {
		return nil, errors.New("informations invalides")
	}

	// Vérifier unicité du nouveau pseudo
	existing, _ := s.userRepo.GetByUsername(ctx, username)
	if existing != nil && existing.ID != guestID {
		return nil, errors.New("ce nom d'utilisateur est déjà utilisé")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	if display == "" {
		display = username
	}

	if err := s.userRepo.UpgradeGuest(ctx, guestID, username, display, email, hash); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, guestID)
	if err != nil {
		return nil, err
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}

// Discord OAuth2
type discordUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
}

func (s *AuthService) GetDiscordAuthURL() string {
	if s.cfg.DiscordClientID == "" {
		return ""
	}
	return fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify%%20email",
		s.cfg.DiscordClientID,
		url.QueryEscape(s.cfg.DiscordRedirectURI),
	)
}

func (s *AuthService) HandleDiscordCallback(ctx context.Context, code string) (*domain.AuthResponse, error) {
	if s.cfg.DiscordClientID == "" || s.cfg.DiscordClientSecret == "" {
		return nil, errors.New("discord OAuth non configuré")
	}

	// Échange du code contre le token
	data := url.Values{}
	data.Set("client_id", s.cfg.DiscordClientID)
	data.Set("client_secret", s.cfg.DiscordClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.cfg.DiscordRedirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenRes struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return nil, err
	}

	// Récupération de l'utilisateur Discord
	userReq, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)

	userResp, err := client.Do(userReq)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()

	var du discordUserResponse
	if err := json.NewDecoder(userResp.Body).Decode(&du); err != nil {
		return nil, err
	}

	// Recherche ou création de l'utilisateur
	user, _ := s.userRepo.GetByDiscordID(ctx, du.ID)
	if user == nil {
		avatarURL := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", du.ID, du.Avatar)
		if du.Avatar == "" {
			avatarURL = fmt.Sprintf("https://api.dicebear.com/7.x/bottts/svg?seed=%s", du.Username)
		}

		user = &domain.User{
			ID:              uuid.New(),
			Username:        fmt.Sprintf("discord_%s", du.Username),
			DisplayUsername: du.Username,
			Email:           &du.Email,
			DiscordID:       &du.ID,
			AvatarURL:       avatarURL,
			IsGuest:         false,
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}
