-- Migration: 000009_create_daily_logs.up.sql
CREATE TABLE IF NOT EXISTS daily_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    log_date DATE NOT NULL,
    steps INTEGER NOT NULL DEFAULT 0,
    calories INTEGER NOT NULL DEFAULT 0,
    active_minutes INTEGER NOT NULL DEFAULT 0,
    score NUMERIC(7,2) NOT NULL DEFAULT 0,
    healthkit_data_hash VARCHAR(128),
    source_bundle_ids TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, challenge_id, log_date),
    CHECK (steps >= 0),
    CHECK (calories >= 0),
    CHECK (active_minutes >= 0),
    CHECK (score >= 0),
    CHECK (steps > 0 OR calories > 0 OR active_minutes > 0)
);

CREATE INDEX idx_daily_logs_user_challenge_date ON daily_logs(user_id, challenge_id, log_date DESC);
CREATE INDEX idx_daily_logs_challenge_date ON daily_logs(challenge_id, log_date DESC);
