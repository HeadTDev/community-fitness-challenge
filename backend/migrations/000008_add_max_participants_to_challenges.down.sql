-- Migration: 000008_add_max_participants_to_challenges.down.sql
ALTER TABLE challenges DROP COLUMN IF EXISTS max_participants;
