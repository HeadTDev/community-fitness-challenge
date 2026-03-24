-- Migration: 000008_add_max_participants_to_challenges.up.sql
ALTER TABLE challenges ADD COLUMN IF NOT EXISTS max_participants INTEGER DEFAULT 0;
