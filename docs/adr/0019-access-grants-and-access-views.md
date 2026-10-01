# ADR-0019 · Access grants live in the model; an access graph and lenses show them

- Status: Accepted
- Date: 2026-10-01
- Extends: ADR-0007 (contract files), ADR-0008 and ADR-0014 (`status`, `target`),
  ADR-0010 (`enforced_by`)

## Context

A segregation-of-accounts review of the CNP platform needed one question answered per
identity: "what can the CMP, a tenant member, a root account touch, today and after the
migration?". The model cannot say it. A connection is traffic (ADR-0003), a reference is
configuration; neither carries "X may write Y, limited to `project-*`". Drawing grants
as arrows would break ADR-0003 and gives about fifty arrows for six identities. A
read/write matrix was tried and rejected: precise, but detached from the architecture.

Security tools show access two ways: a graph of identity, role and resource queried for
one path (PMapper, BloodHound, Cartography), and a blast-radius highlight on a map (Wiz).
Both work on the live state. None declares access in a versioned model with a current and
a target state, next to the architecture.

## Decision

**Grants are facts, so they live in the model**, like `denied` (ADR-0007).

- `memberships: [{id, subject, group, auth, status, target}]`: an identity belongs to a
  group or assumes a role (user to Keycloak group, ServiceAccount to Vault auth role).
- `grants: [{id, subject, resource, level, scope, via, auth, status, target}]`. `level`
  is `read | write | admin | breakglass`; `scope` is free text (`project-*`,
  `namespace a`); `via` lists the elements that give the grant (a Vault policy, a
  RoleBinding, an IAM role) and is checked like `enforced_by` (ADR-0010).
- `auth: {method, lifetime, mfa}` on the hop where an identity proves itself:
  `static-token`, `access-key`, `password`, `oidc`, `k8s-sa`, `federated`.
- Identities, groups and mechanisms are ordinary elements. No new element kind.
- **Before and after use `status`**, not a second model: a grant to remove is
  `deprecated`, a grant to add is `planned` with its `target` (ADR-0008, ADR-0014).
  ADR-0007 keeps one model per system.

**Two views read them; there is no matrix view.**

- `type: access`, computed like landscape and catalog, no layout file. One `focus`
  (a subject, walked forward, or a resource, walked backward) answers one question.
  Columns: identities, groups, mechanisms, resources. `state: current | target | diff`.
  Escalation is derived: a write or admin grant on an element used in some `via` is a
  path to what that element grants, drawn as a warning.
- `lenses` on a topology view: the same drawing and layout, one output per lens.
  Elements the subject reaches keep their colour and get a level badge; the others are
  out of focus. This is the static form of phase 3's "focus and dim".

**Settled with Brian on 2026-10-01** (spec 004's open questions):

- **Phase 6b**, right after 6a: the access view needs `internal/graph`, lenses need only
  the draw.io renderer. Interactive lenses stay in phase 3.
- **`v1alpha1`, no bump.** Every field is optional and additive. A bump is for breaking
  changes, as in Kubernetes API versioning; ADR-0014 and ADR-0015 did the same.
- **Memberships are a separate list**, as RBAC separates user assignment from
  permission assignment (NIST RBAC, Kubernetes subjects and rules). Group nesting is
  allowed; a membership cycle is a check error.
- **Access edges are a third style.** Membership: thin line, hollow diamond on the group
  (UML aggregation). Grant: solid line, filled circle on the resource, width and colour
  by level, label always present. Escalation: orange dashed line labelled `escalation`.
  Connections keep block arrows, references keep dotted open arrows.
- **Out of focus** means neutral slate fill and stroke, icon in greyscale at 40 %, solid
  border. No hatch and no dash, so it never reads as `planned` or `deprecated`. Dimmed
  context is exempt from 4.5:1 like inactive UI (WCAG 1.4.3) but the lint keeps it at
  3:1. In focus: a level badge top right (a target pill moves bottom right), border
  width by level. A grant lost in `state: target` uses the `deprecated` dotted border
  with a `removed` badge.

## Consequences

- `v1alpha1` additions, no version bump, as ADR-0014 and ADR-0015 did.
- The access view needs reachability queries, so it follows `internal/graph`
  (spec 003 T09). The MCP server (phase 2b) can answer "who can write X?" from it.
- Out-of-focus styling must stay distinct from `planned` and keep text legible; the
  renderer spec sets the values and the lint checks them.
- Discovery adapters can later emit memberships and grants (RBAC, Argo CD
  `policy.csv`, Vault policies, IAM) into fragments (ADR-0015). Not in this ADR.
