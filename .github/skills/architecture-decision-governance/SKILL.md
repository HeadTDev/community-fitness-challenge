---
name: architecture-decision-governance
description: Principal architecture decision governance for this stack. Use for ADR quality, irreversible decision control, and system-wide consistency.
---

# Architecture Decision Governance

Drive architecture decisions with explicit tradeoffs, reversibility analysis, and measurable acceptance criteria.

## Decision Criteria

- Require ADRs for irreversible choices: persistence model changes, event schema shape, auth model changes, websocket protocol changes, and deployment topology changes.
- Evaluate each option against this project's constraints: Go + Gin backend, PostgreSQL as source of truth, Redis acceleration, LocalStack-to-AWS parity, iOS offline behavior.
- Define rollback and migration strategy before approving any decision that touches data shape, API contracts, or message formats.
- Reject decisions without observability impact definitions (logs, metrics, alerts) and verification commands.

## Principal Practices

- Use one canonical ADR structure: context, options, tradeoffs, decision, rollout plan, rollback trigger, validation checklist.
- Tie architecture decisions directly to Dev Plan milestones so sequence dependencies are explicit (for example scoring before leaderboard optimization).
- Enforce versioning strategy for APIs/events before implementation begins.
- Require cross-skill review for coupled changes (architecture + security + reliability + mobile client impact).

## Failure Modes & Anti-Patterns

- "Implement first, document later" decisions create incompatible behavior across API, worker, and iOS clients.
- Optimizing for local convenience (for example endpoint shortcuts) without production parity guidance.
- Approving decisions that depend on hidden assumptions about Redis availability or eventual consistency without fallback path.
- Architecture notes that omit concrete verification commands.

## Project-Specific Examples

- Redis leaderboard as primary read path must include PostgreSQL fallback query path and rebuild command definition before merge.
- Introducing websocket rank updates requires event contract, reconnection semantics, and iOS rendering fallback in the same decision packet.

## Related Skills

- `go-hexagonal-architect`
- `technical-risk-tradeoff-management`
- `release-management-governance`
