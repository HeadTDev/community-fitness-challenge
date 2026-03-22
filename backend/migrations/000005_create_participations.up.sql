-- Migration: 000005_create_participations.up.sql
CREATE TABLE IF NOT EXISTS participations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    current_score INTEGER NOT NULL DEFAULT 0,
    rank INTEGER DEFAULT 0,
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, challenge_id)
);

CREATE INDEX idx_participations_user ON participations(user_id);
CREATE INDEX idx_participations_challenge ON participations(challenge_id);
CREATE INDEX idx_participations_score ON participations(current_score DESC);
