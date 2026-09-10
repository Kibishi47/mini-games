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

// RoomPlayer représente un joueur dans une salle
type RoomPlayer struct {
	ID          uuid.UUID  `json:"id"`
	Nickname    string     `json:"nickname"`
	Mascot      string     `json:"mascot"` // "dice", "domino", "card", "knight", "d20", "meeple"
	Color       string     `json:"color"`  // Hex code (ex: #FFD300, #1D4ED8, etc.)
	Role        PlayerRole `json:"role"`
	IsMaster    bool       `json:"is_master"`
	IsSpectator bool       `json:"is_spectator"`
	IsMuted     bool       `json:"is_muted"`
	IsConnected bool       `json:"is_connected"`
	JoinedAt    time.Time  `json:"joined_at"`
	LastSeenAt  time.Time  `json:"last_seen_at"`
	Score       int        `json:"score"`
}

// RoomSettings configure la salle de jeu
type RoomSettings struct {
	GameType      string `json:"game_type"`      // "wordle"
	WordLength    int    `json:"word_length"`    // 3 à 8 (défaut 5)
	RoundDuration int    `json:"round_duration"` // secondes (défaut 60)
	MaxRounds     int    `json:"max_rounds"`     // manches (défaut 3)
	MaxAttempts   int    `json:"max_attempts"`   // essais max (défaut 6)
	Language      string `json:"language"`       // "fr"
}

// Room représente l'état global d'une salle
type Room struct {
	Code         string       `json:"code"`
	Status       RoomStatus   `json:"status"`
	MasterID     uuid.UUID    `json:"master_id"`
	Settings     RoomSettings `json:"settings"`
	CurrentRound int          `json:"current_round"`
	SecretWord   string       `json:"secret_word,omitempty"` // Masqué pour les clients en cours de jeu
	EndsAt       *time.Time   `json:"ends_at,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	Players      []RoomPlayer `json:"players"`
}

// SessionData stockée dans Redis sous session:{token} avec TTL 45s pour reconnexion transparente
type SessionData struct {
	UserID   uuid.UUID `json:"user_id"`
	RoomCode string    `json:"room_code"`
	Nickname string    `json:"nickname"`
	Mascot   string    `json:"mascot"`
	Color    string    `json:"color"`
}

// ChatMessage représente un message de discussion ou système
type ChatMessage struct {
	ID        string    `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Sender    string    `json:"sender"`
	Mascot    string    `json:"mascot"`
	Color     string    `json:"color"`
	Content   string    `json:"content"`
	IsSystem  bool      `json:"is_system"`
	CreatedAt time.Time `json:"created_at"`
}

// RoundHistory récapitule une manche passée
type RoundHistory struct {
	RoundNum   int                    `json:"round_num"`
	SecretWord string                 `json:"secret_word"`
	WinnerID   *uuid.UUID             `json:"winner_id,omitempty"`
	WinnerName string                 `json:"winner_name,omitempty"`
	Scores     map[string]int         `json:"scores"` // UserID string -> points de la manche
	EndedAt    time.Time              `json:"ended_at"`
}

// WSMessage format standard de message WebSocket
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
