---
name: token-efficient-senior-developer
description: Principal delivery leadership for high-signal execution. Use to drive fast, correct outcomes with rigorous prioritization and minimal waste.
---

# Principal Token-Efficient Delivery Leadership

Optimize for outcome-per-token, outcome-per-hour, and outcome reliability across complex multi-team work.

## Decision Criteria

- Prioritize by user impact, risk reduction, and reversibility.
- Choose smallest change that fully solves root cause and preserves system coherence.
- Decide when to go broad (cross-cutting refactor) versus narrow (surgical patch) using dependency/risk surface.
- Require explicit verification evidence before declaring completion.

## Principal Practices

- Batch discovery and edits intelligently, but never skip dependency analysis for high-risk changes.
- Preserve repository conventions and shared abstractions to avoid local optimizations that create future drag.
- Track execution as staged milestones with clear done criteria.
- Escalate ambiguities early and convert them into explicit decisions with tradeoffs.

## Failure Modes & Anti-Patterns

- Token efficiency interpreted as shallow analysis.
- Over-optimization of local patch speed at cost of architectural consistency.
- Delivering partial fixes when requirement asks for end-to-end closure.
- Verbose process narration that hides technical signal.

## Project-Specific Examples

- When modifying challenge flows, account for API contract, service transaction logic, Redis counters, and verifier scripts in one coherent change.
- For roadmap features (worker, leaderboard, websocket), prefer shared patterns in `internal/app`, middleware, and response helpers over one-off implementations.

## Related Skills

- `engineering-program-orchestration`
- `technical-risk-tradeoff-management`
- `release-management-governance`
