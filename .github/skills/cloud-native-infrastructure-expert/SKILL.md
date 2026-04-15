---
name: cloud-native-infrastructure-expert
description: Principal cloud-native infrastructure leadership for LocalStack-to-AWS parity. Use for resilient service integration, operational controls, and production readiness.
---

# Principal Cloud-Native Infrastructure

Design AWS-integrated systems that behave predictably in local simulation and production environments without hidden coupling.

## Decision Criteria

- Treat LocalStack as compatibility aid, not behavior guarantee; production assumptions must be explicitly validated.
- For each AWS service, define error class handling, retry policy, and fallback behavior before integration.
- Keep credentials and endpoint configuration environment-driven; never hard-code production-sensitive values.
- Require service health verification commands for S3/SQS/SES/Secrets in CI and release checklists.

## Principal Practices

- Keep all AWS clients behind `internal/aws` interfaces and inject through `internal/app` composition root.
- Use queue naming, bucket prefixing, and secret path naming conventions that support multi-environment isolation.
- Ensure S3 object lifecycle and prefix structure align with profile avatars, challenge covers, and GDPR deletion workflows.
- Enforce queue redrive and operational visibility for worker pipelines.

## Failure Modes & Anti-Patterns

- Shipping code that requires `AWS_ENDPOINT_URL` in production.
- Treating SES success in LocalStack logs as proof of real delivery reliability.
- Storing unversioned secrets or mixing test and production secret names.
- Direct AWS SDK usage in handlers/services bypassing shared adapters.

## Project-Specific Examples

- `fitchallenge-assets` object layout must support selective deletes (`avatars/`, `challenges/`) during GDPR and challenge lifecycle operations.
- Queue `fitchallenge-jobs` workload types (`log_submitted`, `send_email`, `gdpr_delete`) require explicit retry and DLQ behavior.

## Related Skills

- `event-driven-workflow-orchestration`
- `production-environment-hardening`
- `secrets-key-management-governance`
