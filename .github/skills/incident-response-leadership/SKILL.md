---
name: incident-response-leadership
description: Principal incident response leadership for production events. Use for detection, containment, recovery, and learning loops.
---

# Incident Response Leadership

Lead incidents with speed and rigor while preserving system integrity and organizational clarity.

## Decision Criteria

- Incident severity is based on user impact, data integrity risk, and security exposure.
- Response actions must prioritize containment and correctness over convenience.
- Communication cadence and ownership must be defined at incident start.
- Recovery is complete only when both service and data consistency are verified.

## Principal Practices

- Use explicit incident command roles (lead, communications, ops executor, domain owner).
- Apply runbooks for common scenarios (Redis outage, queue backlog, auth failure spike, data drift).
- Capture timeline with key decisions and evidence for postmortem quality.
- Convert incident learnings into preventive backlog with accountable owners.

## Failure Modes & Anti-Patterns

- Solving symptoms without confirming root cause.
- Silent "fixes" without stakeholder communication and impact assessment.
- Restoring traffic before data reconciliation completion.
- Postmortems that stop at human error labels.

## Project-Specific Examples

- During leaderboard inconsistency incident, first contain writes if needed, then reconcile Redis from PostgreSQL before re-enabling live rank updates.
- During JWT/auth outage, rotate/validate secrets safely and verify token refresh and middleware validation flows before declaring resolved.

## Related Skills

- `observability-engineering`
- `disaster-recovery-business-continuity`
- `postmortem-continuous-improvement-leadership`
