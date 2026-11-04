# Plan 008 · Lint robustness

Spec: [`spec.md`](spec.md). Decisions: ADR-0006, ADR-0025; clarifications 1 to 5 of the spec.

## Approach

Every fix is either a new lint rule in its own file (ADR-0006: one rule per file) or a
layout correction in the resolver or `internal/house`. No existing rule loses strength
(C04). The lint can only measure what it can compute, so the two places where the
renderer computes geometry on its own move to `internal/house`, which both import (C06):

```
resolve ──▶ diagram.Spec ──┬─▶ lint      (rules read house geometry)
                           └─▶ render    (draws the same house geometry)
house: NodeFootprint, ZoneTitleBox, PortBadges, + HeaderBox, CardLayout (new)
```

- **Header** (FR-003): `house.HeaderBox(spec)` returns the rectangle the renderer's
  `header()` draws today (dot at 40, title at y 30, sub and meta from y 74, measured line
  by line). New rule `header_overlap.go`.
- **Cards, notes, legend** (FR-001, FR-002): `house.CardLayout(spec, card)` returns the
  rows the renderer draws (title, listed lines with their bullet, or the legend's two
  columns of rows and its notes block), each with its rectangle and font size. The
  renderer's `card()` and `legend()` draw from it; new rule `card_text.go` measures it.
- **Text margin** (FR-004): one helper, `house.FitsWithMargin(textW, availW, fontSize)`, used by
  `node_text.go` and by `resolve.boxWidth`. The measuring width follows clarification 1.
- **Port badges** (FR-005): `ports.go` compares the distinct badges of each node pairwise.
- **Annotations** (FR-006, FR-007): `resolve.edges` joins label and badges with a
  non-empty join; the resolver collects annotations with no drawn target and returns them
  for the form chosen in clarification 2.
- **Zone sub** (FR-008): `house.ZoneTitleBox` and the renderer's zone label follow
  clarification 3 together, so obstacle and drawing stay one computation.
- **Canvas bounds** (FR-009): new rule `canvas_bounds.go`; `resolve/landscape.go`
  computes the band width from the canvas minus the side column instead of capping the
  natural width first.

## Packages

| Package | Change |
|---|---|
| `internal/house` | `HeaderBox`, `CardLayout`, `FitsWithMargin(textW, availW, fontSize)`; `ZoneTitleBox` per clarification 3 |
| `internal/lint` | new `header_overlap.go`, `card_text.go`, `canvas_bounds.go`; `ports.go` pairwise; `node_text.go` margin; registered in `rules` |
| `internal/resolve` | `sequence.go` box width with margin; `resolve.go` badge join and undrawn annotations; `landscape.go` side column |
| `cmd/noodle` | `render.go` header, card, legend and zone label drawn from `house` |
| `examples/`, `docs/diagrams/` | diagrams the new rules reject are fixed (FR-010); fixtures `regression-72`, `regression-78`, `regression-87` with committed `.drawio` and PNGs |
| `cmd/noodle` tests | `TestRender_CardLayoutGolden` on frozen copies in `testdata/cardlayout/`; `TestExamples_DrawioCurrent` re-renders every committed fixture `.drawio` |

## Alternatives set aside

- Checking text overflow on the exported PNG (OCR or pixel diff): slow, needs draw.io in
  the loop, and arrives after the lint instead of in it.
- Per-rule opt-outs in views: would let a diagram pass by switching a rule off (C04).

## Risks

| Risk | Mitigation |
|---|---|
| The margin (FR-004) rejects existing examples | FR-010: fix the examples in the same task, show before/after (C13) |
| Moving card and legend layout changes the rendered output | T01 is a pure refactor: every example's `.drawio` byte-identical to `main` |
| Batch B touches `resolve.go` at the same time | changes here are local to `edges` and `sequence.go`; rebase on wave order |
