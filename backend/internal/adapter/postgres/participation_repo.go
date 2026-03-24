package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/HeadTDev/fitchallenge/internal/domain/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	)

type ParticipationRepo struct {
	db domain.DB
}

func NewParticipationRepo(pool *pgxpool.Pool) *ParticipationRepo {
	return &ParticipationRepo{db: pool}
}

func (r *ParticipationRepo) WithTx(tx any) repositories.ParticipationRepository {
	return &ParticipationRepo{db: tx.(domain.DB)}
}

const participationColumns = `id, user_id, challenge_id, current_score, rank, joined_at, updated_at`

func (r *ParticipationRepo) Add(ctx context.Context, p *models.Participation) error {
	query := `
		INSERT INTO participations (id, user_id, challenge_id, current_score, rank, joined_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.UserID, p.ChallengeID, p.CurrentScore, p.Rank, p.JoinedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error adding participation: %w", err)
	}
	return nil
}

func (r *ParticipationRepo) Remove(ctx context.Context, userID, challengeID uuid.UUID) error {
	query := `DELETE FROM participations WHERE user_id = $1 AND challenge_id = $2`
	_, err := r.db.Exec(ctx, query, userID, challengeID)
	if err != nil {
		return fmt.Errorf("error removing participation: %w", err)
	}
	return nil
}

func (r *ParticipationRepo) Get(ctx context.Context, userID, challengeID uuid.UUID) (*models.Participation, error) {
	query := fmt.Sprintf("SELECT %s FROM participations WHERE user_id = $1 AND challenge_id = $2", participationColumns)
	p := &models.Participation{}
	err := r.db.QueryRow(ctx, query, userID, challengeID).Scan(
		&p.ID, &p.UserID, &p.ChallengeID, &p.CurrentScore, &p.Rank, &p.JoinedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error getting participation: %w", err)
	}
	return p, nil
}

func (r *ParticipationRepo) GetParticipantsCount(ctx context.Context, challengeID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM participations WHERE challenge_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, challengeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting participants count: %w", err)
	}
	return count, nil
}

func (r *ParticipationRepo) ListByChallenge(ctx context.Context, challengeID uuid.UUID) ([]*models.Participation, error) {
	query := fmt.Sprintf("SELECT %s FROM participations WHERE challenge_id = $1 ORDER BY current_score DESC", participationColumns)
	rows, err := r.db.Query(ctx, query, challengeID)
	if err != nil {
		return nil, fmt.Errorf("error listing participations: %w", err)
	}
	defer rows.Close()

	var participations []*models.Participation
	for rows.Next() {
		p := &models.Participation{}
		err := rows.Scan(
			&p.ID, &p.UserID, &p.ChallengeID, &p.CurrentScore, &p.Rank, &p.JoinedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning participation: %w", err)
		}
		participations = append(participations, p)
	}
	return participations, nil
}
