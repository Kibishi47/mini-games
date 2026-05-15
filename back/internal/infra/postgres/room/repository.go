package room

import (
	"context"

	domainroom "github.com/Kibishi47/mini-games/back/internal/domain/room"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRoomRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRoomRepository(pool *pgxpool.Pool) domainroom.Repository {
	return &postgresRoomRepository{pool: pool}
}

func (r *postgresRoomRepository) Create(ctx context.Context, room *domainroom.Room) error {
	query := `
		INSERT INTO room (code, status, host_id, max_players)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(ctx, query,
		room.Code,
		room.Status,
		room.HostID,
		room.MaxPlayers,
	).Scan(&room.ID, &room.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *postgresRoomRepository) GetByCode(ctx context.Context, code string) (*domainroom.Room, error) {
	query := `
		SELECT id, code, status, host_id, max_players, created_at
		FROM room
		WHERE code = $1
	`

	var room domainroom.Room
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&room.ID,
		&room.Code,
		&room.Status,
		&room.HostID,
		&room.MaxPlayers,
		&room.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainroom.ErrRoomNotFound
		}
		return nil, err
	}

	return &room, nil
}
