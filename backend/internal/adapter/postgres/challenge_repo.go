package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChallengeRepo struct {
	pool *pgxpool.Pool
}

func NewChallengeRepo(pool *pgxpool.Pool) *ChallengeRepo {
	return &ChallengeRepo{pool: pool}
}

const challengeColumns = `id, creator_id, title, description, image_url, start_date, end_date, status, type, goal, created_at, updated_at, deleted_at`

func (r *ChallengeRepo) Create(ctx context.Context, c *models.Challenge) error {
	query := `
		INSERT INTO challenges (id, creator_id, title, description, image_url, start_date, end_date, status, type, goal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.pool.Exec(ctx, query,
		c.ID, c.CreatorID, c.Title, c.Description, c.ImageURL, c.StartDate, c.EndDate,
		c.Status, c.Type, c.Goal, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating challenge: %w", err)
	}
	return nil
}

func (r *ChallengeRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Challenge, error) {
	query := fmt.Sprintf("SELECT %s FROM challenges WHERE id = $1 AND deleted_at IS NULL", challengeColumns)
	return r.scanChallenge(r.pool.QueryRow(ctx, query, id))
}

func (r *ChallengeRepo) Update(ctx context.Context, c *models.Challenge) error {
	query := `
		UPDATE challenges
		SET creator_id = $2, title = $3, description = $4, image_url = $5, start_date = $6, end_date = $7, status = $8, type = $9, goal = $10, updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query,
		c.ID, c.CreatorID, c.Title, c.Description, c.ImageURL, c.StartDate, c.EndDate,
		c.Status, c.Type, c.Goal, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error updating challenge: %w", err)
	}
	return nil
}

func (r *ChallengeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE challenges SET deleted_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting challenge: %w", err)
	}
	return nil
}

func (r *ChallengeRepo) List(ctx context.Context, status *models.ChallengeStatus) ([]*models.Challenge, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(fmt.Sprintf("SELECT %s FROM challenges WHERE deleted_at IS NULL", challengeColumns))
	
	args := []interface{}{}
	if status != nil {
		queryBuilder.WriteString(" AND status = $1")
		args = append(args, *status)
	}
	
	queryBuilder.WriteString(" ORDER BY start_date ASC")

	rows, err := r.pool.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("error listing challenges: %w", err)
	}
	defer rows.Close()

	var challenges []*models.Challenge
	for rows.Next() {
		c, err := r.scanChallenge(rows)
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, c)
	}

	return challenges, nil
}

func (r *ChallengeRepo) scanChallenge(row pgx.Row) (*models.Challenge, error) {
	c := &models.Challenge{}
	err := row.Scan(
		&c.ID, &c.CreatorID, &c.Title, &c.Description, &c.ImageURL, &c.StartDate, &c.EndDate,
		&c.Status, &c.Type, &c.Goal, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error scanning challenge: %w", err)
	}
	return c, nil
}
