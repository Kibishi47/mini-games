package auth

import (
	"context"
	"time"

	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	tokenRepo    TokenRepository
	identityRepo IdentityRepository
	userRepo     user.Repository
	tokenTTL     time.Duration
}

func NewService(
	tokenRepo TokenRepository,
	identityRepo IdentityRepository,
	userRepo user.Repository,
	tokenTTL time.Duration,
) *Service {
	return &Service{
		tokenRepo:    tokenRepo,
		identityRepo: identityRepo,
		userRepo:     userRepo,
		tokenTTL:     tokenTTL,
	}
}

func (s *Service) LoginLocal(ctx context.Context, username string, password string) (rawToken string, expiresAt time.Time, err error) {
	userID, err := s.identityRepo.FindUserIDByProvider(ctx, ProviderLocal, username)
	if err != nil {
		return "", time.Time{}, ErrInvalidCredentials
	}

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", time.Time{}, ErrInvalidCredentials
	}

	if u.PasswordHash == nil {
		return "", time.Time{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)); err != nil {
		return "", time.Time{}, ErrInvalidCredentials
	}

	return s.generateToken(ctx, userID)
}

func (s *Service) RegisterLocal(ctx context.Context, username string, password string) (rawToken string, expiresAt time.Time, err error) {
	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return "", time.Time{}, err
	}
	if exists {
		return "", time.Time{}, ErrUsernameTaken
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", time.Time{}, err
	}
	hash := string(hashBytes)

	u := &user.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: &hash,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return "", time.Time{}, err
	}

	if err := s.identityRepo.CreateIdentity(ctx, u.ID, ProviderLocal, username); err != nil {
		return "", time.Time{}, err
	}

	return s.generateToken(ctx, u.ID)
}

func (s *Service) Authenticate(ctx context.Context, rawToken string) (uuid.UUID, error) {
	if rawToken == "" {
		return uuid.Nil, ErrInvalidToken
	}

	tokenHash := HashToken(rawToken)

	userID, expiresAt, err := s.tokenRepo.FindUserIDByTokenHash(ctx, tokenHash)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	if time.Now().After(expiresAt) {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	tokenHash := HashToken(rawToken)
	return s.tokenRepo.DeleteTokenHash(ctx, tokenHash)
}

func (s *Service) generateToken(ctx context.Context, userID uuid.UUID) (string, time.Time, error) {
	rawToken, tokenHash, err := GenerateToken()
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(s.tokenTTL)

	if err := s.tokenRepo.CreateToken(ctx, userID, tokenHash, expiresAt); err != nil {
		return "", time.Time{}, err
	}

	return rawToken, expiresAt, nil
}
