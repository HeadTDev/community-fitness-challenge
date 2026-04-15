---
name: postmortem-continuous-improvement-leadership
description: Principal postmortem and continuous improvement leadership. Use to convert incidents and delivery misses into durable system improvement.
---

# Postmortem & Continuous Improvement Leadership

Turn failures into measurable reliability and delivery gains without blame theater.

## Decision Criteria

- Postmortem scope includes technical root cause, detection gaps, response quality, and process contributors.
- Actions must be concrete, owned, prioritized, and time-bounded.
- Improvement work is prioritized by recurrence risk and blast radius.
- Learning is only complete when preventive controls are verified in practice.

## Principal Practices

- Build timeline from observable evidence, not memory alone.
- Differentiate trigger, contributing conditions, and systemic weaknesses.
- Track action completion and effectiveness metrics in subsequent releases.
- Feed lessons into architecture decisions, test strategy, and runbooks.

## Failure Modes & Anti-Patterns

- Blame-focused analysis with no systems insight.
- Action lists without owners or verification criteria.
- Repeating incidents because mitigations were partial or never validated.
- Treating postmortems as documentation ritual instead of change mechanism.

## Project-Specific Examples

- Leaderboard drift incident should result in reconciliation automation, fallback monitoring, and release gate updates.
- Auth outage postmortem should produce stronger secret rotation process and token-lifecycle contract tests.

## Related Skills

- `incident-response-leadership`
- `technical-risk-tradeoff-management`
- `test-strategy-architecture`
