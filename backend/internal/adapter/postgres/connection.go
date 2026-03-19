package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnection(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	// Senior tip: Konfiguráljunk pool méretet és élettartamot a stabilitás érdekében a konfiguráció alapján.
	poolCfg.MaxConns = cfg.DBMaxConns
	poolCfg.MinConns = cfg.DBMinConns
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("error connecting to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging postgres: %w", err)
	}

	log.Printf("🐘 Connected to PostgreSQL (pool size: %d-%d)", cfg.DBMinConns, cfg.DBMaxConns)
	return pool, nil
}
