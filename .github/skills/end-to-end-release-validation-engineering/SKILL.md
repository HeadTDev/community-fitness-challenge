---
name: end-to-end-release-validation-engineering
description: Principal end-to-end release validation engineering. Use for proving complete user flows before release promotion.
---

# End-to-End Release Validation Engineering

Validate production-critical journeys across API, workers, cache layers, and client expectations.

## Decision Criteria

- Release candidates must pass full-path scenarios representing real user value and highest-risk flows.
- E2E scope must include asynchronous eventual-consistency windows.
- Validation suites require deterministic setup and teardown for repeatability.
- Any skipped E2E check must include signed risk acceptance.

## Principal Practices

- Encode canonical release journey: auth -> challenge lifecycle -> daily log -> scoring -> leaderboard -> notification -> profile asset behavior.
- Validate both normal and degraded modes (Redis outage fallback, delayed worker processing).
- Include production-profile checks for env hardening assumptions.
- Tie E2E results directly to release gate status.

## Failure Modes & Anti-Patterns

- "Smoke test only" releases for high-blast-radius changes.
- Green E2E runs based on stale fixture state.
- Ignoring client-visible delay windows for async updates.
- Manual untracked E2E checks with no reproducible evidence.

## Project-Specific Examples

- Confirm rank updates propagate from log submission to websocket event and iOS-visible ordering behavior.
- Validate GDPR deletion journey removes profile visibility and associated asset references end-to-end.

## Related Skills

- `release-management-governance`
- `test-strategy-architecture`
- `incident-response-leadership`
