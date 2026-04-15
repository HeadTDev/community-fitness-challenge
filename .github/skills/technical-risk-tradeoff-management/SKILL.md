---
name: technical-risk-tradeoff-management
description: Principal technical risk and tradeoff management. Use for selecting options under uncertainty and controlling delivery risk.
---

# Technical Risk & Tradeoff Management

Turn uncertainty into explicit, testable decisions before it becomes production debt.

## Decision Criteria

- Classify risks by impact domain: data integrity, security/privacy, availability, cost, release confidence, and client compatibility.
- Choose options by measurable outcomes (p95 latency, error budget, replay safety, migration effort), not preference.
- Require kill-switch or rollback path for high-impact changes (auth, leaderboard consistency, websocket fanout, worker behavior).
- Force explicit ownership and verification commands for every accepted risk.

## Principal Practices

- Maintain a living risk register tied to roadmap milestones and critical flows.
- Use staged rollout plans for high-blast-radius work (new worker job types, scoring logic revisions, auth changes).
- Convert unknowns into bounded experiments with clear acceptance/rejection criteria.
- Escalate cross-domain risks early to architecture, security, and release governance skills.

## Failure Modes & Anti-Patterns

- Merging "temporary" bypasses without expiration policy.
- Treating LocalStack success as automatic production readiness.
- Shipping coupled changes (schema + API + worker + client) without coordinated rollback strategy.
- Accepting silent data drift between Redis and PostgreSQL as "eventual consistency."

## Project-Specific Examples

- Before enabling anti-cheat rejection in scoring, run shadow mode and compare leaderboard impact against baseline.
- Before introducing websocket-driven live rank updates, define behavior when Redis Pub/Sub drops or client reconnects after long offline periods.

## Related Skills

- `architecture-decision-governance`
- `release-management-governance`
- `incident-response-leadership`
