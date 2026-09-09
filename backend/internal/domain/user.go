package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID  `json:"id"`
	Username        string     `json:"username"`
	DisplayUsername string     `json:"display_username"`
	Email           *string    `json:"email,omitempty"`
	PasswordHash    *string    `json:"-"`
	DiscordID       *string    `json:"discord_id,omitempty"`
	AvatarURL       string     `json:"avatar_url"`
	IsGuest         bool       `json:"is_guest"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type UserStats struct {
	UserID       uuid.UUID `json:"user_id"`
	GameType     string    `json:"game_type"`
	GamesPlayed  int       `json:"games_played"`
	GamesWon     int       `json:"games_won"`
	WinStreak    int       `json:"win_streak"`
	HighestScore int       `json:"highest_score"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
