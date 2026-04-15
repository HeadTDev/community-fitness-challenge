---
name: container-runtime-security-engineering
description: Principal container and runtime security engineering. Use for image hardening, runtime isolation, and secure service execution.
---

# Container & Runtime Security Engineering

Harden containerized services so compromise requires multiple failures, not one misconfiguration.

## Decision Criteria

- Base image, package footprint, and runtime privileges must be justified per service.
- Security controls must preserve required runtime behavior without relying on undocumented exceptions.
- Any new exposed port or mount point requires threat review and least-privilege analysis.
- Service startup dependencies must not require broad host access.

## Principal Practices

- Use minimal images, pinned versions, and regular vulnerability scanning.
- Run workloads as non-root; apply read-only filesystem and dropped Linux capabilities where compatible.
- Restrict inter-service network access to required traffic only (data-net/app-net intent).
- Isolate secrets from logs and environment dumps.

## Failure Modes & Anti-Patterns

- Broad Docker socket exposure in production-equivalent environments.
- Debug tools left installed in runtime images.
- Shared mutable volumes between unrelated services.
- Running worker/API with unrestricted outbound access when not required.

## Project-Specific Examples

- `Dockerfile.dev` and future production Dockerfiles should diverge intentionally: dev convenience (`air`) must not leak into production images.
- Nginx/websocket/api containers should enforce strict proxy path and header handling without permissive defaults.

## Related Skills

- `production-environment-hardening`
- `threat-modeling-secure-design`
- `disaster-recovery-business-continuity`
