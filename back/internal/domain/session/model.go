package session

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type GameSession struct {
	ID        uuid.UUID      `json:"id"`
	RoomID    uuid.UUID      `json:"roomId"`
	Game      string         `json:"game"`
	Config    map[string]any `json:"config"`
	StartedAt *time.Time     `json:"startedAt"`
	EndedAt   *time.Time     `json:"endedAt"`
}

type Repository interface {
	Create(ctx context.Context, session *GameSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*GameSession, error)
	GetActiveSessionByRoomID(ctx context.Context, roomID uuid.UUID) (*GameSession, error)
	EndSession(ctx context.Context, id uuid.UUID) error
}
