---
name: healthkit-data-integrity-engineering
description: Principal HealthKit data integrity engineering. Use for permission handling, data provenance, and trustworthy metric ingestion.
---

# HealthKit Data Integrity Engineering

Ensure health metrics are collected, transformed, and submitted with integrity and auditability.

## Decision Criteria

- Define metric provenance and confidence rules before ingestion into scoring pipelines.
- Permission flow must be explicit, reversible, and transparent to users.
- Data synchronization must handle missing intervals, duplicate samples, and timezone boundaries.
- HealthKit integration choices must account for simulator mock parity and real-device behavior differences.

## Principal Practices

- Keep HealthKit access in isolated manager/actor with deterministic query windows.
- Normalize collected units and timestamps before persistence/submission.
- Track ingestion metadata (source, query window, sample count) for downstream anti-cheat and debugging.
- Maintain robust mock manager fixtures to reproduce edge cases in CI/dev.

## Failure Modes & Anti-Patterns

- Using local device timezone assumptions without backend-aligned daily boundary policy.
- Trusting single sample spikes without aggregation sanity checks.
- Permission-denied flows that break app navigation/state unexpectedly.
- Mock behavior diverging from production collection semantics.

## Project-Specific Examples

- Daily log generation must align HealthKit collection windows with backend one-log-per-day lock semantics.
- Suspicious cross-metric combinations should carry provenance metadata to anti-cheat adjudication pipeline.

## Related Skills

- `anti-cheat-detection-engineering`
- `scoring-ranking-engine-design`
- `cross-platform-system-design-authority`
