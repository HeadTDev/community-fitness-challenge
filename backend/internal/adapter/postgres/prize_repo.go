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

type PrizeRepo struct {
	pool *pgxpool.Pool
}

func NewPrizeRepo(pool *pgxpool.Pool) *PrizeRepo {
	return &PrizeRepo{pool: pool}
}

const prizeColumns = `id, challenge_id, title, description, image_url, rank_required, created_at, updated_at`

func (r *PrizeRepo) Create(ctx context.Context, prize *models.Prize) error {
	query := `
		INSERT INTO prizes (id, challenge_id, title, description, image_url, rank_required, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		prize.ID, prize.ChallengeID, prize.Title, prize.Description, prize.ImageURL, prize.RankRequired,
		prize.CreatedAt, prize.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating prize: %w", err)
	}
	return nil
}

func (r *PrizeRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Prize, error) {
	query := fmt.Sprintf("SELECT %s FROM prizes WHERE id = $1", prizeColumns)
	return r.scanPrize(r.pool.QueryRow(ctx, query, id))
}

func (r *PrizeRepo) GetByChallengeID(ctx context.Context, challengeID uuid.UUID) ([]*models.Prize, error) {
	query := fmt.Sprintf("SELECT %s FROM prizes WHERE challenge_id = $1 ORDER BY rank_required ASC", prizeColumns)
	rows, err := r.pool.Query(ctx, query, challengeID)
	if err != nil {
		return nil, fmt.Errorf("error querying prizes by challenge id: %w", err)
	}
	defer rows.Close()

	var prizes []*models.Prize
	for rows.Next() {
		prize, err := r.scanPrize(rows)
		if err != nil {
			return nil, err
		}
		prizes = append(prizes, prize)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during prizes rows iteration: %w", err)
	}

	return prizes, nil
}

func (r *PrizeRepo) scanPrize(row pgx.Row) (*models.Prize, error) {
	prize := &models.Prize{}
	err := row.Scan(
		&prize.ID, &prize.ChallengeID, &prize.Title, &prize.Description, &prize.ImageURL, &prize.RankRequired,
		&prize.CreatedAt, &prize.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error scanning prize: %w", err)
	}
	return prize, nil
}

func (r *PrizeRepo) Update(ctx context.Context, prize *models.Prize) error {
	query := `
		UPDATE prizes
		SET title = $2, description = $3, image_url = $4, rank_required = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		prize.ID, prize.Title, prize.Description, prize.ImageURL, prize.RankRequired, prize.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error updating prize: %w", err)
	}
	return nil
}

func (r *PrizeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM prizes WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting prize: %w", err)
	}
	return nil
}
