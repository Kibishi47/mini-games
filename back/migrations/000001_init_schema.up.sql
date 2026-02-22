-- extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- enums
DO $$ BEGIN
CREATE TYPE game AS ENUM (
    'Wordle','BetweenLines','Checkers','Chess','FourInARow','Trivia',
    'BombDroper','Tapple','Wavelength','Snake','Minesweeper','Werewolf'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
CREATE TYPE room_status AS ENUM ('lobby','running','closed');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
CREATE TYPE auth_provider AS ENUM ('local','discord');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- tables
CREATE TABLE IF NOT EXISTS "user" (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username varchar NOT NULL UNIQUE,
    password_hash varchar NULL,
    display_name varchar NULL,
    avatar_url varchar NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth_identity (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    provider auth_provider NOT NULL,
    provider_id varchar NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS auth_identity_user_provider_uq
    ON auth_identity(user_id, provider);

CREATE UNIQUE INDEX IF NOT EXISTS auth_identity_provider_providerid_uq
    ON auth_identity(provider, provider_id);

CREATE TABLE IF NOT EXISTS auth_token (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    token_hash varchar NOT NULL,
    expires_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS auth_token_user_id_idx
    ON auth_token(user_id);

CREATE INDEX IF NOT EXISTS auth_token_expires_at_idx
    ON auth_token(expires_at);

CREATE TABLE IF NOT EXISTS room (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code varchar NOT NULL UNIQUE,
    status room_status NOT NULL,
    host_id uuid NOT NULL REFERENCES "user"(id),
    max_players integer NOT NULL DEFAULT 2,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS game_session (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id uuid NOT NULL REFERENCES room(id) ON DELETE CASCADE,
    game game NOT NULL,
    config jsonb NOT NULL DEFAULT '{}'::jsonb,
    started_at timestamptz NULL,
    ended_at timestamptz NULL
);

CREATE TABLE IF NOT EXISTS game_result (
    user_id uuid NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    session_id uuid NOT NULL REFERENCES game_session(id) ON DELETE CASCADE,
    score integer NOT NULL,
    details jsonb NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (user_id, session_id)
);