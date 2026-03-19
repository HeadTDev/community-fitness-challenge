---
name: fc-backend-architect
description: Senior Go architect for the Community Fitness Challenge. Use for implementing clean architecture, Go/Gin API endpoints, PostgreSQL migrations, and Redis integration.
---

# Community Fitness Challenge - Go Backend Architecture

This skill provides the architectural guardrails for the Go 1.26 backend. Adhere to these senior-level standards to ensure a scalable, maintainable, and professional codebase.

## 🏗️ Clean Architecture Standards

- **Entry Points (`cmd/`)**: Keep `main.go` minimal. Initialize config, DB, Redis, AWS SDK, and then start the server.
- **Internal Logic (`internal/`)**:
  - `domain/models/`: Pure Go structs (User, Challenge, Log).
  - `domain/services/`: Business logic (Scoring, Anti-cheat).
  - `adapter/postgres/`: SQL implementations of domain repositories.
  - `handler/http/`: Gin routes and handlers. Use standard response formats.
- **Migrations (`migrations/`)**: Use raw SQL files (`.up.sql`, `.down.sql`). All changes must be reversible.

## 📡 API & Middleware Standards

- **Standard Response**: Use `internal/pkg/response/json.go`.
  - Success: `{"success": true, "data": {...}, "meta": {"timestamp": "..."}}`
  - Error: `{"success": false, "error": {"code": "...", "message": "..."}, "meta": {"timestamp": "..."}}`
- **Auth**: JWT-based via `Authorization: Bearer <token>`. Access and Refresh tokens required.
- **Middleware**: Mandatory Logging, Recovery, CORS, and Rate Limiting (Redis-based).

## 🗄️ Database & Cache Patterns

- **PostgreSQL**: Use `pgx/v5` for connections. Prefer prepared statements for security.
- **Redis**: 
  - **Leaderboards**: Use Sorted Sets (`ZSET`) for real-time ranking.
  - **Locking**: Use `SET NX` for preventing double-logging in a single day.
  - **Counter**: Atomic `INCR/DECR` for challenge participant counts.

## 🧪 Testing Protocol

- **Unit Tests**: Mandatory for all `domain/services` logic (e.g., scoring).
- **Integration Tests**: Required for repositories using a test database.
- **Mocking**: Use interfaces to allow mocking of DB and AWS clients.
