package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LeaderboardRepo struct {
	db domain.DB
}

func NewLeaderboardRepo(pool *pgxpool.Pool) *LeaderboardRepo {
	return &LeaderboardRepo{db: pool}
}

func (r *LeaderboardRepo) UpdateScore(ctx context.Context, challengeID, userID uuid.UUID, score float64) error {
	query := `
		UPDATE participations
		SET current_score = $3, updated_at = NOW()
		WHERE challenge_id = $1 AND user_id = $2
	`
	tag, err := r.db.Exec(ctx, query, challengeID, userID, int(score))
	if err != nil {
		return fmt.Errorf("failed to update participation score: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LeaderboardRepo) GetRank(ctx context.Context, challengeID, userID uuid.UUID) (int, error) {
	query := `
		WITH ranked AS (
			SELECT user_id, ROW_NUMBER() OVER (ORDER BY current_score DESC, user_id ASC) AS rank
			FROM participations
			WHERE challenge_id = $1
		)
		SELECT rank FROM ranked WHERE user_id = $2
	`

	var rank int
	err := r.db.QueryRow(ctx, query, challengeID, userID).Scan(&rank)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrNotFound
		}
		return 0, fmt.Errorf("failed to query rank from postgres: %w", err)
	}
	return rank, nil
}

func (r *LeaderboardRepo) GetTopN(ctx context.Context, challengeID uuid.UUID, limit int64) ([]models.LeaderboardEntry, error) {
	if limit <= 0 {
		return []models.LeaderboardEntry{}, nil
	}

	query := `
		SELECT user_id, current_score, rank FROM (
			SELECT
				user_id,
				current_score,
				ROW_NUMBER() OVER (ORDER BY current_score DESC, user_id ASC) AS rank
			FROM participations
			WHERE challenge_id = $1
		) ranked
		ORDER BY rank
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, challengeID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top leaderboard from postgres: %w", err)
	}
	defer rows.Close()

	entries := make([]models.LeaderboardEntry, 0)
	for rows.Next() {
		var e models.LeaderboardEntry
		var score int
		if scanErr := rows.Scan(&e.UserID, &score, &e.Rank); scanErr != nil {
			return nil, fmt.Errorf("failed to scan leaderboard row: %w", scanErr)
		}
		e.Score = float64(score)
		entries = append(entries, e)
	}
	return entries, nil
}

func (r *LeaderboardRepo) GetAroundUser(ctx context.Context, challengeID, userID uuid.UUID, radius int64) ([]models.LeaderboardEntry, error) {
	if radius < 0 {
		radius = 0
	}

	query := `
		WITH ranked AS (
			SELECT
				user_id,
				current_score,
				ROW_NUMBER() OVER (ORDER BY current_score DESC, user_id ASC) AS rank
			FROM participations
			WHERE challenge_id = $1
		),
		me AS (
			SELECT rank FROM ranked WHERE user_id = $2
		)
		SELECT r.user_id, r.current_score, r.rank
		FROM ranked r, me
		WHERE r.rank BETWEEN me.rank - $3 AND me.rank + $3
		ORDER BY r.rank
	`

	rows, err := r.db.Query(ctx, query, challengeID, userID, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to query relative leaderboard from postgres: %w", err)
	}
	defer rows.Close()

	entries := make([]models.LeaderboardEntry, 0)
	for rows.Next() {
		var e models.LeaderboardEntry
		var score int
		if scanErr := rows.Scan(&e.UserID, &score, &e.Rank); scanErr != nil {
			return nil, fmt.Errorf("failed to scan relative leaderboard row: %w", scanErr)
		}
		e.Score = float64(score)
		entries = append(entries, e)
	}
	if len(entries) == 0 {
		return nil, domain.ErrNotFound
	}
	return entries, nil
}

func (r *LeaderboardRepo) GetTotalCount(ctx context.Context, challengeID uuid.UUID) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM participations WHERE challenge_id = $1`
	if err := r.db.QueryRow(ctx, query, challengeID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to query leaderboard total count: %w", err)
	}
	return count, nil
}
