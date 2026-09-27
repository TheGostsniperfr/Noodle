# ADR-0010 · A denied connection names what enforces it

- Status: Accepted
- Date: 2026-09-27
- Refines: ADR-0007 (`denied` on connections)

## Context

ADR-0007 puts `denied: true` in the model so the MCP server sees refusals. Writing the
CNP example showed two meanings under one flag: cross-tenant traffic must not happen
(the design intent), yet gap G1 says no Cilium policy blocks it. Read as "a policy
refuses it", the flag would state something false. Read as intent, an agent answering
"who can reach this database?" would miss that nothing enforces it.

## Decision

- `denied: true` is the intent: this connection must not happen.
- `enforced_by: [ids]` lists the model elements that make it so: a NetworkPolicy, a
  CiliumClusterwideNetworkPolicy, a security group. Every id must exist in the model.
- A denied connection without `enforced_by` is intent only. The lint reports it, which
  turns a gap like G1 from hand-written text into a finding.
- `enforced_by` is allowed on denied connections only.

A bare `intended | enforced` flag was rejected: "enforced" would be a claim with nothing
to check. Pointing at the enforcing object lets `model.Check` verify it exists and lets
a discovery adapter (phase 6) confirm or contradict it in the drift report.

## Consequences

- Enforced refusals need the policy as an element, which a security review wants anyway.
- Rendering can tell enforced from intended refusals; the style is decided with the
  renderer, not here.
