---
name: integration-contract-testing-engineering
description: Principal integration and contract testing engineering. Use for API envelope stability, queue schema safety, and adapter correctness.
---

# Integration & Contract Testing Engineering

Guarantee that component boundaries remain compatible as the system evolves.

## Decision Criteria

- Contract tests are required for all externally consumed boundaries: REST DTOs, websocket events, queue payloads.
- Integration scope must include actual adapters and middleware behaviors.
- Backward compatibility windows must be test-encoded for versioned contracts.
- High-change interfaces require schema snapshot and diff enforcement.

## Principal Practices

- Assert response envelope invariants for all public endpoints.
- Validate domain error mapping to stable API error codes.
- Create queue payload schema tests for each job type and version.
- Include fallback-mode contract tests (Redis unavailable, queue lag, S3 temporary failure).

## Failure Modes & Anti-Patterns

- Handler tests that bypass middleware and context setup.
- Queue consumers accepting malformed payloads silently.
- Endpoint tests that verify status code only, not payload contract.
- Missing contract checks after model refactors.

## Project-Specific Examples

- Test `POST /v1/challenges/:id/join` conflict/full conditions against expected codes and messages used by iOS.
- Validate websocket `rank_update` payload schema with sequence and ranking fields expected by leaderboard UI.

## Related Skills

- `api-contract-governance`
- `event-driven-workflow-orchestration`
- `cross-platform-system-design-authority`
