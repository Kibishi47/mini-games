package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"minigames-backend/internal/domain"
)

var (
	ErrUserNotFound      = errors.New("utilisateur non trouvé")
	ErrUserAlreadyExists = errors.New("nom d'utilisateur ou email déjà utilisé")
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, username, display_username, email, password_hash, discord_id, avatar_url, is_guest, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		u.ID, u.Username, u.DisplayUsername, u.Email, u.PasswordHash, u.DiscordID, u.AvatarURL, u.IsGuest, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("création utilisateur: %w", err)
	}

	// Initialisation des stats par défaut (Wordle)
	_, _ = r.pool.Exec(ctx, `
		INSERT INTO user_stats (user_id, game_type, games_played, games_won, win_streak, highest_score)
		VALUES ($1, 'wordle', 0, 0, 0, 0)
		ON CONFLICT DO NOTHING
	`, u.ID)

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, username, display_username, email, password_hash, discord_id, avatar_url, is_guest, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.DisplayUsername, &u.Email, &u.PasswordHash, &u.DiscordID, &u.AvatarURL, &u.IsGuest, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, display_username, email, password_hash, discord_id, avatar_url, is_guest, created_at, updated_at
		FROM users
		WHERE LOWER(username) = LOWER($1)
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.DisplayUsername, &u.Email, &u.PasswordHash, &u.DiscordID, &u.AvatarURL, &u.IsGuest, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByDiscordID(ctx context.Context, discordID string) (*domain.User, error) {
	query := `
		SELECT id, username, display_username, email, password_hash, discord_id, avatar_url, is_guest, created_at, updated_at
		FROM users
		WHERE discord_id = $1
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, discordID).Scan(
		&u.ID, &u.Username, &u.DisplayUsername, &u.Email, &u.PasswordHash, &u.DiscordID, &u.AvatarURL, &u.IsGuest, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, displayUsername string, avatarURL string) error {
	query := `
		UPDATE users
		SET display_username = $2, avatar_url = $3, updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, displayUsername, avatarURL)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) UpgradeGuest(ctx context.Context, id uuid.UUID, username, displayUsername string, email, passwordHash string) error {
	query := `
		UPDATE users
		SET username = $2, display_username = $3, email = $4, password_hash = $5, is_guest = FALSE, updated_at = NOW()
		WHERE id = $1 AND is_guest = TRUE
	`
	tag, err := r.pool.Exec(ctx, query, id, username, displayUsername, email, passwordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("impossible de convertir ce compte invité")
	}
	return nil
}

func (r *UserRepository) GetStats(ctx context.Context, userID uuid.UUID, gameType string) (*domain.UserStats, error) {
	query := `
		SELECT user_id, game_type, games_played, games_won, win_streak, highest_score
		FROM user_stats
		WHERE user_id = $1 AND game_type = $2
	`
	var s domain.UserStats
	err := r.pool.QueryRow(ctx, query, userID, gameType).Scan(
		&s.UserID, &s.GameType, &s.GamesPlayed, &s.GamesWon, &s.WinStreak, &s.HighestScore,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.UserStats{UserID: userID, GameType: gameType}, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *UserRepository) RecordGameResult(ctx context.Context, userID uuid.UUID, gameType string, won bool, score int) error {
	query := `
		INSERT INTO user_stats (user_id, game_type, games_played, games_won, win_streak, highest_score)
		VALUES ($1, $2, 1, CASE WHEN $3 THEN 1 ELSE 0 END, CASE WHEN $3 THEN 1 ELSE 0 END, $4)
		ON CONFLICT (user_id, game_type) DO UPDATE SET
			games_played = user_stats.games_played + 1,
			games_won = user_stats.games_won + CASE WHEN $3 THEN 1 ELSE 0 END,
			win_streak = CASE WHEN $3 THEN user_stats.win_streak + 1 ELSE 0 END,
			highest_score = GREATEST(user_stats.highest_score, $4)
	`
	_, err := r.pool.Exec(ctx, query, userID, gameType, won, score)
	return err
}
