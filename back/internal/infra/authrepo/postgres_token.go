package authrepo

import (
	"context"
	"time"

	"github.com/Kibishi47/mini-games/back/internal/domain/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTokenRepository struct {
	db *pgxpool.Pool
}

func NewPostgresTokenRepository(db *pgxpool.Pool) auth.TokenRepository {
	return &PostgresTokenRepository{db: db}
}

func (r *PostgresTokenRepository) CreateToken(
	ctx context.Context,
	userID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO auth_token (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)

	return err
}

func (r *PostgresTokenRepository) FindUserIDByTokenHash(
	ctx context.Context,
	tokenHash string,
) (uuid.UUID, time.Time, error) {

	var userID uuid.UUID
	var expiresAt time.Time

	err := r.db.QueryRow(
		ctx,
		`SELECT user_id, expires_at
		 FROM auth_token
		 WHERE token_hash = $1`,
		tokenHash,
	).Scan(&userID, &expiresAt)

	if err != nil {
		return uuid.Nil, time.Time{}, err
	}

	return userID, expiresAt, nil
}

func (r *PostgresTokenRepository) DeleteTokenHash(
	ctx context.Context,
	tokenHash string,
) error {

	_, err := r.db.Exec(
		ctx,
		`DELETE FROM auth_token WHERE token_hash = $1`,
		tokenHash,
	)

	return err
}
