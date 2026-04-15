# Copilot Instructions for community-fitness-challenge

## Build, test, and lint commands
- Start the full local stack (rebuilds images): `make dev`
- Stop containers: `make stop`
- Run backend Go tests (full suite in `api` container): `make test`
- Run a single package: `docker compose exec -T api go test ./internal/domain/services -v`
- Run a single test: `docker compose exec -T api go test ./internal/domain/services -run '^TestJoinChallenge$' -v`
- Run infra/API verification script: `make verify`
- Apply migrations: `make db-migrate`
- Seed database: `make seed`
- **Lint:** no dedicated lint target is currently defined in active project files.

## High-level architecture
- `backend/cmd/api/main.go` is the API entrypoint: it loads env config, sets JSON `slog`, creates the app graph, and handles graceful shutdown.
- `backend/internal/app/app.go` is the dependency wiring root. It initializes PostgreSQL, Redis, AWS config/S3 client, JWT manager, repositories, and services, then registers Gin middleware/routes.
- Request flow is layered: `handler/http` (transport + validation) -> `domain/services` (business rules/transactions) -> `adapter/postgres` and `adapter/redis` (persistence/cache), with `internal/aws` for cloud clients.
- `docker-compose.yml` defines runtime orchestration: `api` depends on healthy `postgres`, `redis`, `localstack`, and a one-shot `migrate` service; `verifier` runs end-to-end checks from `tests/verify_infra.sh`.

## Key conventions in this codebase
- Return API responses through `internal/pkg/response` helpers so payloads keep the shared envelope (`success`, `data`/`error`, `meta.timestamp`, `meta.request_id`).
- Middleware order in `app.NewRouter` is intentional: Request ID -> structured logger -> global Redis rate limit -> recovery. Protected APIs are grouped under `/v1` with JWT auth middleware.
- Use context keys from `internal/handler/http/middleware/context.go` (`user_id`, `user_role`, `X-Request-ID`) when passing request metadata.
- Repositories map storage errors to domain sentinels (`internal/domain/errors.go`), and handlers map those with `errors.Is(...)` to HTTP responses.
- PostgreSQL repositories implement soft delete via `deleted_at`; reads consistently filter `deleted_at IS NULL`.
- Challenge participation logic keeps DB and Redis counters aligned: join/leave updates run in DB transactions first, then adjust Redis `challenge_count:%s`.
- Integration tests are environment-gated (for example `DB_HOST` and `AWS_ENDPOINT_URL`) and are meant to run against the Docker/LocalStack stack.
- `/auth/register-dev` is a development-only endpoint and is blocked when `APP_ENV=production`.

## Integrated principal skill system
- Full skill index is maintained in `.github/skills-registry.md`.
- Skill definitions for Copilot use live in `.github/skills/<skill-name>/SKILL.md`.
- Source mirror remains in `.gemini/skills/<skill-name>/SKILL.md`.
- For each non-trivial task, select:
  1. one primary skill,
  2. one reliability/security companion skill,
  3. one delivery/governance skill.
- Always combine architecture + execution skills for cross-cutting changes (API + worker + Redis + iOS).
- For changes touching scoring/leaderboard/real-time paths, always include consistency and observability skills from the registry.
- For production-facing changes, always include security hardening and release governance skills from the registry.
