package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"minigames-backend/internal/domain"
)

var (
	ErrRoomNotFound   = errors.New("salle introuvable ou fermée")
	ErrPlayerNotFound = errors.New("joueur introuvable dans la salle")
	ErrSessionNotFound = errors.New("session introuvable ou expirée")
	ErrPlayerBanned   = errors.New("vous avez été banni de cette salle")
)

type RoomRepository struct {
	client *redis.Client
}

func NewRoomRepository(client *redis.Client) *RoomRepository {
	return &RoomRepository{client: client}
}

// Helpers pour les clés Redis
func (r *RoomRepository) metaKey(code string) string {
	return fmt.Sprintf("room:%s:meta", code)
}

func (r *RoomRepository) playersKey(code string) string {
	return fmt.Sprintf("room:%s:players", code)
}

func (r *RoomRepository) scoresKey(code string) string {
	return fmt.Sprintf("room:%s:scores", code)
}

func (r *RoomRepository) historyKey(code string) string {
	return fmt.Sprintf("room:%s:history", code)
}

func (r *RoomRepository) chatKey(code string) string {
	return fmt.Sprintf("room:%s:chat", code)
}

func (r *RoomRepository) bansKey(code string) string {
	return fmt.Sprintf("room:%s:bans", code)
}

func (r *RoomRepository) sessionKey(token string) string {
	return fmt.Sprintf("session:%s", token)
}

// -----------------------------------------------------------------------------
// SESSIONS ÉPHÉMÈRES (TTL 45s pour reconnexion transparente)
// -----------------------------------------------------------------------------

func (r *RoomRepository) CreateSession(ctx context.Context, token string, data *domain.SessionData) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.sessionKey(token), bytes, 24*time.Hour).Err()
}

func (r *RoomRepository) GetSession(ctx context.Context, token string) (*domain.SessionData, error) {
	bytes, err := r.client.Get(ctx, r.sessionKey(token)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	var data domain.SessionData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *RoomRepository) RefreshSession(ctx context.Context, token string) error {
	return r.client.Expire(ctx, r.sessionKey(token), 24*time.Hour).Err()
}

// -----------------------------------------------------------------------------
// SALLES (Création, Lecture, Mise à jour, Fermeture)
// -----------------------------------------------------------------------------

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
		"secret_word":   room.SecretWord,
		"created_at":    room.CreatedAt.Format(time.RFC3339),
	}

	pipe.HSet(ctx, metaKey, metaValues)
	pipe.Expire(ctx, metaKey, 4*time.Hour)
	pipe.SAdd(ctx, "rooms:active", room.Code)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) GetRoom(ctx context.Context, code string) (*domain.Room, error) {
	metaKey := r.metaKey(code)
	data, err := r.client.HGetAll(ctx, metaKey).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrRoomNotFound
	}
	if data["status"] == string(domain.RoomStatusClosed) {
		return nil, ErrRoomNotFound
	}

	masterID, _ := uuid.Parse(data["master_id"])
	currentRound, _ := strconv.Atoi(data["current_round"])
	createdAt, _ := time.Parse(time.RFC3339, data["created_at"])

	var settings domain.RoomSettings
	if sJSON, ok := data["settings"]; ok {
		_ = json.Unmarshal([]byte(sJSON), &settings)
	}

	var endsAt *time.Time
	if eaStr, ok := data["ends_at"]; ok && eaStr != "" {
		if t, err := time.Parse(time.RFC3339, eaStr); err == nil {
			endsAt = &t
		}
	}

	var nextRoundAt *time.Time
	if nraStr, ok := data["next_round_at"]; ok && nraStr != "" {
		if t, err := time.Parse(time.RFC3339, nraStr); err == nil {
			nextRoundAt = &t
		}
	}

	players, _ := r.GetPlayers(ctx, code)

	return &domain.Room{
		Code:         code,
		Status:       domain.RoomStatus(data["status"]),
		RoundState:   domain.RoundSubState(data["round_state"]),
		MasterID:     masterID,
		Settings:     settings,
		CurrentRound: currentRound,
		SecretWord:   data["secret_word"],
		RevealedWord: data["revealed_word"],
		EndsAt:       endsAt,
		NextRoundAt:  nextRoundAt,
		CreatedAt:    createdAt,
		Players:      players,
	}, nil
}

func (r *RoomRepository) UpdateRoomStatus(ctx context.Context, code string, status domain.RoomStatus) error {
	return r.client.HSet(ctx, r.metaKey(code), "status", string(status)).Err()
}

func (r *RoomRepository) SetRoundState(ctx context.Context, code string, state domain.RoundSubState, revealedWord string, nextRoundAt *time.Time) error {
	vals := map[string]interface{}{
		"round_state":   string(state),
		"revealed_word": revealedWord,
	}
	if nextRoundAt != nil {
		vals["next_round_at"] = nextRoundAt.Format(time.RFC3339)
	} else {
		vals["next_round_at"] = ""
	}
	return r.client.HSet(ctx, r.metaKey(code), vals).Err()
}

func (r *RoomRepository) UpdateRoomSettings(ctx context.Context, code string, settings domain.RoomSettings) error {
	b, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	return r.client.HSet(ctx, r.metaKey(code), "settings", string(b)).Err()
}

func (r *RoomRepository) UpdateRoomMaster(ctx context.Context, code string, newMasterID uuid.UUID) error {
	pipe := r.client.Pipeline()
	pipe.HSet(ctx, r.metaKey(code), "master_id", newMasterID.String())

	// Mettre à jour l'ancien et le nouveau master dans le hash players
	players, err := r.GetPlayers(ctx, code)
	if err == nil {
		for _, p := range players {
			if p.ID == newMasterID {
				p.Role = domain.RoleMaster
				p.IsMaster = true
				b, _ := json.Marshal(p)
				pipe.HSet(ctx, r.playersKey(code), p.ID.String(), string(b))
			} else if p.IsMaster {
				p.Role = domain.RolePlayer
				p.IsMaster = false
				b, _ := json.Marshal(p)
				pipe.HSet(ctx, r.playersKey(code), p.ID.String(), string(b))
			}
		}
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) SetSecretWord(ctx context.Context, code string, secretWord string) error {
	return r.client.HSet(ctx, r.metaKey(code), "secret_word", secretWord).Err()
}

func (r *RoomRepository) UpdateRound(ctx context.Context, code string, round int, endsAt *time.Time) error {
	vals := map[string]interface{}{
		"current_round": round,
	}
	if endsAt != nil {
		vals["ends_at"] = endsAt.Format(time.RFC3339)
	} else {
		vals["ends_at"] = ""
	}
	return r.client.HSet(ctx, r.metaKey(code), vals).Err()
}

func (r *RoomRepository) CloseRoom(ctx context.Context, code string) error {
	pipe := r.client.Pipeline()
	pipe.SRem(ctx, "rooms:active", code)
	pipe.Del(ctx,
		r.metaKey(code),
		r.playersKey(code),
		r.scoresKey(code),
		r.historyKey(code),
		r.chatKey(code),
		r.bansKey(code),
	)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) GetActiveRoomCodes(ctx context.Context) ([]string, error) {
	return r.client.SMembers(ctx, "rooms:active").Result()
}

// -----------------------------------------------------------------------------
// JOUEURS & PRÉSENCE
// -----------------------------------------------------------------------------

func (r *RoomRepository) AddPlayer(ctx context.Context, code string, player *domain.RoomPlayer) error {
	// Vérifier si banni
	isBanned, err := r.client.SIsMember(ctx, r.bansKey(code), player.ID.String()).Result()
	if err == nil && isBanned {
		return ErrPlayerBanned
	}

	data, err := json.Marshal(player)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.HSet(ctx, r.playersKey(code), player.ID.String(), string(data))
	pipe.Expire(ctx, r.playersKey(code), 4*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) GetPlayer(ctx context.Context, code string, userID uuid.UUID) (*domain.RoomPlayer, error) {
	val, err := r.client.HGet(ctx, r.playersKey(code), userID.String()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrPlayerNotFound
		}
		return nil, err
	}

	var p domain.RoomPlayer
	if err := json.Unmarshal([]byte(val), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *RoomRepository) GetPlayers(ctx context.Context, code string) ([]domain.RoomPlayer, error) {
	data, err := r.client.HGetAll(ctx, r.playersKey(code)).Result()
	if err != nil {
		return nil, err
	}

	players := make([]domain.RoomPlayer, 0, len(data))
	for _, val := range data {
		var p domain.RoomPlayer
		if err := json.Unmarshal([]byte(val), &p); err == nil {
			players = append(players, p)
		}
	}

	// Tri par date d'arrivée pour préserver l'ordre d'ancienneté (passation Master)
	sort.Slice(players, func(i, j int) bool {
		return players[i].JoinedAt.Before(players[j].JoinedAt)
	})

	return players, nil
}

func (r *RoomRepository) RemovePlayer(ctx context.Context, code string, userID uuid.UUID) error {
	return r.client.HDel(ctx, r.playersKey(code), userID.String()).Err()
}

func (r *RoomRepository) UpdatePlayerActivity(ctx context.Context, code string, userID uuid.UUID, isConnected bool) error {
	p, err := r.GetPlayer(ctx, code, userID)
	if err != nil {
		return err
	}

	p.IsConnected = isConnected
	p.LastSeenAt = time.Now()

	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return r.client.HSet(ctx, r.playersKey(code), userID.String(), string(data)).Err()
}

func (r *RoomRepository) SetPlayerMuted(ctx context.Context, code string, userID uuid.UUID, muted bool) error {
	p, err := r.GetPlayer(ctx, code, userID)
	if err != nil {
		return err
	}

	p.IsMuted = muted
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return r.client.HSet(ctx, r.playersKey(code), userID.String(), string(data)).Err()
}

func (r *RoomRepository) SetPlayerSpectator(ctx context.Context, code string, userID uuid.UUID, isSpectator bool) error {
	p, err := r.GetPlayer(ctx, code, userID)
	if err != nil {
		return err
	}

	p.IsSpectator = isSpectator
	if isSpectator {
		p.Role = domain.RoleSpectator
	} else if p.IsMaster {
		p.Role = domain.RoleMaster
	} else {
		p.Role = domain.RolePlayer
	}

	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return r.client.HSet(ctx, r.playersKey(code), userID.String(), string(data)).Err()
}

func (r *RoomRepository) BanPlayer(ctx context.Context, code string, userID uuid.UUID) error {
	pipe := r.client.Pipeline()
	pipe.SAdd(ctx, r.bansKey(code), userID.String())
	pipe.HDel(ctx, r.playersKey(code), userID.String())
	pipe.Expire(ctx, r.bansKey(code), 4*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) IsPlayerBanned(ctx context.Context, code string, userID uuid.UUID) (bool, error) {
	return r.client.SIsMember(ctx, r.bansKey(code), userID.String()).Result()
}

// -----------------------------------------------------------------------------
// SCOREBOARD DE SESSION (ZSet / Hash)
// -----------------------------------------------------------------------------

func (r *RoomRepository) AddScore(ctx context.Context, code string, userID uuid.UUID, points int) (int, error) {
	newScore, err := r.client.ZIncrBy(ctx, r.scoresKey(code), float64(points), userID.String()).Result()
	if err != nil {
		return 0, err
	}
	r.client.Expire(ctx, r.scoresKey(code), 4*time.Hour)

	// Synchroniser dans l'objet joueur
	if p, err := r.GetPlayer(ctx, code, userID); err == nil {
		p.Score = int(newScore)
		if data, err := json.Marshal(p); err == nil {
			_ = r.client.HSet(ctx, r.playersKey(code), userID.String(), string(data))
		}
	}

	return int(newScore), nil
}

func (r *RoomRepository) GetScores(ctx context.Context, code string) (map[string]int, error) {
	scores, err := r.client.ZRevRangeWithScores(ctx, r.scoresKey(code), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	res := make(map[string]int, len(scores))
	for _, z := range scores {
		res[fmt.Sprint(z.Member)] = int(z.Score)
	}
	return res, nil
}

// -----------------------------------------------------------------------------
// HISTORIQUE DES MANCHES
// -----------------------------------------------------------------------------

func (r *RoomRepository) AddRoundHistory(ctx context.Context, code string, history *domain.RoundHistory) error {
	bytes, err := json.Marshal(history)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.RPush(ctx, r.historyKey(code), bytes)
	pipe.Expire(ctx, r.historyKey(code), 4*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) GetRoundHistory(ctx context.Context, code string) ([]domain.RoundHistory, error) {
	items, err := r.client.LRange(ctx, r.historyKey(code), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	history := make([]domain.RoundHistory, 0, len(items))
	for _, item := range items {
		var h domain.RoundHistory
		if err := json.Unmarshal([]byte(item), &h); err == nil {
			history = append(history, h)
		}
	}
	return history, nil
}

// -----------------------------------------------------------------------------
// CHAT CIRCULAIRE (50 derniers messages)
// -----------------------------------------------------------------------------

func (r *RoomRepository) AddChatMessage(ctx context.Context, code string, msg *domain.ChatMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	chatKey := r.chatKey(code)
	pipe.RPush(ctx, chatKey, string(data))
	pipe.LTrim(ctx, chatKey, -50, -1) // Ne garder que les 50 plus récents
	pipe.Expire(ctx, chatKey, 4*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RoomRepository) GetRecentChatMessages(ctx context.Context, code string) ([]domain.ChatMessage, error) {
	items, err := r.client.LRange(ctx, r.chatKey(code), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]domain.ChatMessage, 0, len(items))
	for _, item := range items {
		var msg domain.ChatMessage
		if err := json.Unmarshal([]byte(item), &msg); err == nil {
			messages = append(messages, msg)
		}
	}
	return messages, nil
}
