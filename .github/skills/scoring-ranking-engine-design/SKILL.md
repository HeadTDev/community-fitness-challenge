---
name: scoring-ranking-engine-design
description: Principal scoring and ranking engine design. Use for metric normalization, deterministic scoring, and fairness under edge cases.
---

# Scoring & Ranking Engine Design

Design scoring that is mathematically stable, auditable, and robust against adversarial or noisy input data.

## Decision Criteria

- Define normalization windows and weight coefficients as versioned configuration, not hard-coded magic values.
- Ensure score computation is deterministic for identical inputs across API, worker, and rebuild jobs.
- Separate eligibility checks (valid/suspicious/rejected) from score calculation.
- Require explainability output for each score component to support disputes and audits.

## Principal Practices

- Use decimal-safe arithmetic strategy and explicit rounding rules before persistence.
- Version score formulas and keep migration strategy for historical comparability.
- Implement unit/property tests around boundary values and cross-metric outliers.
- Keep leaderboard ranking stable under tie conditions with deterministic secondary sort keys.

## Failure Modes & Anti-Patterns

- Formula changes without versioning, causing silent historical drift.
- Mixing anti-cheat suspicion penalties directly into raw metric normalization without traceability.
- Non-deterministic sorting leading to rank flicker in websocket updates.
- Accepting impossible metric combinations without flagging.

## Project-Specific Examples

- Planned reference case (`650 kcal`, `12000 steps`, `45 active minutes`) must remain stable under formula v1 test fixtures.
- Rejected daily logs must contribute zero rank impact while still persisting review evidence.

## Related Skills

- `anti-cheat-detection-engineering`
- `data-consistency-reconciliation-engineering`
- `integration-contract-testing-engineering`
