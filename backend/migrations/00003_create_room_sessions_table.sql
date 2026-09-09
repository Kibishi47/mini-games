-- Migration 00003: Création de la table room_sessions
CREATE TABLE IF NOT EXISTS room_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_code VARCHAR(16) NOT NULL,
    master_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'in_lobby',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_room_sessions_code ON room_sessions(room_code);
CREATE INDEX IF NOT EXISTS idx_room_sessions_status ON room_sessions(status);
