---
name: finops-capacity-engineering
description: Principal FinOps and capacity engineering for cloud workloads. Use for cost/performance tradeoffs and sustainable scaling.
---

# FinOps & Capacity Engineering

Scale performance without uncontrolled cloud cost growth.

## Decision Criteria

- Capacity plans must include throughput targets, latency SLOs, and monthly cost guardrails.
- Choose service configuration from workload profile (burst vs steady-state, latency sensitivity, retention needs).
- Evaluate cost impact of every asynchronous and real-time feature before rollout.
- Cost controls must be observable and enforceable, not spreadsheet-only.

## Principal Practices

- Right-size DB/Redis connection pools and worker concurrency against real usage patterns.
- Tune SQS polling and batch parameters for cost-latency balance.
- Apply S3 lifecycle and object retention policy for avatars/covers and transient artifacts.
- Track per-feature cost attribution (notifications, websocket fanout, leaderboard rebuilds).

## Failure Modes & Anti-Patterns

- Overprovisioning based on peak fear without autoscaling strategy.
- Aggressive polling intervals that inflate cost with negligible user benefit.
- Unbounded log retention or object growth without lifecycle policy.
- Ignoring egress-heavy websocket traffic in cost planning.

## Project-Specific Examples

- Rank-update websocket frequency should be throttled/debounced to protect mobile battery and backend egress costs.
- Notification pipeline must quantify cost per delivered event across SQS + SES and define budget alarms.

## Related Skills

- `distributed-systems-optimizer`
- `ci-cd-pipeline-engineering`
- `technical-risk-tradeoff-management`
