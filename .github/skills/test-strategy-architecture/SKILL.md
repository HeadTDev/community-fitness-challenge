---
name: test-strategy-architecture
description: Principal test strategy architecture for backend, async workers, and client-facing contracts.
---

# Test Strategy Architecture

Design test systems that expose correctness risk early across synchronous and asynchronous paths.

## Decision Criteria

- Test pyramid distribution must reflect risk, not tradition.
- Critical user journeys require coverage across unit, integration, and end-to-end layers.
- Asynchronous pipelines must be tested for retries, duplication, and ordering variance.
- Strategy must define deterministic fixtures for scoring, ranking, and anti-cheat behavior.

## Principal Practices

- Keep fast unit suites for services/domain logic and contract-style integration suites for adapters and handlers.
- Run integration tests against real dependencies (PostgreSQL, Redis, LocalStack) in controlled environments.
- Add replay/idempotency tests for worker job handlers.
- Track coverage by risk area (auth, scoring, leaderboard, GDPR deletion) rather than raw percentage only.

## Failure Modes & Anti-Patterns

- Chasing coverage numbers while missing high-risk behavioral paths.
- Integration tests that mock away external behavior they intend to validate.
- Non-deterministic tests relying on current time without control points.
- Missing regression suite for known incident classes.

## Project-Specific Examples

- Ensure journey test exists for create challenge -> join -> log -> rank update -> leave under both Redis-healthy and Redis-fallback modes.
- Verify daily lock behavior with duplicate same-day submissions and queue side-effect assertions.

## Related Skills

- `integration-contract-testing-engineering`
- `end-to-end-release-validation-engineering`
- `ci-cd-pipeline-engineering`
