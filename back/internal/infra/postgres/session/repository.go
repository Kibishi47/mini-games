package session

import (
	"context"

	domainsession "github.com/Kibishi47/mini-games/back/internal/domain/session"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresSessionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSessionRepository(pool *pgxpool.Pool) domainsession.Repository {
	return &postgresSessionRepository{pool: pool}
}

func (r *postgresSessionRepository) Create(ctx context.Context, session *domainsession.GameSession) error {
	query := `
		INSERT INTO game_session (room_id, game, config, started_at, ended_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		session.RoomID,
		session.Game,
		session.Config,
		session.StartedAt,
		session.EndedAt,
	).Scan(&session.ID)

	return err
}

func (r *postgresSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainsession.GameSession, error) {
	query := `
		SELECT id, room_id, game, config, started_at, ended_at
		FROM game_session
		WHERE id = $1
	`

	var session domainsession.GameSession
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.RoomID,
		&session.Game,
		&session.Config,
		&session.StartedAt,
		&session.EndedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // or return custom error
		}
		return nil, err
	}

	return &session, nil
}

func (r *postgresSessionRepository) GetActiveSessionByRoomID(ctx context.Context, roomID uuid.UUID) (*domainsession.GameSession, error) {
	query := `
		SELECT id, room_id, game, config, started_at, ended_at
		FROM game_session
		WHERE room_id = $1 AND ended_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`

	var session domainsession.GameSession
	err := r.pool.QueryRow(ctx, query, roomID).Scan(
		&session.ID,
		&session.RoomID,
		&session.Game,
		&session.Config,
		&session.StartedAt,
		&session.EndedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &session, nil
}
func (r *postgresSessionRepository) EndSession(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE game_session
		SET ended_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
