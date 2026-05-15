package session

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CreateSession(ctx context.Context, roomID uuid.UUID, game string, config map[string]any) (*GameSession, error)
	GetActiveSessionByRoomID(ctx context.Context, roomID uuid.UUID) (*GameSession, error)
	EndSession(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateSession(ctx context.Context, roomID uuid.UUID, game string, config map[string]any) (*GameSession, error) {
	now := time.Now()
	session := &GameSession{
		RoomID:    roomID,
		Game:      game,
		Config:    config,
		StartedAt: &now,
	}

	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *service) GetActiveSessionByRoomID(ctx context.Context, roomID uuid.UUID) (*GameSession, error) {
	return s.repo.GetActiveSessionByRoomID(ctx, roomID)
}

func (s *service) EndSession(ctx context.Context, id uuid.UUID) error {
	return s.repo.EndSession(ctx, id)
}
