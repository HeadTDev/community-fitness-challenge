-- Migration: 000006_add_creator_id_to_challenges.down.sql
ALTER TABLE challenges DROP CONSTRAINT IF EXISTS fk_challenges_creator;
DROP INDEX IF EXISTS idx_challenges_creator;
ALTER TABLE challenges DROP COLUMN IF EXISTS creator_id;
ALTER TABLE challenges ALTER COLUMN status SET DEFAULT 'upcoming';
