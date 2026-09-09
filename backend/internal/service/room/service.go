package room

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"minigames-backend/internal/domain"
	"minigames-backend/internal/repository/postgres"
	"minigames-backend/internal/repository/redis"
)

type RoomService struct {
	roomRepo    *redis.RoomRepository
	sessionRepo *postgres.RoomSessionRepository
	userRepo    *postgres.UserRepository
}

func NewRoomService(roomRepo *redis.RoomRepository, sessionRepo *postgres.RoomSessionRepository, userRepo *postgres.UserRepository) *RoomService {
	return &RoomService{
		roomRepo:    roomRepo,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
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

// CreateRoom instancie une nouvelle salle avec le créateur en Master
func (s *RoomService) CreateRoom(ctx context.Context, masterID uuid.UUID, settings *domain.RoomSettings) (*domain.Room, error) {
	master, err := s.userRepo.GetByID(ctx, masterID)
	if err != nil {
		return nil, errors.New("utilisateur introuvable")
	}

	code := s.GenerateRoomCode()

	defaultSettings := domain.RoomSettings{
		GameType:      "wordle",
		WordLength:    5,
		RoundDuration: 60,
		MaxRounds:     3,
		MaxAttempts:   6,
		Language:      "fr",
	}
	if settings != nil {
		if settings.GameType != "" {
			defaultSettings.GameType = settings.GameType
		}
		if settings.WordLength >= 4 && settings.WordLength <= 8 {
			defaultSettings.WordLength = settings.WordLength
		}
		if settings.RoundDuration >= 30 && settings.RoundDuration <= 300 {
			defaultSettings.RoundDuration = settings.RoundDuration
		}
		if settings.MaxRounds >= 1 && settings.MaxRounds <= 10 {
			defaultSettings.MaxRounds = settings.MaxRounds
		}
		if settings.Language != "" {
			defaultSettings.Language = settings.Language
		}
	}

	now := time.Now()
	masterPlayer := domain.RoomPlayer{
		UserID:          master.ID,
		Username:        master.Username,
		DisplayUsername: master.DisplayUsername,
		AvatarURL:       master.AvatarURL,
		Role:            domain.RoleMaster,
		IsMuted:         false,
		IsConnected:     false, // deviendra true à la connexion WS
		JoinedAt:        now,
		LastSeenAt:      now,
		Score:           0,
	}

	room := &domain.Room{
		Code:         code,
		Status:       domain.RoomStatusInLobby,
		MasterID:     master.ID,
		Settings:     defaultSettings,
		CurrentRound: 1,
		CreatedAt:    now,
		Players:      []domain.RoomPlayer{masterPlayer},
	}

	// Sauvegarde Redis
	if err := s.roomRepo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}
	if err := s.roomRepo.AddPlayer(ctx, code, &masterPlayer); err != nil {
		return nil, err
	}

	// Persistance PostgreSQL
	_, _ = s.sessionRepo.CreateSession(ctx, code, master.ID)

	return room, nil
}

// JoinRoom prépare l'accès d'un utilisateur à une room
func (s *RoomService) JoinRoom(ctx context.Context, code string, userID uuid.UUID) (*domain.RoomPlayer, *domain.Room, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	// Vérifier si banni
	if s.roomRepo.IsBanned(ctx, code, userID) {
		return nil, nil, errors.New("vous êtes banni de cette salle")
	}

	room, err := s.roomRepo.GetRoom(ctx, code)
	if err != nil {
		return nil, nil, errors.New("salle introuvable")
	}

	if room.Status == domain.RoomStatusClosed {
		return nil, nil, errors.New("cette salle est fermée")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, errors.New("utilisateur introuvable")
	}

	// Vérifier si le joueur est déjà présent (reconnexion)
	existingPlayer, _ := s.roomRepo.GetPlayer(ctx, code, userID)
	if existingPlayer != nil {
		// Le joueur existant garde son rôle et son score
		existingPlayer.IsConnected = true
		existingPlayer.LastSeenAt = time.Now()
		_ = s.roomRepo.AddPlayer(ctx, code, existingPlayer)
		return existingPlayer, room, nil
	}

	// Nouveau joueur
	role := domain.RolePlayer
	// Si la partie est déjà en cours, entre en SPECTATEUR automatique
	if room.Status == domain.RoomStatusInGame {
		role = domain.RoleSpectator
	}

	now := time.Now()
	newPlayer := &domain.RoomPlayer{
		UserID:          user.ID,
		Username:        user.Username,
		DisplayUsername: user.DisplayUsername,
		AvatarURL:       user.AvatarURL,
		Role:            role,
		IsMuted:         false,
		IsConnected:     true,
		JoinedAt:        now,
		LastSeenAt:      now,
		Score:           0,
	}

	if err := s.roomRepo.AddPlayer(ctx, code, newPlayer); err != nil {
		return nil, nil, err
	}

	return newPlayer, room, nil
}

func (s *RoomService) GetRoom(ctx context.Context, code string) (*domain.Room, error) {
	return s.roomRepo.GetRoom(ctx, strings.ToUpper(strings.TrimSpace(code)))
}
