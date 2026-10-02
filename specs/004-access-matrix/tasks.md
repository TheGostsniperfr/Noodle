# Tasks 004 · Access matrix

Plan: [`plan.md`](plan.md). One task per commit. Each ends with `go vet ./... && go test ./...`
and `./scripts/init.sh` green. Tick the box in the same commit. Rendering tasks render the
example PNG and look at it.

- [x] **T01** Model, checks and schemas for memberships, grants, auth; view type
  `matrix`, `rows`, `columns`, `state`. Table-driven tests. Flip **P6b-01**.
- [x] **T02** `internal/access`: groups, grants, levels per state, grant ids per
  resource. Table-driven tests. Flip **P6b-02**.
- [x] **T03** `resolve.Matrix` and `diagram.Matrix`; the matrix text lint. Test: a small
  system resolves to the expected cells and changes. Flip **P6b-03**.
- [x] **T04** Render cells, frames, tags, hatches, counts and legend in both themes.
  Flip **P6b-04**.
- [ ] **T05** `examples/cnp-access` in `init.sh` and CI; skill reference and README.
  Flip **P6b-05**.
