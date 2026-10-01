# Tasks 004 · Access graph and lenses

Plan: [`plan.md`](plan.md). One task per commit. Each ends with `go vet ./... && go test ./...`
and `./scripts/init.sh` green. Tick the box in the same commit. Rendering tasks render the
example PNG and look at it.

## A · Model (P6b-01)

- [x] **T01** `memberships`, `grants`, `auth` in the model; view type `access`, `focus`,
  `state`, `lenses`. Checks of FR-001, membership cycles, view fields. Schemas.
  Table-driven tests. Flip **P6b-01**.

## B · Semantics (P6b-02)

- [x] **T02** `internal/access`: graph per state, `Reach`, `Reachers`, `Escalations`.
  Table-driven tests on a small model. Flip **P6b-02**.

## C · Access view (P6b-03, P6b-04)

- [x] **T03** `resolve.Access`: layers, slots, lanes, ports, labels, cards. Test: a
  small system resolves to the expected layers and passes the lint.
- [x] **T04** Render the new edge kinds, the access legend and the diff state. Flip
  **P6b-03**, **P6b-04**.

## D · Lenses (P6b-05)

- [ ] **T05** `resolve.Lens`, `-lens`, dim, level badge, removed outline; the 3:1 test.
  Flip **P6b-05**.

## E · Example (P6b-06)

- [ ] **T06** `examples/access`: two tenants, a provisioning service, a root account,
  current and planned grants; access views in three states, a topology with lenses.
  Dark and light in `init.sh` and CI. Flip **P6b-06**.
