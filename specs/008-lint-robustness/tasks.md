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

- [x] **T09** Fix every example, regression fixture and docs diagram the new rules reject;
  before/after PNGs in the PR (FR-010). Flip **LNT-08**.

## Convergence

Tasks the evaluator found missing, appended here; never edit the waves above.

- [x] **T10** LNT-04 #3: a shared port is tested with two connections that meet only on the
  final point (no collinear finding), asserting 0 findings; two different ports on one point
  now yield a finding instead of one badge silently dropping the other (FR-005).

- [x] **T11** LNT-01 #1: a 420 px legend with the `link` edge kind yields exactly one finding:
  the step and port rows get the room up to the card's inner edge when no right-column row
  shares their height (`CardRow.Room`). The shrink-probe test sets `H` before use, covers
  notes, and a render test checks every title and listed line is drawn (FR-001).

- [x] **T12** LNT-05 #1 and #5: the sequence cases assert the thresholds' values (`!!⚠ G1!!`,
  `HTTPS · !!⚠ G1!!`, `HTTPS`) on the label after its step number, with a message whose text is
  `""` over a connection that has no verb or protocol (FR-006, FR-012). A sequence numbers every
  message (ADR-0011), so the label itself keeps its `[1] ` prefix.

- [x] **T13** LNT-06: for a zone with an icon the wrapped sub is drawn after the icon, so
  `ZoneTitleBox` adds the icon offset to the widest line; tests for the box and for an edge
  beside and through the sub of an icon zone (FR-008).

- [x] **T14** LNT-07: #90 reproduced from the real view of agent-playbook PR #5 (commit
  e119552c18, side column restored) on c2cf233: the side column sits inside the page in
  the `.drawio` (right edge 2140 of 2180), in the PNG export of #90 (`-s 2 --border 20`,
  4422 px) and in the SVG export (2241 px). The resolver places it from the canvas width by
  construction, so no resolver change is made; the cause of the reported clipping was not
  in noodle's geometry. Fixture `landscape-90-playbook` and
  `TestLandscape_SideColumnFits_Issue90View` keep that view inside the canvas, and the
  canvas-bounds lint (T06) catches any element past it from now on (FR-009).
- [x] **T15** LNT-05 #1 and #5: the contract names the test that exists and states the
  sequence label after its `[1] ` step number (ADR-0011 numbers every message).

