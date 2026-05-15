package room

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomNotLobby = errors.New("room is not in lobby status")
)

type Service interface {
	CreateRoom(ctx context.Context, hostID uuid.UUID) (*Room, error)
	JoinRoom(ctx context.Context, userID uuid.UUID, code string) (*Room, error)
	GetByCode(ctx context.Context, code string) (*Room, error)
}

type service struct {
	repo     Repository
	userRepo user.Repository
	redis    *goredis.Client
}

func NewService(repo Repository, userRepo user.Repository, redis *goredis.Client) Service {
	return &service{repo: repo, userRepo: userRepo, redis: redis}
}

func generateCode() string {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "ABCDEF" // Fallback
	}
	return strings.ToUpper(hex.EncodeToString(bytes))
}

func (s *service) CreateRoom(ctx context.Context, hostID uuid.UUID) (*Room, error) {
	r := &Room{
		Code:       generateCode(),
		Status:     StatusLobby,
		HostID:     hostID,
		MaxPlayers: 8, // default
	}

	if err := s.repo.Create(ctx, r); err != nil {
		return nil, err
	}

	// Host joins the room in Redis (Hash: userID -> username)
	u, err := s.userRepo.GetByID(ctx, hostID)
	if err == nil && u != nil {
		name := u.Username
		if u.DisplayName != nil && *u.DisplayName != "" {
			name = *u.DisplayName
		}
		redisKey := fmt.Sprintf("room:%s:players", r.Code)
		s.redis.HSet(ctx, redisKey, u.ID.String(), name)
	}

	return r, nil
}

func (s *service) JoinRoom(ctx context.Context, userID uuid.UUID, code string) (*Room, error) {
	r, err := s.repo.GetByCode(ctx, strings.ToUpper(code))
	if err != nil {
		return nil, err
	}

	if r.Status != StatusLobby {
		return nil, ErrRoomNotLobby
	}

	// Fetch user to get username
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	name := u.Username
	if u.DisplayName != nil && *u.DisplayName != "" {
		name = *u.DisplayName
	}

	// Add to Redis (Hash: userID -> username)
	redisKey := fmt.Sprintf("room:%s:players", r.Code)
	s.redis.HSet(ctx, redisKey, u.ID.String(), name)

	return r, nil
}

func (s *service) GetByCode(ctx context.Context, code string) (*Room, error) {
	return s.repo.GetByCode(ctx, strings.ToUpper(code))
}
