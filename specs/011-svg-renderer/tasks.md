# Tasks 011 · Own SVG and PNG renderer

Plan: [`plan.md`](plan.md). One task per commit; each ends with `make check` green.
Tick the box in the same commit. Tasks in one wave do not depend on each other, except
where a task names another. The five questions are answered (2026-11-05) and ADR-0033
to ADR-0036 accepted.

## Wave 1 · Groundwork, no visible change

- [x] **T01** Spec, ADRs, plan, tasks, contract; C04 amended; `P2-*` features added.
- [ ] **T02** `internal/scene` and `internal/render/drawio`: the draw.io writer reads the
  scene for every view type; every `.drawio` byte-identical (FR-007). Flip **P2-07**.
- [ ] **T03** `assets/fonts` and `internal/font`: advances, coverage, subsets; text width
  in `house` from the font, floored at `ExportCharWidthEm` when a `.drawio` is written;
  `lint.Finding` gains `Rule` and `Rect`, set by every rule; rule `glyph_coverage`
  (FR-004). P2-04 flips with T04, whose SVG output its row #3b needs.

## Wave 2 · Topology to SVG and PNG

- [ ] **T04** `internal/render/svg` for topology views, with the diff and highlight states:
  ids, styles, text runs, font subsets, icons inlined (FR-001 for topology). Flip **P2-04**.
- [ ] **T05** Rounded corners, jump arcs and label backgrounds in the scene, drawn by
  the SVG writer.
- [ ] **T06** `internal/render/png`: stroker, text, `-scale`; the 3725 px case (FR-002).
- [ ] **T07** Icon reader for the PNG; every curated icon in both variants renders.
  P2-01 and P2-02 are graded on noodle's committed images, so they flip after T10.
- [ ] **T08** `-animate` (FR-005). Flip **P2-05**.
- [ ] **T09** `-draft` (FR-006), drawing `Finding.Rule` at `Finding.Rect`, refusing
  `.drawio`. Flip **P2-06**.
- [ ] **T10** Topology examples commit noodle's `.svg` and PNGs; `make examples` checks
  them; PR with before/after crops of every example. Flip **P2-01** and **P2-02**.

## Wave 3 · Sequence and landscape

- [ ] **T11** Sequence views in the SVG and PNG writers; examples switched (FR-003).
- [ ] **T12** Landscape views; examples switched (FR-003).

## Wave 4 · Catalog and access matrix, documentation

- [ ] **T13** Catalog and access matrix views; examples switched (FR-003). Flip **P2-03**.
- [ ] **T14** Skill, README and tutorial export without draw.io; `docs/diagrams/`
  regenerated (FR-008). Flip **P2-08**.

## Convergence

Tasks the evaluator found missing, appended here; never edit the waves above.
