package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash *string
	DisplayName  *string
	AvatarURL    *string
	CreatedAt    time.Time
}
