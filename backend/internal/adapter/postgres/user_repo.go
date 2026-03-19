package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, apple_id, email, display_name, avatar_url, timezone, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.AppleID, user.Email, user.DisplayName, user.AvatarURL,
		user.Timezone, user.Role, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, apple_id, email, display_name, avatar_url, timezone, role, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.AppleID, &user.Email, &user.DisplayName, &user.AvatarURL,
		&user.Timezone, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByAppleID(ctx context.Context, appleID string) (*models.User, error) {
	query := `
		SELECT id, apple_id, email, display_name, avatar_url, timezone, role, created_at, updated_at, deleted_at
		FROM users
		WHERE apple_id = $1 AND deleted_at IS NULL
	`
	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, appleID).Scan(
		&user.ID, &user.AppleID, &user.Email, &user.DisplayName, &user.AvatarURL,
		&user.Timezone, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by apple_id: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, apple_id, email, display_name, avatar_url, timezone, role, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`
	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.AppleID, &user.Email, &user.DisplayName, &user.AvatarURL,
		&user.Timezone, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET apple_id = $2, email = $3, display_name = $4, avatar_url = $5, timezone = $6, role = $7, updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.AppleID, user.Email, user.DisplayName, user.AvatarURL,
		user.Timezone, user.Role, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	return nil
}
