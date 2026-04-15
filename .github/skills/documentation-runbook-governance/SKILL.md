---
name: documentation-runbook-governance
description: Principal documentation and runbook governance. Use for operational clarity, onboarding speed, and incident-time execution quality.
---

# Documentation & Runbook Governance

Treat documentation as executable operational infrastructure.

## Decision Criteria

- Documentation must answer operational questions under time pressure, not only explain concepts.
- Runbooks are required for high-frequency and high-impact tasks.
- Docs are accepted only when validated against current commands and environment behavior.
- Ownership and update triggers must be defined for each critical document.

## Principal Practices

- Keep command-level docs aligned with active Makefile/compose/runtime behavior.
- Maintain runbooks for migrations, queue replay, leaderboard rebuild, and environment recovery.
- Link architecture decisions to implementation docs and verification commands.
- Include failure signatures and triage paths in operational runbooks.

## Failure Modes & Anti-Patterns

- Docs that describe intended behavior but not actual runtime behavior.
- Runbooks lacking preconditions or rollback steps.
- Documentation scattered without canonical source hierarchy.
- Release changes merged without doc impact update.

## Project-Specific Examples

- Verifier script expectations should be mirrored in runbooks for manual diagnosis when `make verify` fails.
- Local-to-production transition docs must explicitly cover environment variable differences and security implications.

## Related Skills

- `engineering-program-orchestration`
- `release-management-governance`
- `incident-response-leadership`
