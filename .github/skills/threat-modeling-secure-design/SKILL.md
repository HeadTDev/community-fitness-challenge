---
name: threat-modeling-secure-design
description: Principal threat modeling and secure design discipline. Use before implementing high-impact API, worker, websocket, and data workflows.
---

# Threat Modeling & Secure Design

Model adversary paths early so security controls are designed into architecture, not bolted on after incidents.

## Decision Criteria

- Threat model is mandatory for new trust boundaries: auth, queue ingestion, websocket channels, file upload, and admin/creator actions.
- Analyze abuse paths for both authenticated and unauthenticated actors.
- Prioritize mitigations by exploitability and user/data impact.
- Security decisions must include testable controls and operational detection signals.

## Principal Practices

- Use attack-surface maps spanning HTTP endpoints, background jobs, Redis channels, and object storage.
- Define abuse cases for rate limiting bypass, token misuse, replay attacks, and payload tampering.
- Pair preventive controls with detective controls (metrics, anomaly alerts, audit logs).
- Revisit threat models when contracts or infrastructure topology changes.

## Failure Modes & Anti-Patterns

- Treating queue payloads as trusted internal input.
- Security review limited to code diff without dataflow analysis.
- Threat models that ignore mobile token storage and reconnect behavior.
- Lack of explicit owner for mitigation follow-through.

## Project-Specific Examples

- Avatar/cover upload paths need validation against content-type spoofing and object-key abuse.
- Websocket rank-update channel requires authorization and replay protection assumptions aligned with JWT/session semantics.

## Related Skills

- `system-integrity-guardian`
- `mobile-security-privacy-engineering`
- `incident-response-leadership`
