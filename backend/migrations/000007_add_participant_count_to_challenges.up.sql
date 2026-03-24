-- Migration: 000007_add_participant_count_to_challenges.up.sql
ALTER TABLE challenges ADD COLUMN IF NOT EXISTS participant_count INTEGER DEFAULT 0;
