package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"minigames-backend/internal/domain"
)

var (
	ErrRoomNotFound = errors.New("salle introuvable ou fermée")
)

type RoomRepository struct {
	client *redis.Client
}

func NewRoomRepository(client *redis.Client) *RoomRepository {
	return &RoomRepository{client: client}
}

func (r *RoomRepository) metaKey(code string) string {
	return fmt.Sprintf("room:%s:meta", code)
}

func (r *RoomRepository) playersKey(code string) string {
	return fmt.Sprintf("room:%s:players", code)
}

func (r *RoomRepository) chatKey(code string) string {
	return fmt.Sprintf("room:%s:chat", code)
}

func (r *RoomRepository) sessionKey(token string) string {
	return fmt.Sprintf("session:%s", token)
}

// CreateRoom initialise une room dans Redis
func (r *RoomRepository) CreateRoom(ctx context.Context, room *domain.Room) error {
	settingsJSON, err := json.Marshal(room.Settings)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	metaKey := r.metaKey(room.Code)

	metaValues := map[string]interface{}{
		"code":          room.Code,
		"status":        string(room.Status),
		"master_id":     room.MasterID.String(),
		"settings":      string(settingsJSON),
		"current_round": room.CurrentRound,
		"created_at":    room.CreatedAt.Format(time.RFC3339),
	}

	pipe.HSet(ctx, metaKey, metaValues)
	pipe.Expire(ctx, metaKey, 2*time.Hour)

	// Ajouter au set global des rooms actives
	pipe.SAdd(ctx, "rooms:active", room.Code)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) GetRoom(ctx context.Context, code string) (*domain.Room, error) {
	metaKey := r.metaKey(code)
	exists, err := r.client.Exists(ctx, metaKey).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, ErrRoomNotFound
	}

	meta, err := r.client.HGetAll(ctx, metaKey).Result()
	if err != nil {
		return nil, err
	}

	masterID, _ := uuid.Parse(meta["master_id"])
	createdAt, _ := time.Parse(time.RFC3339, meta["created_at"])

	var settings domain.RoomSettings
	_ = json.Unmarshal([]byte(meta["settings"]), &settings)

	var endsAt *time.Time
	if endsAtStr, ok := meta["ends_at"]; ok && endsAtStr != "" {
		t, err := time.Parse(time.RFC3339, endsAtStr)
		if err == nil {
			endsAt = &t
		}
	}

	players, _ := r.GetPlayers(ctx, code)

	currentRound := 1
	if cr, ok := meta["current_round"]; ok {
		fmt.Sscanf(cr, "%d", &currentRound)
	}

	return &domain.Room{
		Code:         code,
		Status:       domain.RoomStatus(meta["status"]),
		MasterID:     masterID,
		Settings:     settings,
		CurrentRound: currentRound,
		EndsAt:       endsAt,
		CreatedAt:    createdAt,
		Players:      players,
	}, nil
}

func (r *RoomRepository) UpdateRoomStatus(ctx context.Context, code string, status domain.RoomStatus) error {
	return r.client.HSet(ctx, r.metaKey(code), "status", string(status)).Err()
}

func (r *RoomRepository) UpdateRoomSettings(ctx context.Context, code string, settings domain.RoomSettings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	return r.client.HSet(ctx, r.metaKey(code), "settings", string(data)).Err()
}

func (r *RoomRepository) UpdateRound(ctx context.Context, code string, round int, endsAt *time.Time) error {
	pipe := r.client.Pipeline()
	metaKey := r.metaKey(code)
	pipe.HSet(ctx, metaKey, "current_round", round)
	if endsAt != nil {
		pipe.HSet(ctx, metaKey, "ends_at", endsAt.Format(time.RFC3339))
	} else {
		pipe.HDel(ctx, metaKey, "ends_at")
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) SetMaster(ctx context.Context, code string, masterID uuid.UUID) error {
	pipe := r.client.Pipeline()
	pipe.HSet(ctx, r.metaKey(code), "master_id", masterID.String())

	// Mettre à jour le rôle dans le hash des joueurs
	players, _ := r.GetPlayers(ctx, code)
	for _, p := range players {
		if p.UserID == masterID {
			p.Role = domain.RoleMaster
		} else if p.Role == domain.RoleMaster {
			p.Role = domain.RolePlayer
		}
		data, _ := json.Marshal(p)
		pipe.HSet(ctx, r.playersKey(code), p.UserID.String(), string(data))
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Players
func (r *RoomRepository) AddPlayer(ctx context.Context, code string, player *domain.RoomPlayer) error {
	data, err := json.Marshal(player)
	if err != nil {
		return err
	}
	pipe := r.client.Pipeline()
	pipe.HSet(ctx, r.playersKey(code), player.UserID.String(), string(data))
	pipe.Expire(ctx, r.playersKey(code), 2*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) GetPlayer(ctx context.Context, code string, userID uuid.UUID) (*domain.RoomPlayer, error) {
	data, err := r.client.HGet(ctx, r.playersKey(code), userID.String()).Result()
	if err != nil {
		return nil, err
	}
	var p domain.RoomPlayer
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *RoomRepository) GetPlayers(ctx context.Context, code string) ([]domain.RoomPlayer, error) {
	data, err := r.client.HGetAll(ctx, r.playersKey(code)).Result()
	if err != nil {
		return nil, err
	}

	var players []domain.RoomPlayer
	for _, pStr := range data {
		var p domain.RoomPlayer
		if err := json.Unmarshal([]byte(pStr), &p); err == nil {
			players = append(players, p)
		}
	}

	// Trier par ordre de connexion / joined_at
	sort.Slice(players, func(i, j int) bool {
		return players[i].JoinedAt.Before(players[j].JoinedAt)
	})

	return players, nil
}

func (r *RoomRepository) RemovePlayer(ctx context.Context, code string, userID uuid.UUID) error {
	return r.client.HDel(ctx, r.playersKey(code), userID.String()).Err()
}

func (r *RoomRepository) UpdatePlayerActivity(ctx context.Context, code string, userID uuid.UUID, connected bool) error {
	p, err := r.GetPlayer(ctx, code, userID)
	if err != nil {
		return err
	}
	p.IsConnected = connected
	p.LastSeenAt = time.Now()
	return r.AddPlayer(ctx, code, p)
}

func (r *RoomRepository) UpdatePlayerMute(ctx context.Context, code string, userID uuid.UUID, muted bool) error {
	p, err := r.GetPlayer(ctx, code, userID)
	if err != nil {
		return err
	}
	p.IsMuted = muted
	return r.AddPlayer(ctx, code, p)
}

func (r *RoomRepository) UpdatePlayerScore(ctx context.Context, code string, userID uuid.UUID, scoreDelta int) error {
	p, err := r.GetPlayer(ctx, code, userID)
	if err != nil {
		return err
	}
	p.Score += scoreDelta
	return r.AddPlayer(ctx, code, p)
}

// Chat Buffer (50 messages max)
func (r *RoomRepository) AddChatMessage(ctx context.Context, code string, msg *domain.ChatMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	key := r.chatKey(code)
	pipe := r.client.Pipeline()
	pipe.LPush(ctx, key, string(data))
	pipe.LTrim(ctx, key, 0, 49) // Garder les 50 derniers
	pipe.Expire(ctx, key, 2*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) GetRecentChat(ctx context.Context, code string) ([]domain.ChatMessage, error) {
	key := r.chatKey(code)
	items, err := r.client.LRange(ctx, key, 0, 49).Result()
	if err != nil {
		return nil, err
	}

	var messages []domain.ChatMessage
	for _, item := range items {
		var m domain.ChatMessage
		if err := json.Unmarshal([]byte(item), &m); err == nil {
			messages = append(messages, m)
		}
	}

	// Inverser pour ordre chronologique (du plus ancien au plus récent)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// Banned & Muted sets
func (r *RoomRepository) BanUser(ctx context.Context, code string, userID uuid.UUID, duration time.Duration) error {
	key := fmt.Sprintf("room:%s:bans", code)
	return r.client.Set(ctx, fmt.Sprintf("%s:%s", key, userID.String()), "1", duration).Err()
}

func (r *RoomRepository) IsBanned(ctx context.Context, code string, userID uuid.UUID) bool {
	key := fmt.Sprintf("room:%s:bans:%s", code, userID.String())
	exists, err := r.client.Exists(ctx, key).Result()
	return err == nil && exists > 0
}

// Active rooms listing & cleanup
func (r *RoomRepository) GetActiveRoomCodes(ctx context.Context) ([]string, error) {
	return r.client.SMembers(ctx, "rooms:active").Result()
}

func (r *RoomRepository) CloseAndPurgeRoom(ctx context.Context, code string) error {
	pipe := r.client.Pipeline()
	pipe.SRem(ctx, "rooms:active", code)
	pipe.Del(ctx, r.metaKey(code))
	pipe.Del(ctx, r.playersKey(code))
	pipe.Del(ctx, r.chatKey(code))
	pipe.Del(ctx, fmt.Sprintf("room:%s:game_state", code))
	pipe.Del(ctx, fmt.Sprintf("room:%s:wordle", code))
	_, err := pipe.Exec(ctx)
	return err
}

// Rate Limiting par token-bucket Redis
func (r *RoomRepository) CheckRateLimit(ctx context.Context, key string, maxRequests int, window time.Duration) (bool, error) {
	rateKey := fmt.Sprintf("ratelimit:%s", key)
	count, err := r.client.Incr(ctx, rateKey).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		r.client.Expire(ctx, rateKey, window)
	}
	return count <= int64(maxRequests), nil
}
