# Build and run with Docker Compose
dev:
	docker compose up -d --build

# Stop the services
stop:
	docker compose down

# View logs for the API
logs-api:
	docker compose logs -f api

# Check AWS resources status in LocalStack
aws-status:
	docker compose exec localstack awslocal s3 ls
	docker compose exec localstack awslocal sqs list-queues
	docker compose exec localstack awslocal secretsmanager list-secrets

# Run infrastructure verification tests
verify:
	./tests/verify_infra.sh

# DB migrations
db-migrate:
	docker compose run --rm migrate

db-migrate-down:
	docker compose run --rm migrate go run cmd/migrate/main.go -dir down

# Shell into the database
db-shell:
	docker compose exec postgres psql -U fc_user -d fitchallenge

# Rebuild and restart only the API
restart-api:
	docker compose restart api
