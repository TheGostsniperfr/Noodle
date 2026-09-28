# Plan 002 · Landscape and catalog views

Spec: [`spec.md`](spec.md). Decisions: ADR-0011 (landscape, `tool`), ADR-0012
(offerings, catalog), ADR-0013 (`target`).

## Approach

Same shape as plan 001: new inputs, same drawing. Each new view type gets a resolver
that emits the resolved diagram (`internal/diagram.Spec`), so the geometry lint, the
draw.io renderer, both themes and the icon set apply unchanged. What is genuinely new is
small: a computed grid layout, one node shape (`tile`), one zone style (`band`) and one
card style (`offering`).

```
model.yaml ─┬─ views/tech-stack.yaml (type: landscape) ─▶ resolve.Landscape ─┐
            └─ views/catalog.yaml   (type: catalog)    ─▶ resolve.Catalog   ─┴▶ diagram.Spec ─▶ lint ─▶ render
                                            no layouts/ file: geometry computed
```

## Packages

| Package | Change |
|---|---|
| `internal/model` | `Element.Target`; kind `tool`; `Model.Offerings`; `View.Bands`, `View.Side`, `View.Columns`; checks: landscape items exist and are unique, `backed_by` ids exist, `target` requires `planned`, catalog `include` names offerings |
| `schemas/v1alpha1` | model: `tool`, `target`, `offerings`; view: `type` gains `landscape` and `catalog`, `bands`, `side`, `columns` |
| `internal/diagram` | `Node.Status`, `Node.Target`, `Zone.Status`, `Zone.Style` (`band`); `Card.Offering` fields (summary, provides, request, logos) |
| `internal/house` | tile and band metrics (logo size, title line, band header width), hatch pattern per theme |
| `internal/resolve` | `landscape.go`, `catalog.go`, `status.go` (zone status and target inheritance, shared with topology) |
| `internal/lint` | `node_icon.go` (warning); card text overflow covers offering cards (part of B-07) |
| `cmd/noodle` | render `status` and `target` (B-05), `tile` shape, `band` zone, `offering` card; `render` accepts the two new view types |

## Grid layout, computed

- **Landscape.** Canvas width fixed (default 2000 px, view field `width` optional).
  The side column takes 280 px when present. A band is a row: header block 240 px, then
  sections left to right, each as wide as its item count allows, wrapped to a new row
  inside the band when the sum exceeds the width. A tile is 170 × 64 px: logo 36 px left,
  title and tech right. Gaps follow ADR-0006: 24 px between boxes, 60 px between zones.
- **Catalog.** `columns` equal cards; a card's height comes from its content (the same
  text metrics the lint uses), and all cards in a row share the tallest height.

Heights are derived, so no `h: auto` contract is needed (B-08 stays open for topology).

## Risks

- **Planned rendering parity with DockAir's script.** The script is the reference for
  FR-009; PNG compared with `magick compare`, small anti-aliasing diffs judged by eye.
- **Icons.** The DockAir stack needs about 25 logos the set lacks. The `icon-curator`
  agent fetches them: generic products into `assets/icons`, DockAir-only into its
  `.noodle/icons`.
- **Band header text.** Long band subtitles overflow at 240 px; the node-text lint rule
  already reports it, so the example must pass it rather than widen the header.
