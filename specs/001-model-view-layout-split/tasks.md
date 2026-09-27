# Tasks 001 · Split model, views and layout

Plan: [`plan.md`](plan.md). One task per commit. Each ends with `go vet ./... && go test ./...`
and `./scripts/init.sh` green. Tick the box in the same commit.

## A · Contracts (P1-02)

- [x] **T01** Move `Spec`, `Zone`, `Node`, `Edge`, `Note`, `Card`, `Rect`, `Point` from
  `cmd/noodle/spec.go` to `internal/diagram`. No behaviour change.
- [x] **T02** `internal/model`: `Model`, `View`, `Layout` types per ADR-0007 and
  ADR-0008, strict decoding (`KnownFields`), `LoadSystem(dir)` returning all three.
  Tests: a minimal system loads; an unknown field fails with file and line.
- [x] **T03** Reference checks: every id used in a view or layout exists in the model;
  a topology layout covers every included element; ports named by connections exist on
  the target. Table-driven tests, one failing case per rule.
- [ ] **T04** `schemas/v1alpha1/{model,view,layout}.schema.json` and a test that
  validates every `examples/**/model.yaml`, `views/*.yaml`, `layouts/*.yaml`.
  Flip **P1-02**.

## B · Resolver (P1-03)

- [ ] **T05** `internal/resolve`, topology with absolute values only: elements, zones,
  connections, references, cards, notes → `diagram.Spec`. Test: a small system resolves
  to the expected `Spec`.
- [ ] **T06** Positions relative to the parent zone.
- [ ] **T07** Symbolic endpoints `id.side[@ratio]`; error on unknown side or ratio out of
  0–100 %.
- [ ] **T08** Named lanes: `lane:<name>` waypoints; test that two edges on one lane stay
  parallel when the lane moves.
- [ ] **T09** Steps, labels and badges: `[n]`/`[A]` prefixes from view steps, default
  `verb · protocol`, overrides, annotation badges, `denied` → blocked style.
- [ ] **T10** CLI: `noodle render <dir> --view <id> [-theme] [-o]`, lint only without
  `-o`. The v0 single-file mode stays.

## C · Lint (P1-05)

- [ ] **T11** `internal/lint`: one file per rule from `cmd/noodle/lint.go`, same messages.
  Each rule has a passing and a failing table case. Flip **P1-05**.

## D · Migration (P1-04, P1-03)

- [ ] **T12** `examples/cnp-runtime/{model.yaml,views/runtime.yaml,layouts/runtime.yaml}`
  with absolute-equivalent values. Parity: PNG of the split format equals the v0 PNG
  (`magick compare -metric AE` = 0). Add to `init.sh` and CI. Flip **P1-03**, **P1-04**.
- [ ] **T13** Rewrite the CNP layout with relative positions, symbolic endpoints and
  lanes where they apply; parity check again.

## E · Sequence (P1-06)

- [ ] **T14** Sequence resolver: participants in first-appearance order or from
  `participants`, one row per step, dashed returns only in sequence views.
- [ ] **T15** `examples/cnp-runtime/views/oidc-login.yaml` from the existing OIDC
  connections; both views render from one `model.yaml`, lint green. Flip **P1-06**.

## F · Cleanup

- [ ] **T16** Migrate `examples/platform-regression` to one directory per diagram.
- [ ] **T17** Remove the v0 reader and `runtime-request-path.yaml` (FR-009); update the
  skill and README examples to the split format.
