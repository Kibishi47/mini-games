package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RoomStatus string

const (
	RoomStatusInLobby RoomStatus = "in_lobby"
	RoomStatusInGame  RoomStatus = "in_game"
	RoomStatusClosed  RoomStatus = "closed"
)

type PlayerRole string

const (
	RoleMaster    PlayerRole = "master"
	RolePlayer    PlayerRole = "player"
	RoleSpectator PlayerRole = "spectator"
)

type RoomPlayer struct {
	UserID          uuid.UUID  `json:"user_id"`
	Username        string     `json:"username"`
	DisplayUsername string     `json:"display_username"`
	AvatarURL       string     `json:"avatar_url"`
	Role            PlayerRole `json:"role"`
	IsMuted         bool       `json:"is_muted"`
	IsConnected     bool       `json:"is_connected"`
	JoinedAt        time.Time  `json:"joined_at"`
	LastSeenAt      time.Time  `json:"last_seen_at"`
	Score           int        `json:"score"`
}

type RoomSettings struct {
	GameType      string `json:"game_type"`       // "wordle"
	WordLength    int    `json:"word_length"`     // 5, 6, 7
	RoundDuration int    `json:"round_duration"`  // en secondes (ex: 60, 90, 120)
	MaxRounds     int    `json:"max_rounds"`      // ex: 3, 5
	MaxAttempts   int    `json:"max_attempts"`    // ex: 6
	Language      string `json:"language"`        // "fr", "en"
}

type Room struct {
	Code         string       `json:"code"`
	Status       RoomStatus   `json:"status"`
	MasterID     uuid.UUID    `json:"master_id"`
	Settings     RoomSettings `json:"settings"`
	CurrentRound int          `json:"current_round"`
	EndsAt       *time.Time   `json:"ends_at,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	Players      []RoomPlayer `json:"players"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Sender    string    `json:"sender"`
	AvatarURL string    `json:"avatar_url"`
	Content   string    `json:"content"`
	IsSystem  bool      `json:"is_system"`
	CreatedAt time.Time `json:"created_at"`
}

// WebSocket Message Standardisé
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
