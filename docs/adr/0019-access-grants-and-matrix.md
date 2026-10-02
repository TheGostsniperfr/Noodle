# ADR-0019 · Access grants live in the model; a matrix view shows them

- Status: Accepted
- Date: 2026-10-02
- Extends: ADR-0007 (contract files), ADR-0008 and ADR-0014 (`status`, `target`),
  ADR-0010 (`enforced_by`)

## Context

A segregation-of-accounts review of the CNP platform needed one answer per identity:
what the CMP, a tenant member, the root accounts can touch, today and after the plan.
The model cannot say it: a connection is traffic (ADR-0003), a reference is
configuration.

Four pictures were tried on the real platform. An access graph (identity, mechanism,
resource, laid out in layers) was correct but unreadable: one subject with a dozen
grants turns the first corridor into a knot. Lenses on a topology read well but say
little. A role model (identities, role cards, components by plane) needs roles as a
first-class concept. An access control matrix was the one the platform owner read
without help.

The matrix is also what the industry asks for: ISO/IEC 27001:2022 A.5.15 and A.5.18
expect access rules and access rights documented and reviewed, usually as a role or
access matrix. Graphs belong to attack-path tools (BloodHound, PMapper) that query a
live environment.

## Decision

**Grants are facts, so they live in the model**, like `denied` (ADR-0007).

- `memberships: [{id, subject, group, auth, status, target}]`: an identity belongs to a
  group or assumes a role.
- `grants: [{id, subject, resource, level, scope, via, auth, status, target}]`. `level`
  is `read | write | admin | breakglass`; `scope` is free text; `via` lists the
  elements that give the grant and is checked like `enforced_by` (ADR-0010).
- `auth: {method, lifetime, mfa}` on the hop where an identity proves itself.
- **Before and after use `status`**, not a second model: `deprecated` is what a plan
  removes, `planned` what it adds. `target` names the plan phase on both, since a
  removal happens in a phase too. This widens ADR-0014 for memberships and grants only.

**One view reads them: `type: matrix`.** Rows are identities, columns are resources,
both in titled groups the view declares. A cell shows the level after the plan; in
`state: diff` its frame and tag show what the plan changes (added, changed, removed)
and in which phase. Two separate codes: the fill is the level, the frame is the change.
The legend sits under the matrix.

The access graph and lenses are set aside, not rejected for good: their code is kept on
the closed branch `feat/access-views` (#53) for a later look.

## Consequences

- `v1alpha1` additions, no version bump, as ADR-0014 and ADR-0015 did.
- A matrix shows the effective level per identity and resource, not the mechanism. The
  mechanism stays in the model (`via`) for the MCP server and a later path view.
- Discovery adapters can later emit memberships and grants (RBAC, Argo CD `policy.csv`,
  Vault policies, IAM) into fragments (ADR-0015).
