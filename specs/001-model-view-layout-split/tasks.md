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
- [x] **T04** `schemas/v1alpha1/{model,view,layout}.schema.json` and a test that
  validates every `examples/**/model.yaml`, `views/*.yaml`, `layouts/*.yaml`.
  Flip **P1-02**.
- [x] **T04b** Write the CNP example in the split format next to the v0 file
  (`examples/cnp-runtime/{model.yaml,views/runtime.yaml,layouts/runtime.yaml}`), checked
  by the schemas and `model.Check` in `go test`. Findings folded into the contract:
  endpoints line up with their neighbour by default and accept `@NNpx`; annotations may
  target connections; `denied` is intent and `enforced_by` names what enforces it
  (ADR-0010).

## B · Resolver (P1-03)

- [x] **T05** `internal/resolve`, topology with absolute values only: elements, zones,
  connections, references, cards, notes → `diagram.Spec`. Test: a small system resolves
  to the expected `Spec`.
- [x] **T06** Positions relative to the parent zone.
- [x] **T07** Symbolic endpoints `id.side[@NN%|@NNpx]`. Without `@`, the endpoint lines
  up with the next waypoint, or with the other end when there is none; a `from` with
  neither sits at 50 %. Error on unknown side, ratio out of 0–100 %, offset past the side.
- [x] **T08** Named lanes: `lane:<name>` waypoints; test that two edges on one lane stay
  parallel when the lane moves.
- [x] **T09** Steps, labels and badges: `[n]`/`[A]` prefixes from view steps, default
  `verb · protocol` (references: their `kind`), overrides, annotation badges on
  elements and ` · !!⚠ Gx!!` after the label of annotated edges, `denied` → blocked style.
- [x] **T10** CLI: `noodle render <dir> --view <id> [-theme] [-o]`, lint only without
  `-o`. The v0 single-file mode stays.

## C · Lint (P1-05)

- [x] **T11** `internal/lint`: one file per rule from `cmd/noodle/lint.go`, same messages.
  Each rule has a passing and a failing table case. Flip **P1-05**.

## D · Migration (P1-04, P1-03)

- [x] **T12** Render the split CNP example written in T04b. Parity: its PNG equals the
  v0 PNG (`magick compare -metric AE` = 0). Add to `init.sh` and CI. Flip **P1-03**,
  **P1-04**.
- [x] **T13** Rewrite the CNP layout with relative positions, symbolic endpoints and
  lanes where they apply; parity check again.

## E · Sequence (P1-06)

- [x] **T14** Sequence resolver: participants in first-appearance order or from
  `participants`, one row per step, dashed returns only in sequence views.
- [x] **T15** `examples/cnp-runtime/views/oidc-login.yaml` from the existing OIDC
  connections; both views render from one `model.yaml`, lint green. Flip **P1-06**.

## F · Cleanup

- [x] **T16** Migrate `examples/platform-regression` to one directory per diagram.
  Done with `noodle migrate`, draw.io byte-identical to v0 in both themes.
- [x] **T17** Remove the v0 reader and `runtime-request-path.yaml` (FR-009); update the
  skill and README examples to the split format. The CLI no longer renders v0; its
  reader stays only behind `noodle migrate` (agreed 2026-09-30) until PAE, DockAir and
  cnp-docs have moved, then goes.
