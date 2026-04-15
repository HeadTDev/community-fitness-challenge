---
name: release-management-governance
description: Principal release management governance. Use for readiness gates, rollout policy, and rollback-safe production changes.
---

# Release Management Governance

Ship confidently by enforcing objective release readiness and controlled rollout behavior.

## Decision Criteria

- Release approval requires green checks across correctness, reliability, security, and operational readiness.
- Rollout strategy must match blast radius and reversibility profile.
- Every release needs explicit rollback trigger conditions and owner responsibilities.
- Documentation and runbook readiness is part of release quality, not post-release cleanup.

## Principal Practices

- Maintain release checklist covering migrations, worker compatibility, API contracts, and client impacts.
- Use staged promotion with post-stage validation before full rollout.
- Require release notes with user-visible changes, operational risks, and rollback steps.
- Keep release artifacts tied to immutable build/test evidence.

## Failure Modes & Anti-Patterns

- Releasing cross-service changes without compatibility matrix.
- "Hotfix in prod" without retrospective hardening plan.
- Tagging releases before E2E async flows are validated.
- Rollback scripts untested against current schema/data conditions.

## Project-Specific Examples

- Before v1 tagging, validate full journey including queue-driven scoring updates and websocket rank refresh behavior.
- Production cutover must verify LocalStack-only assumptions are removed and AWS integration paths are correct.

## Related Skills

- `ci-cd-pipeline-engineering`
- `end-to-end-release-validation-engineering`
- `incident-response-leadership`
