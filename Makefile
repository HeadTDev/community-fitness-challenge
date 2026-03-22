# --- Community Fitness Challenge Makefile ---

.PHONY: dev stop logs-api aws-status verify db-migrate db-migrate-down db-shell db-tables restart-api test clean status help

# Start services in the background
dev:
	docker compose up -d --build

# Stop all services and remove containers
stop:
	docker compose down

# View real-time logs for the API service
logs-api:
	docker compose logs -f api

# Quick status check of LocalStack resources
aws-status:
	@echo "--- S3 Buckets ---"
	@docker compose exec -T localstack awslocal s3 ls
	@echo "\n--- SQS Queues ---"
	@docker compose exec -T localstack awslocal sqs list-queues
	@echo "\n--- Secrets ---"
	@docker compose exec -T localstack awslocal secretsmanager list-secrets

# Run the full infrastructure and API verification script
verify:
	chmod +x tests/verify_infra.sh
	./tests/verify_infra.sh

# Run all UP database migrations
db-migrate:
	docker compose run --rm migrate

# Rollback one database migration
db-migrate-down:
	docker compose run --rm migrate go run cmd/migrate/main.go -dir down

# Open a PSQL shell into the postgres container
db-shell:
	docker compose exec postgres psql -U fc_user -d fitchallenge

# List all tables in the public schema (Day 14 verification)
db-tables:
	@docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\dt"

# Run all Go tests inside the API container
test:
	docker compose exec -T api go test ./... -v

# Clean up build artifacts and temporary files
clean:
	rm -rf backend/tmp/
	docker compose down -v --remove-orphans

# Show status of all running containers
status:
	docker compose ps

# Help command to list available targets
help:
	@echo "Available commands:"
	@echo "  dev             - Start all services (detached)"
	@echo "  stop            - Stop and remove containers"
	@echo "  logs-api        - Follow API container logs"
	@echo "  aws-status      - List simulated AWS resources"
	@echo "  verify          - Run Day-to-Day verification script"
	@echo "  db-migrate      - Apply pending migrations"
	@echo "  db-migrate-down - Rollback last migration"
	@echo "  db-tables       - List all database tables"
	@echo "  db-shell        - Enter PostgreSQL shell"
	@echo "  test            - Run backend Go tests"
	@echo "  restart-api     - Restart the API service"
	@echo "  clean           - Wipe volumes and build artifacts"
	@echo "  status          - Show docker container status"
