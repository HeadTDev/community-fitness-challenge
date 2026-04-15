---
name: domain-modeling-stewardship
description: Principal stewardship for domain models and invariants. Use when defining entities, state transitions, and aggregate boundaries.
---

# Domain Modeling Stewardship

Protect domain integrity by making business invariants explicit, testable, and enforceable across handlers, services, and repositories.

## Decision Criteria

- Model each aggregate around transactional invariants, not table convenience.
- Introduce or change domain fields only with explicit lifecycle ownership (who writes, who reads, when it is authoritative).
- Distinguish computed values (for example leaderboard score) from persisted source values (daily logs, participation records).
- Any state machine change must include allowed transitions and forbidden transitions.

## Principal Practices

- Keep `Challenge`, `Participation`, `Prize`, `User`, and upcoming `DailyLog` invariants centralized in service/domain logic.
- Encode role-based permissions using domain roles (`participant`, `creator`, `admin`) and enforce once in service boundaries.
- Preserve monotonic safety rules (for example participant count synchronization, one-log-per-day constraints, score inclusion eligibility).
- Define event payloads from domain semantics, not transport layer shapes.

## Failure Modes & Anti-Patterns

- Leaking persistence concerns into model semantics (`NULL` handling dictating business rules).
- Duplicating invariant checks across handler and repository code paths.
- Allowing "draft-only" operations (prize modifications, publish transitions) to bypass status rules.
- Mixing anti-cheat outputs with source metrics, making auditability impossible.

## Project-Specific Examples

- Daily log domain should separate raw health inputs from normalized scoring outputs to support anti-cheat re-evaluation.
- Notification domain should persist delivery intent and delivery state separately so SQS retries are idempotent.

## Related Skills

- `go-hexagonal-architect`
- `scoring-ranking-engine-design`
- `anti-cheat-detection-engineering`
