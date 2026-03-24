-- Migration: 000006_add_creator_id_to_challenges.up.sql
ALTER TABLE challenges ADD COLUMN IF NOT EXISTS creator_id UUID;
ALTER TABLE challenges ALTER COLUMN status SET DEFAULT 'draft';

-- Add foreign key constraint (assuming users table exists)
ALTER TABLE challenges ADD CONSTRAINT fk_challenges_creator FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_challenges_creator ON challenges(creator_id);
