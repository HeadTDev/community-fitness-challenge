-- Migration: 000007_add_participant_count_to_challenges.down.sql
ALTER TABLE challenges DROP COLUMN IF EXISTS participant_count;
