---
name: chaos-resilience-engineering
description: Principal chaos and resilience engineering. Use for controlled fault injection and confidence in degraded operation modes.
---

# Chaos & Resilience Engineering

Prove the system fails gracefully and recovers predictably before real outages do it for you.

## Decision Criteria

- Chaos scenarios must target known single points of failure and hidden coupling paths.
- Experiments need measurable hypotheses and rollback criteria.
- Resilience readiness requires both service continuity and data integrity outcomes.
- Chaos tests should run in production-like environments with safe blast radius controls.

## Principal Practices

- Inject failures for Redis downtime, worker crash loops, queue backlog, and partial AWS service outages.
- Verify fallback paths, replay safety, and recovery automation after each experiment.
- Include client-impact assertions (stale leaderboard behavior, delayed notification tolerance).
- Maintain experiment catalogue mapped to incident history and roadmap risks.

## Failure Modes & Anti-Patterns

- One-time chaos demo with no recurring cadence.
- Running fault injection without observability instrumentation.
- Measuring only uptime while ignoring correctness drift.
- Resilience assumptions untested after architecture changes.

## Project-Specific Examples

- Stop Redis during leaderboard traffic and validate PostgreSQL fallback + post-recovery rebuild behavior.
- Restart worker during `log_submitted` burst and verify no duplicate score effect after message retries.

## Related Skills

- `data-consistency-reconciliation-engineering`
- `disaster-recovery-business-continuity`
- `incident-response-leadership`
