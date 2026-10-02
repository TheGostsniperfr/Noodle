# Plan 004 · Access matrix

Spec: [`spec.md`](spec.md). Decisions: ADR-0019 (grants, matrix), ADR-0020 (built now,
`internal/access`).

## Approach

Same shape as plan 002: a new input and a computed grid, drawn by the same renderer.

```
model.yaml (memberships, grants) ─▶ internal/access ─▶ resolve.Matrix ─▶ diagram.Matrix ─▶ lint ─▶ render
views/<id>.yaml (type: matrix)  ─────────────────────┘
```

## Packages

| Package | Change |
|---|---|
| `internal/model` | `Membership`, `Grant`, `Auth`; view type `matrix` (computed), `Rows`, `Columns` (reusing `Section`), `State`; checks of FR-001 |
| `schemas/v1alpha1` | model: `memberships`, `grants`; view: type `matrix`, `rows`, `columns`, `state` |
| `internal/access` | `Graph` per state: `Groups`, `Grants`, `Levels`, `GrantIDs` |
| `internal/diagram` | `Matrix`: header groups, column and row headers, cells (level, before, change, phase, note), counts, legend |
| `internal/house` | matrix metrics; level and change colours per theme |
| `internal/resolve` | `matrix.go`: grid from the view, cells from two `access.Graph`s |
| `internal/lint` | `matrix_text.go`: header, cell and legend text fits its box |
| `cmd/noodle` | render the matrix: cells, frames, tags, hatches, legend |

## Grid, computed

Row header 260 px, cells 104 × 58 px, 4 px apart; a 16 px gap between column groups and
between row groups; group header rows above. The count column is 140 px. The legend is
two rows of sample cells under the grid. Width and height follow from the counts.

## Risks

- **Header text.** Long titles wrap on two lines through `labels` (`<br>`); the lint
  reports any overflow.
- **Hatch fills.** draw.io's hatch is the one used for `planned` boxes; the break-glass
  and removed cells reuse it with their own colours.
