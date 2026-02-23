package authrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Kibishi47/mini-games/back/internal/domain/auth"
)

type PostgresIdentityRepository struct {
	db *pgxpool.Pool
}

func NewPostgresIdentityRepository(db *pgxpool.Pool) auth.IdentityRepository {
	return &PostgresIdentityRepository{db: db}
}

func (r *PostgresIdentityRepository) FindUserIDByProvider(
	ctx context.Context,
	provider auth.Provider,
	providerID string,
) (uuid.UUID, error) {

	var userID uuid.UUID

	err := r.db.QueryRow(
		ctx,
		`SELECT user_id
		 FROM auth_identity
		 WHERE provider = $1 AND provider_id = $2`,
		string(provider),
		providerID,
	).Scan(&userID)

	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func (r *PostgresIdentityRepository) CreateIdentity(
	ctx context.Context,
	userID uuid.UUID,
	provider auth.Provider,
	providerID string,
) error {

	_, err := r.db.Exec(
		ctx,
		`INSERT INTO auth_identity (user_id, provider, provider_id)
		 VALUES ($1, $2, $3)`,
		userID,
		string(provider),
		providerID,
	)

	return err
}
