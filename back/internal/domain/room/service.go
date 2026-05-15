package room

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomNotLobby = errors.New("room is not in lobby status")
)

type Service interface {
	CreateRoom(ctx context.Context, hostID uuid.UUID) (*Room, error)
	JoinRoom(ctx context.Context, userID uuid.UUID, code string) (*Room, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
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

	return r, nil
}
