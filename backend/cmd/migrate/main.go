package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	direction := flag.String("dir", "up", "Migration direction: up or down")
	flag.Parse()

	cfg := config.LoadConfig()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("❌ Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Ensure migrations table exists
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("❌ Failed to create schema_migrations table: %v", err)
	}

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		log.Fatalf("❌ Failed to find migration files: %v", err)
	}

	if *direction == "up" {
		runUp(ctx, conn, files)
	} else {
		runDown(ctx, conn, files)
	}
}

func runUp(ctx context.Context, conn *pgx.Conn, files []string) {
	sort.Strings(files)
	for _, file := range files {
		if !strings.HasSuffix(file, ".up.sql") {
			continue
		}

		version := getVersion(file)
		var exists bool
		err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", version).Scan(&exists)
		if err != nil {
			log.Fatalf("❌ Failed to check migration status: %v", err)
		}

		if exists {
			continue
		}

		fmt.Printf("🚀 Applying migration: %s\n", filepath.Base(file))
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("❌ Failed to read migration file: %v", err)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			log.Fatalf("❌ Failed to start transaction: %v", err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			log.Fatalf("❌ Migration failed for %s: %v", file, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback(ctx)
			log.Fatalf("❌ Failed to record migration: %v", err)
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("❌ Failed to commit migration: %v", err)
		}
	}
	fmt.Println("✅ All migrations applied successfully.")
}

func runDown(ctx context.Context, conn *pgx.Conn, files []string) {
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	for _, file := range files {
		if !strings.HasSuffix(file, ".down.sql") {
			continue
		}

		version := getVersion(file)
		var exists bool
		err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", version).Scan(&exists)
		if err != nil {
			log.Fatalf("❌ Failed to check migration status: %v", err)
		}

		if !exists {
			continue
		}

		fmt.Printf("⏪ Rolling back migration: %s\n", filepath.Base(file))
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("❌ Failed to read migration file: %v", err)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			log.Fatalf("❌ Failed to start transaction: %v", err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			log.Fatalf("❌ Rollback failed for %s: %v", file, err)
		}

		if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
			tx.Rollback(ctx)
			log.Fatalf("❌ Failed to remove migration record: %v", err)
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("❌ Failed to commit rollback: %v", err)
		}
		// Roll back only one for safety when running down manually
		break
	}
	fmt.Println("✅ Rollback complete.")
}

func getVersion(filename string) int64 {
	base := filepath.Base(filename)
	parts := strings.Split(base, "_")
	var version int64
	fmt.Sscanf(parts[0], "%d", &version)
	return version
}
