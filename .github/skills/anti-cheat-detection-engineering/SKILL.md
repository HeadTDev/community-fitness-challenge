---
name: anti-cheat-detection-engineering
description: Principal anti-cheat detection engineering for health and activity logs. Use for anomaly detection, adjudication policy, and score integrity.
---

# Anti-Cheat Detection Engineering

Protect competitive integrity with explainable, replayable, and fairness-aware detection pipelines.

## Decision Criteria

- Detection policy must distinguish impossible, improbable, and normal behavior with transparent thresholds.
- Anti-cheat outcomes must be explainable and reversible with audit trail.
- Detection latency targets must balance user feedback speed against false-positive risk.
- Scoring impact rules must be explicit per adjudication state (`valid`, `suspicious`, `rejected`).

## Principal Practices

- Apply multi-signal checks: absolute bounds, cross-metric consistency, temporal plausibility, and user historical baseline deviation.
- Run adjudication asynchronously via worker pipeline; avoid blocking API writes with heavy analysis.
- Keep feature/threshold versioning and preserve decision metadata for reprocessing.
- Build reviewer tooling interfaces (or logs) that support override workflows without data corruption.

## Failure Modes & Anti-Patterns

- Permanent score penalties from opaque heuristics without evidence trail.
- High-volume false positives from static thresholds ignoring user profile context.
- Hard-coding detection constants with no change control.
- Coupling anti-cheat logic directly into transport handlers.

## Project-Specific Examples

- Planned outlier case (`14000 kcal`, `100 steps`, `5 min`) must classify as rejectable with explicit rule trace.
- Rejected logs should remain queryable for audit while excluded from leaderboard/scoring projections.

## Related Skills

- `scoring-ranking-engine-design`
- `worker-reliability-engineering`
- `system-integrity-guardian`
