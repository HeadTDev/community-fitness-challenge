---
name: worker-reliability-engineering
description: Principal reliability engineering for background workers. Use for SQS polling loops, graceful shutdown, and retry-safe execution.
---

# Worker Reliability Engineering

Build workers that maintain correctness under retries, restarts, poison messages, and partial downstream failures.

## Decision Criteria

- Polling configuration must be selected from workload characteristics (message volume, SLA, cost sensitivity).
- Every handler path must be idempotent before enabling retries.
- Message delete is allowed only after side effects and durable status updates succeed.
- Graceful shutdown design must define in-flight work handling and visibility timeout interaction.

## Principal Practices

- Use long polling with bounded concurrency and explicit worker pool sizing.
- Persist processing checkpoints/idempotency markers in PostgreSQL or Redis with TTL policy.
- Route repeated failures to DLQ with actionable error classification.
- Emit structured worker logs and metrics per job type, success/failure reason, and processing latency.

## Failure Modes & Anti-Patterns

- Fire-and-forget processing without durable checkpointing.
- Unlimited parallelism causing DB pool starvation or Redis saturation.
- Logging only generic failure messages without payload correlation identifiers.
- Graceful shutdown that exits before visibility extensions or handler completion.

## Project-Specific Examples

- `cmd/worker/main.go` must finish currently handled SQS message before termination signal shutdown path exits.
- `gdpr_delete` and `send_email` handlers must be safe for duplicate delivery and partial downstream outage.

## Related Skills

- `event-driven-workflow-orchestration`
- `incident-response-leadership`
- `data-consistency-reconciliation-engineering`
