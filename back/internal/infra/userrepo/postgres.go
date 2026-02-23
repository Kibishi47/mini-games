package userrepo

import (
	"context"

	"github.com/Kibishi47/mini-games/back/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) user.Repository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u := &user.User{}

	err := r.db.QueryRow(
		ctx,
		`SELECT id, username, password_hash, display_name, avatar_url, created_at
		FROM "user"
		WHERE id = $1`,
		id,
	).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.DisplayName,
		&u.AvatarURL,
		&u.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *user.User) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO "user" (id, username, password_hash, display_name, avatar_url, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID,
		u.Username,
		u.PasswordHash,
		u.DisplayName,
		u.AvatarURL,
		u.CreatedAt,
	)

	return err
}

func (r *PostgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM "user" WHERE username = $1)`,
		username,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}
