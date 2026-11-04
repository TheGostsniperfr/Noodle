# Tasks 008 · Lint robustness

Plan: [`plan.md`](plan.md). One task per commit; each ends with `make check` green.
Tick the box in the same commit. Tasks in one wave do not depend on each other.
Clarifications 1 to 5 of the spec are answered (2026-11-04).

## Wave 1

- [x] **T01** `house.HeaderBox` and `house.CardLayout`; `render.go` draws header, cards and
  legend from them; no behaviour change, every example's `.drawio` byte-identical to `main`
  (groundwork for FR-001, FR-002, FR-003).

## Wave 2

- [x] **T02** `card_text.go`: width of every card, note and legend row, and card height
  (FR-001, FR-002). Flip **LNT-01** after `make evaluate FEATURE=LNT-01`; flip backlog
  **B-07** and **B-26** with `docs/evidence/` pointing at that evidence.
- [x] **T03** `header_overlap.go` (FR-003). Flip **LNT-02**; flip backlog **B-25**.
- [x] **T04** Pairwise port badge overlap in `ports.go`, the report's repro as a fixture
  (FR-005). Flip **LNT-04**.
- [x] **T05** Badge join on empty labels, `hide_annotations`, and the ADR-0026 checks of
  undrawn annotations and `hidden_annotations` (FR-006, FR-007, FR-011, FR-012). Flip
  **LNT-05**.
- [x] **T06** `canvas_bounds.go` and the landscape side column (FR-009). Flip **LNT-07**.

## Wave 3

- [x] **T07** Text margin in `node_text.go` and `resolve.boxWidth`, with the measuring
  width of clarification 1; the four participant labels of #72 as a fixture (FR-004).
  Flip **LNT-03**.
- [x] **T08** Zone sub layout and title obstacle per clarification 3 (FR-008). Flip
  **LNT-06**; flip backlog **B-20**; close #22 and #87.

## Wave 4

- [ ] **T09** Fix every example, regression fixture and docs diagram the new rules reject;
  before/after PNGs in the PR (FR-010). Flip **LNT-08**.

## Convergence

Tasks the evaluator found missing, appended here; never edit the waves above.
