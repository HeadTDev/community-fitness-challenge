---
name: production-environment-hardening
description: Principal production hardening for runtime, configuration, and deployment posture. Use before enabling internet-facing production traffic.
---

# Production Environment Hardening

Harden runtime behavior so production safety does not depend on developer discipline.

## Decision Criteria

- Production profile must fail fast on insecure defaults, missing secrets, or debug-only endpoints.
- Any environment variable that changes security posture requires explicit documented default strategy.
- Exposed surface area (ports, routes, proxies) must be minimized and intentionally justified.
- Hardening is accepted only when verifiable by repeatable commands/checks.

## Principal Practices

- Disable development-only routes (`/auth/register-dev`) in production by enforced config policy.
- Remove `AWS_ENDPOINT_URL` and ensure IAM-backed AWS access path is used in production.
- Enforce strict container/runtime permissions, read-only mounts where possible, and non-root execution.
- Separate compose/dev runtime concerns from production deployment topology and secrets management.

## Failure Modes & Anti-Patterns

- Running production with development JWT secrets or permissive CORS assumptions.
- Keeping LocalStack compatibility flags enabled in production profile.
- Exposing internal data-plane services directly instead of through controlled ingress.
- Hardening checklists that are not automated or enforced in release gates.

## Project-Specific Examples

- Production auth flow must rely on Apple sign-in endpoint and prohibit temporary dev token generation paths.
- Nginx and API health endpoints must reveal operational state without leaking internal topology or credentials.

## Related Skills

- `container-runtime-security-engineering`
- `identity-access-security-engineering`
- `release-management-governance`
