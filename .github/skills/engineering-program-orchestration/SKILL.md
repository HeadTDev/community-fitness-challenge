---
name: engineering-program-orchestration
description: Principal engineering program orchestration. Use for coordinating cross-domain delivery across backend, iOS, platform, and security tracks.
---

# Engineering Program Orchestration

Coordinate complex delivery streams so architecture intent survives execution reality.

## Decision Criteria

- Break work into milestones by dependency graph and risk, not arbitrary component boundaries.
- Sequence tasks to validate assumptions early (contracts before implementations, reliability before scale exposure).
- Define ownership and integration checkpoints for each cross-cutting feature.
- Program health must be measured through delivery risk indicators, not activity volume.

## Principal Practices

- Maintain a single dependency-aware roadmap linking API, worker, websocket, iOS, and infra milestones.
- Set entry/exit criteria for each milestone with verification commands and acceptance tests.
- Surface blockers with mitigation options and decision deadlines.
- Drive alignment between architectural decisions and day-to-day implementation plans.

## Failure Modes & Anti-Patterns

- Parallelizing tightly coupled work without contract lock.
- Milestones that "complete" without integration validation.
- Hidden dependency chains discovered during release week.
- Program tracking that ignores reliability/security readiness.

## Project-Specific Examples

- Sequence scoring/anti-cheat before final leaderboard trust metrics and real-time UX commitments.
- Coordinate Sign in with Apple, production config hardening, and GDPR workflows before v1 release tagging.

## Related Skills

- `architecture-decision-governance`
- `release-management-governance`
- `postmortem-continuous-improvement-leadership`
