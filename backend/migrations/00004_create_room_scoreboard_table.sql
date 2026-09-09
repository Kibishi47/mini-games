-- Migration 00004: Création de la table room_scoreboard
CREATE TABLE IF NOT EXISTS room_scoreboard (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_session_id UUID NOT NULL REFERENCES room_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_type VARCHAR(32) NOT NULL DEFAULT 'wordle',
    round_number INT NOT NULL DEFAULT 1,
    score_delta INT NOT NULL DEFAULT 0,
    final_score INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_room_scoreboard_session ON room_scoreboard(room_session_id);
