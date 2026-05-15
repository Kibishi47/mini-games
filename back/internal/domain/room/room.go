package room

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusLobby   Status = "lobby"
	StatusRunning Status = "running"
	StatusClosed  Status = "closed"
)

type Room struct {
	ID         uuid.UUID `json:"id"`
	Code       string    `json:"code"`
	Status     Status    `json:"status"`
	HostID     uuid.UUID `json:"hostId"`
	MaxPlayers int       `json:"maxPlayers"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Repository interface {
	Create(ctx context.Context, room *Room) error
	GetByCode(ctx context.Context, code string) (*Room, error)
}
