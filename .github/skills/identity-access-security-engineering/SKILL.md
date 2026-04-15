---
name: identity-access-security-engineering
description: Principal identity and access security engineering. Use for JWT lifecycle, role authorization, and sign-in trust chains.
---

# Identity & Access Security Engineering

Ensure authentication and authorization stay correct under retries, token refresh, and multi-channel client behavior.

## Decision Criteria

- Define token issuance, refresh, revocation, and expiry behavior as explicit lifecycle policy.
- Role/permission checks must be centrally enforceable and auditable.
- Identity-provider integration changes require trust-chain and fallback analysis.
- Access model must be consistent across REST and websocket access paths.

## Principal Practices

- Validate JWT signing method, expiry, and claims defensively in middleware.
- Keep role checks in service layer for sensitive operations (create/publish challenges, prize management, analytics actions).
- Introduce Apple sign-in flow with explicit user-linking and account lifecycle rules.
- Implement secure refresh token handling with replay-resistance strategy.

## Failure Modes & Anti-Patterns

- Token validation that accepts unexpected algorithms or missing claims.
- Authorization checks distributed inconsistently across handlers.
- Assuming dev-only identity paths are harmless in non-dev environments.
- Missing audit trail for privileged action execution.

## Project-Specific Examples

- `POST /auth/apple` must map identity token validation to existing/new user record strategy without creating duplicate accounts.
- `/v1/challenges/:id/publish` authorization must remain consistent for creator/admin roles across API and worker-triggered side effects.

## Related Skills

- `system-integrity-guardian`
- `api-contract-governance`
- `mobile-security-privacy-engineering`
