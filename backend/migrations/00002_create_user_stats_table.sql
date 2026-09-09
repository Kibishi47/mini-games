-- Migration 00002: Création de la table user_stats
CREATE TABLE IF NOT EXISTS user_stats (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_type VARCHAR(32) NOT NULL DEFAULT 'wordle',
    games_played INT NOT NULL DEFAULT 0,
    games_won INT NOT NULL DEFAULT 0,
    win_streak INT NOT NULL DEFAULT 0,
    highest_score INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, game_type)
);
