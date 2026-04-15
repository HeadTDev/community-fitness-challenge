package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/HeadTDev/fitchallenge/internal/domain/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DailyLogRepo struct {
	db domain.DB
}

func NewDailyLogRepo(pool *pgxpool.Pool) *DailyLogRepo {
	return &DailyLogRepo{db: pool}
}

func (r *DailyLogRepo) WithTx(tx any) repositories.DailyLogRepository {
	return &DailyLogRepo{db: tx.(domain.DB)}
}

const dailyLogColumns = `id, user_id, challenge_id, log_date, steps, calories, active_minutes, score, healthkit_data_hash, source_bundle_ids, created_at, updated_at`

func (r *DailyLogRepo) Create(ctx context.Context, log *models.DailyLog) error {
	query := `
		INSERT INTO daily_logs (id, user_id, challenge_id, log_date, steps, calories, active_minutes, score, healthkit_data_hash, source_bundle_ids, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	logDate := normalizeDate(log.LogDate)
	_, err := r.db.Exec(ctx, query,
		log.ID, log.UserID, log.ChallengeID, logDate, log.Steps, log.Calories, log.ActiveMinutes, log.Score,
		log.HealthKitDataHash, log.SourceBundleIDs, log.CreatedAt, log.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return domain.ErrAlreadyExists
			case "23514":
				return domain.ErrInvalidInput
			}
		}
		return fmt.Errorf("error creating daily log: %w", err)
	}
	return nil
}

func (r *DailyLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.DailyLog, error) {
	query := fmt.Sprintf("SELECT %s FROM daily_logs WHERE id = $1", dailyLogColumns)
	return r.scanDailyLog(r.db.QueryRow(ctx, query, id))
}

func (r *DailyLogRepo) GetByUserChallengeDate(ctx context.Context, userID, challengeID uuid.UUID, logDate time.Time) (*models.DailyLog, error) {
	query := fmt.Sprintf("SELECT %s FROM daily_logs WHERE user_id = $1 AND challenge_id = $2 AND log_date = $3", dailyLogColumns)
	return r.scanDailyLog(r.db.QueryRow(ctx, query, userID, challengeID, normalizeDate(logDate)))
}

func (r *DailyLogRepo) ListByUserAndChallenge(ctx context.Context, userID, challengeID uuid.UUID) ([]*models.DailyLog, error) {
	query := fmt.Sprintf("SELECT %s FROM daily_logs WHERE user_id = $1 AND challenge_id = $2 ORDER BY log_date DESC", dailyLogColumns)
	rows, err := r.db.Query(ctx, query, userID, challengeID)
	if err != nil {
		return nil, fmt.Errorf("error listing daily logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.DailyLog
	for rows.Next() {
		log, scanErr := r.scanDailyLog(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (r *DailyLogRepo) scanDailyLog(row pgx.Row) (*models.DailyLog, error) {
	log := &models.DailyLog{}
	err := row.Scan(
		&log.ID,
		&log.UserID,
		&log.ChallengeID,
		&log.LogDate,
		&log.Steps,
		&log.Calories,
		&log.ActiveMinutes,
		&log.Score,
		&log.HealthKitDataHash,
		&log.SourceBundleIDs,
		&log.CreatedAt,
		&log.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error scanning daily log: %w", err)
	}
	return log, nil
}

func normalizeDate(t time.Time) time.Time {
	year, month, day := t.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
