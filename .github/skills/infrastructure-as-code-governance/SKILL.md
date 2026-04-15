---
name: infrastructure-as-code-governance
description: Principal governance for infrastructure as code evolution. Use for environment reproducibility, policy controls, and change safety.
---

# Infrastructure as Code Governance

Control infrastructure change with the same rigor as application code.

## Decision Criteria

- Any persistent infra configuration beyond local compose must be codified, versioned, and reviewable.
- IaC choices must support parity goals across local, staging, and production while preserving environment-specific policies.
- Infra changes require drift detection and rollback procedure definition.
- Security and cost implications must be evaluated in the same review.

## Principal Practices

- Keep environment modules explicit (networking, data stores, queues, object storage, secrets, compute) with ownership boundaries.
- Enforce policy checks for encryption, least-privilege access, logging, and retention.
- Use plan/apply approval gates tied to release governance.
- Maintain migration path from compose-local assumptions to cloud-managed services without implicit behavior shifts.

## Failure Modes & Anti-Patterns

- Editing infrastructure manually in cloud console without code reflection.
- Coupling application release cadence to untracked infrastructure drift.
- Shared resources without environment isolation and naming policy.
- Missing destroy protection for stateful resources.

## Project-Specific Examples

- Queue, bucket, and secret resources used by `fitchallenge-jobs` and `fitchallenge-assets` should be provisioned from IaC with tagged ownership.
- Production replacement for LocalStack bootstrap scripts must preserve service contracts and startup dependencies currently encoded in compose.

## Related Skills

- `cloud-native-infrastructure-expert`
- `production-environment-hardening`
- `finops-capacity-engineering`
