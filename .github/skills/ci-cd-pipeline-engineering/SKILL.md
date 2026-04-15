---
name: ci-cd-pipeline-engineering
description: Principal CI/CD pipeline engineering for this stack. Use for deterministic verification, release gating, and rollback-safe deployments.
---

# CI/CD Pipeline Engineering

Build pipelines that prove system behavior end-to-end, not just compile success.

## Decision Criteria

- Every pipeline stage must map to a concrete failure class: correctness, security, compatibility, performance, release safety.
- Release gates are based on measurable checks (tests, coverage targets, infra verification, migration checks), not manual confidence.
- Pipeline changes must preserve reproducibility across local, CI, and release environments.
- Rollback path must be validated before rollout automation is approved.

## Principal Practices

- Run backend unit/integration suite (`go test ./... -v -race -cover`) and containerized verification (`make verify`) in CI.
- Include schema migration forward/backward checks and seed safety checks for critical environments.
- Add contract checks for API envelope and key endpoint behaviors used by iOS.
- Require artifact traceability: commit SHA, migration version, and environment promotion metadata.

## Failure Modes & Anti-Patterns

- Green pipelines that skip integration dependencies (PostgreSQL/Redis/LocalStack).
- Long-running flaky checks with no quarantine and no reliability owner.
- Deploying build artifacts that were not validated against migration state.
- Manual production deploy steps without auditable approval trail.

## Project-Specific Examples

- Pipeline should validate `make test` plus at least one containerized end-to-end flow matching verifier expectations.
- Release candidate should prove queue-driven worker path and leaderboard fallback behavior before tag promotion.

## Related Skills

- `test-strategy-architecture`
- `end-to-end-release-validation-engineering`
- `release-management-governance`
