---
name: event-driven-workflow-orchestration
description: Principal event-driven orchestration for SQS-based workflows. Use for job contracts, sequencing, and side-effect isolation.
---

# Event-Driven Workflow Orchestration

Design asynchronous workflows that are replay-safe, observable, and contract-stable across producers and workers.

## Decision Criteria

- Every job type must have schema version, idempotency key, producer service, and retry policy.
- Publish events only after authoritative state mutation is committed.
- Separate command events (request side effects) from fact events (state changed) to avoid semantic confusion.
- Require consumer behavior definitions for unknown versions and partial payloads.

## Principal Practices

- Standardize job envelope for `log_submitted`, `send_email`, `gdpr_delete`, and future workflows.
- Include correlation fields (`request_id`, `challenge_id`, `user_id`) for traceability across API logs and worker logs.
- Apply at-least-once safety assumptions in all consumers.
- Define dead-letter routing and replay tooling before enabling new message types.

## Failure Modes & Anti-Patterns

- Emitting event before database commit completion.
- Job handlers relying on implicit ordering guarantees from SQS standard queues.
- Encoding business-critical decisions in free-form message text.
- Mixing user-facing API contracts with internal queue payload contracts.

## Project-Specific Examples

- Daily log API should publish `log_submitted` only after the daily-lock/idempotency decision and persistence complete.
- Challenge publish flow should fan out `send_email` jobs with deterministic recipient selection and deduplicated message IDs.

## Related Skills

- `worker-reliability-engineering`
- `cloud-native-infrastructure-expert`
- `observability-engineering`
