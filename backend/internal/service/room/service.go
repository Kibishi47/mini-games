package room

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"minigames-backend/internal/domain"
	"minigames-backend/internal/repository/redis"
)

type RoomService struct {
	roomRepo *redis.RoomRepository
}

func NewRoomService(roomRepo *redis.RoomRepository) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
	}
}

// GenerateRoomCode crée un code court alphanumérique type ABCD-12
func (s *RoomService) GenerateRoomCode() string {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	const digits = "23456789"

	b := make([]byte, 4)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[n.Int64()]
	}

	d := make([]byte, 2)
	for i := range d {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		d[i] = digits[n.Int64()]
	}

	return fmt.Sprintf("%s-%s", string(b), string(d))
}

// GenerateSessionToken crée un token sécurisé aléatoire de 32 caractères pour reconnexion
func (s *RoomService) GenerateSessionToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateRoom instancie une nouvelle salle avec l'invité créateur en Master
func (s *RoomService) CreateRoom(ctx context.Context, nickname, mascot, color string, settings *domain.RoomSettings) (*domain.Room, string, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		nickname = "Joueur"
	}
	if mascot == "" {
		mascot = "dice"
	}
	if color == "" {
		color = "#FFD300"
	}

	code := s.GenerateRoomCode()
	masterID := uuid.New()

	st := domain.RoomSettings{
		GameType:      "wordle",
		WordLength:    5,
		RoundDuration: 60,
		MaxRounds:     3,
		MaxAttempts:   6,
		Language:      "fr",
	}
	if settings != nil {
		if settings.WordLength >= 3 && settings.WordLength <= 8 {
			st.WordLength = settings.WordLength
		}
		if settings.RoundDuration >= 30 && settings.RoundDuration <= 180 {
			st.RoundDuration = settings.RoundDuration
		}
		if settings.MaxRounds >= 1 && settings.MaxRounds <= 10 {
			st.MaxRounds = settings.MaxRounds
		}
		if settings.MaxAttempts >= 4 && settings.MaxAttempts <= 8 {
			st.MaxAttempts = settings.MaxAttempts
		}
	}

	room := &domain.Room{
		Code:         code,
		Status:       domain.RoomStatusInLobby,
		MasterID:     masterID,
		Settings:     st,
		CurrentRound: 0,
		CreatedAt:    time.Now(),
		Players:      make([]domain.RoomPlayer, 0),
	}

	if err := s.roomRepo.CreateRoom(ctx, room); err != nil {
		return nil, "", err
	}

	// Ajouter le créateur comme Master
	masterPlayer := &domain.RoomPlayer{
		ID:          masterID,
		Nickname:    nickname,
		Mascot:      mascot,
		Color:       color,
		Role:        domain.RoleMaster,
		IsMaster:    true,
		IsSpectator: false,
		IsMuted:     false,
		IsConnected: true,
		JoinedAt:    time.Now(),
		LastSeenAt:  time.Now(),
		Score:       0,
		Location:    "lobby",
	}

	if err := s.roomRepo.AddPlayer(ctx, code, masterPlayer); err != nil {
		return nil, "", err
	}

	// Créer le token de session éphémère (TTL 45s)
	token := s.GenerateSessionToken()
	_ = s.roomRepo.CreateSession(ctx, token, &domain.SessionData{
		UserID:   masterID,
		RoomCode: code,
		Nickname: nickname,
		Mascot:   mascot,
		Color:    color,
	})

	room.Players = []domain.RoomPlayer{*masterPlayer}
	return room, token, nil
}

// JoinRoom fait rejoindre un joueur invité
func (s *RoomService) JoinRoom(ctx context.Context, code, nickname, mascot, color string) (*domain.RoomPlayer, *domain.Room, string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	room, err := s.roomRepo.GetRoom(ctx, code)
	if err != nil {
		return nil, nil, "", errors.New("salle introuvable ou fermée")
	}

	if room.Status == domain.RoomStatusClosed {
		return nil, nil, "", errors.New("cette salle est fermée")
	}

	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		nickname = "Joueur"
	}
	if mascot == "" {
		mascot = "domino"
	}
	if color == "" {
		color = "#1D4ED8"
	}

	userID := uuid.New()

	// Si la partie est déjà en cours, le nouvel arrivant est spectateur automatique
	isSpectator := room.Status == domain.RoomStatusInGame
	role := domain.RolePlayer
	location := "lobby"
	if isSpectator {
		role = domain.RoleSpectator
		location = "in_game"
	}

	player := &domain.RoomPlayer{
		ID:          userID,
		Nickname:    nickname,
		Mascot:      mascot,
		Color:       color,
		Role:        role,
		IsMaster:    false,
		IsSpectator: isSpectator,
		IsMuted:     false,
		IsConnected: true,
		JoinedAt:    time.Now(),
		LastSeenAt:  time.Now(),
		Score:       0,
		Location:    location,
	}

	if err := s.roomRepo.AddPlayer(ctx, code, player); err != nil {
		return nil, nil, "", err
	}

	// Créer le session token
	token := s.GenerateSessionToken()
	_ = s.roomRepo.CreateSession(ctx, token, &domain.SessionData{
		UserID:   userID,
		RoomCode: code,
		Nickname: nickname,
		Mascot:   mascot,
		Color:    color,
	})

	updatedRoom, _ := s.roomRepo.GetRoom(ctx, code)
	return player, updatedRoom, token, nil
}

// GetRoom récupère les données d'une salle
func (s *RoomService) GetRoom(ctx context.Context, code string) (*domain.Room, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	return s.roomRepo.GetRoom(ctx, code)
}
