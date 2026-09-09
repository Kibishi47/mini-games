package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"minigames-backend/internal/domain"
)

type RoomSessionRepository struct {
	pool *pgxpool.Pool
}

func NewRoomSessionRepository(pool *pgxpool.Pool) *RoomSessionRepository {
	return &RoomSessionRepository{pool: pool}
}

func (r *RoomSessionRepository) CreateSession(ctx context.Context, code string, masterID uuid.UUID) (uuid.UUID, error) {
	id := uuid.New()
	query := `
		INSERT INTO room_sessions (id, room_code, master_user_id, status, created_at)
		VALUES ($1, $2, $3, 'in_lobby', NOW())
	`
	_, err := r.pool.Exec(ctx, query, id, code, masterID)
	return id, err
}

func (r *RoomSessionRepository) UpdateStatus(ctx context.Context, code string, status domain.RoomStatus) error {
	query := `
		UPDATE room_sessions
		SET status = $2, closed_at = CASE WHEN $2 = 'closed' THEN NOW() ELSE closed_at END
		WHERE room_code = $1 AND status != 'closed'
	`
	_, err := r.pool.Exec(ctx, query, code, string(status))
	return err
}

func (r *RoomSessionRepository) RecordRoundScore(ctx context.Context, sessionID, userID uuid.UUID, gameType string, roundNumber, scoreDelta, finalScore int) error {
	query := `
		INSERT INTO room_scoreboard (id, room_session_id, user_id, game_type, round_number, score_delta, final_score, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := r.pool.Exec(ctx, query, uuid.New(), sessionID, userID, gameType, roundNumber, scoreDelta, finalScore)
	return err
}

func (r *RoomSessionRepository) GetSessionIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	var id uuid.UUID
	query := `
		SELECT id FROM room_sessions
		WHERE room_code = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.pool.QueryRow(ctx, query, code).Scan(&id)
	return id, err
}
