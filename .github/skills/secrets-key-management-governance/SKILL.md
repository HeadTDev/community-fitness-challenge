---
name: secrets-key-management-governance
description: Principal governance for secrets and key management. Use for credential lifecycle, rotation policy, and secure secret consumption.
---

# Secrets & Key Management Governance

Treat secrets as rotating operational dependencies, not static configuration text.

## Decision Criteria

- Every secret must have owner, rotation interval, and blast-radius classification.
- Secrets access must follow least privilege and environment separation.
- Secret retrieval failures require explicit service behavior (fail-fast, retry, fallback) per use case.
- Key/secret usage in code must be testable without exposing real values.

## Principal Practices

- Use Secrets Manager path conventions that separate environments and service scopes.
- Never embed JWT secrets, API keys, or provider credentials in source, compose defaults, or logs.
- Require rotation runbooks and validation tests for auth secrets and provider credentials.
- Ensure secret access telemetry exists for auditing and anomaly detection.

## Failure Modes & Anti-Patterns

- Long-lived secrets with no rotation ownership.
- Shared secrets across dev/stage/prod.
- Printing effective config that leaks secret values.
- Handling secret fetch errors by silent insecure defaults.

## Project-Specific Examples

- `JWT_SECRET` and external provider credentials must transition from env defaults to managed secret retrieval in production mode.
- Worker and API services should consume identical secret contract while maintaining distinct least-privilege IAM scopes.

## Related Skills

- `cloud-native-infrastructure-expert`
- `production-environment-hardening`
- `identity-access-security-engineering`
