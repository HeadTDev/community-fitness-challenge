---
name: mobile-security-privacy-engineering
description: Principal mobile security and privacy engineering. Use for secure token handling, sensitive data minimization, and platform trust controls.
---

# Mobile Security & Privacy Engineering

Ensure iOS client behavior protects identity, health data, and session integrity under real-world attack surfaces.

## Decision Criteria

- Sensitive data storage must use platform-secure primitives and data minimization.
- Authentication/session mechanisms must withstand token theft and replay scenarios.
- Privacy-sensitive telemetry must preserve utility without exposing user secrets.
- Client security controls must align with backend trust assumptions and rate-limit behavior.

## Principal Practices

- Store tokens in Keychain with strict accessibility settings and lifecycle handling.
- Minimize health/profile data persisted offline and encrypt where platform policy requires.
- Ensure logout and account deletion purge sensitive local artifacts and pending queue payloads.
- Apply transport security best practices and certificate/pinning strategy when required by threat model.

## Failure Modes & Anti-Patterns

- Persisting tokens or PII in plain user defaults/cache snapshots.
- Debug logging that includes auth headers or health metrics.
- Retaining deleted-user data in local cache after GDPR workflow completion.
- Silent auth expiration leaving app in privileged UI state.

## Project-Specific Examples

- Sign-in with Apple flow must avoid exposing raw identity tokens beyond required verification handoff.
- HealthKit-derived daily logs should be retained only as long as needed for sync and user-visible history policy.

## Related Skills

- `identity-access-security-engineering`
- `gdpr-data-lifecycle-engineering`
- `threat-modeling-secure-design`
