-- Migration: 000004_create_prizes.up.sql
CREATE TABLE IF NOT EXISTS prizes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    image_url VARCHAR(512),
    rank_required INTEGER NOT NULL DEFAULT 1, -- e.g., 1 for 1st place
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prizes_challenge_id ON prizes(challenge_id);
