# Spec 004 · Access graph and lenses

- Status: **Done** (`plan.md`, `tasks.md`)
- Phase: 6b, built now alongside 6a on its own traversal (ADR-0020)
- Decisions it builds on: ADR-0003, ADR-0006, ADR-0007, ADR-0008, ADR-0010, ADR-0014,
  ADR-0015, ADR-0019, ADR-0020

## Why

Segregation of accounts is a review every platform team does, and it is drawn by hand
or not at all. The CNP review (2026-10-01) needed one answer per identity: what the CMP,
a tenant member and the root accounts can touch, today and after the migration. Today
the CMP holds a static Vault token, the Keycloak master admin and a long-lived AWS key;
the target leaves it four write grants on `project-*` prefixes. A reader understands
that in seconds on a picture and in minutes in a policy file.

The facts already sit next to the model (Vault policies, RoleBindings, Argo CD
`policy.csv`, IAM). Declaring them in the model lets one source draw the access graph,
light the infrastructure map per role and, later, answer "who can write X?" over MCP.

## User scenarios

1. **Platform engineer preparing a migration.** Declares the current grants, marks the
   ones to remove `deprecated` and the new ones `planned` with `target: P2`, then renders
   `access` focused on the CMP with `state: current` and `state: target`. Two PNGs go in
   the migration plan.
2. **Security reviewer.** Renders `access` focused on the Vault mount `project-a/`,
   walked backward: every identity that can write it, through which group and policy,
   and with which credential. A static token shows as a warning.
3. **Tech lead presenting to the team.** Renders the platform topology with three
   `lenses` (CMP, tenant member, root): the same map three times, lit differently. The
   PNGs share one layout, so a slide transition between them animates.
4. **AI agent.** Asked "can a tenant member read another tenant's secrets?", walks
   memberships and grants in `model.yaml` and answers with the path, or with "no path".

## Requirements

- **FR-001** `memberships` and `grants` in the model, with the fields of ADR-0019.
  Checks: every id exists; `level` in its enum; `target` only with `status: planned`;
  `via` ids exist; a subject is never its own resource.
- **FR-002** Effective access of a subject is the union of its grants and those of every
  group reachable through memberships, with `status` filtered by the requested state.
  Computed in `internal/access` (ADR-0020), table-tested.
- **FR-003** Derived escalation: a `write` or `admin` grant on an element that appears in
  another grant's `via` yields an escalation path to that grant's resource. Derived
  edges are marked, never stored in the model.
- **FR-004** A view `type: access` with `focus: {subject: id}` or
  `focus: {resource: id}` and `state: current | target | diff` resolves, lints and
  renders with no layout file. Columns: identities, groups, mechanisms, resources;
  row order computed to limit crossings, deterministic across runs.
- **FR-005** Access view content: level and scope on the last hop
  (`write · project-*`), `auth` on the first hop (`oidc`, `static-token`), a warning for
  a credential with no lifetime, escalation paths drawn as warnings, annotation badges
  on unverified grants. Cards: legend, reach summary (count per level, current and
  target), escalation paths found.
- **FR-006** In `state: diff`, `deprecated` grants render dotted and struck, `planned`
  grants hatched with their target pill, as elements do (ADR-0008).
- **FR-007** `lenses: [{id, subject, state}]` on a topology view. One output per lens, same
  layout. Reached elements keep their colour, get a level badge and a border by level;
  the others render out of focus. In `state: target`, an element the subject loses
  shows a ghost outline marked "removed".
- **FR-008** `noodle render <system> -view <id> -lens <lens-id>` renders that lens;
  without `-lens` the view renders as before.
- **FR-009** Out-of-focus style is distinct from `planned` and keeps text legible; the
  lint checks its contrast like any other text pair.
- **FR-010** JSON Schemas in `schemas/v1alpha1/` cover every new field; every example
  validates in `go test`.
- **FR-011** `examples/access`: a platform with two tenants, a provisioning service, a
  root account, current and planned grants. Renders the access view in its three states
  and the topology with two lenses, dark and light, in `init.sh` and CI.

## Success criteria

- The CNP segregation plan's two before/after pictures (CMP, tenant member) render from
  a model, reviewed by the `noodle:diagram-reviewer` agent with no blocking finding.
- An agent answers scenario 4 from the model alone, without reading policy files.
- Features P6b-01 to P6b-06 in `docs/features.json` pass their verify step.

## Out of scope

A matrix or table view (rejected in ADR-0019). Discovering grants from RBAC, Argo CD,
Vault or IAM (a later adapter, ADR-0015). Animated transitions between lenses (phase 2
SVG renderer, phase 3 viewer). Evaluating real policy languages: `scope` stays text.

## Decisions

The five questions left open by the draft were settled on 2026-10-01 and recorded in
ADR-0019:

1. **Phase 6b**, after 6a. Interactive lenses stay in phase 3.
2. **`v1alpha1`, no version bump**: the fields are optional and additive.
3. **`memberships` is a separate list**, group nesting allowed, cycles rejected.
4. **Access edges**: membership with a hollow diamond on the group, grant with a filled
   circle on the resource and width and colour by level, escalation orange and dashed.
5. **Out of focus**: neutral slate, greyscale icon, solid border, text kept at 3:1;
   level badge and border width by level on what is in focus.
