# Spec 004 · Access matrix

- Status: **Done** (`plan.md`, `tasks.md`)
- Phase: 6b, built now alongside 6a on its own traversal (ADR-0020)
- Decisions it builds on: ADR-0006, ADR-0007, ADR-0008, ADR-0010, ADR-0014, ADR-0015,
  ADR-0019, ADR-0020

## Why

Segregation of accounts is a review every platform team does, usually in a spreadsheet
nobody keeps up to date. The facts sit next to the model (Vault policies, RoleBindings,
Argo CD `policy.csv`, IAM). Declared in the model, they render as the access matrix an
ISO 27001 review asks for (A.5.15, A.5.18), today, after a plan, or both in one picture.

## User scenarios

1. **Platform engineer preparing a migration.** Declares today's grants, marks those the
   plan removes `deprecated` and those it adds `planned`, each with its phase as
   `target`, and renders one matrix in `state: diff`. The migration plan carries it.
2. **Security reviewer.** Reads a column: every identity that can touch the tenant A
   secrets, at which level. Reads a row: everything the CMP can do.
3. **AI agent.** Asked "can a tenant member write another tenant's secrets?", walks
   memberships and grants in `model.yaml` and answers from the model alone.

## Requirements

- **FR-001** `memberships` and `grants` in the model, with the fields of ADR-0019.
  Checks: every id exists; `level` in its enum; `target` only with `planned` or
  `deprecated`; `via` ids exist; a subject is never its own resource; no membership
  cycle.
- **FR-002** `internal/access` gives the strongest level of a subject on each resource
  in a state (`current` drops planned items, `target` drops deprecated ones), through
  every group it belongs to. Break-glass counts only where nothing stronger applies.
  Table-tested.
- **FR-003** A view `type: matrix` with `rows` and `columns` (titled groups of element
  ids, optional colour and subtitle), `labels` (header text per element), and `state:
  current | target | diff` resolves, lints and renders with no layout file.
- **FR-004** A cell's fill is the level: admin, write, read, break-glass (hatched),
  none. In `diff`, its frame and corner tag show the change: added (solid, `+`),
  changed (dashed, `→`, the former level under the new one), removed (dotted, `−`,
  hatched grey, the former level struck), each tag carrying the phase. A cell whose
  level is unchanged but whose grants changed is marked changed, `rescoped`.
- **FR-005** A last column counts, per row, the resources it can change (write or admin),
  before and after in `diff`.
- **FR-006** The legend sits under the matrix: one line for levels, one for changes,
  with sample cells and one word each.
- **FR-007** Colours come from the theme in both themes; level and change never share
  a colour; the lint checks that every header and cell text fits.
- **FR-008** JSON Schemas cover every new field; every example validates in `go test`.
- **FR-009** `examples/cnp-access`: the CNP platform's access model, a matrix in `diff`
  state, dark and light in `init.sh` and CI.

## Success criteria

- Features P6b-01 to P6b-05 in `docs/features.json` pass their verify step.
- The CNP access matrix renders from `cnp-docs/architecture/cnp`, reviewed by the
  platform owner.

## Out of scope

The access graph and lenses (ADR-0019, kept on `feat/access-views`). Path and
escalation analysis. Discovering grants from RBAC, Argo CD, Vault or IAM (a later
adapter, ADR-0015).
