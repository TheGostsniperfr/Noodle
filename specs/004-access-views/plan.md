# Plan 004 · Access graph and lenses

Spec: [`spec.md`](spec.md). Decisions: ADR-0019 (grants, views, styles), ADR-0020
(built now, `internal/access`, lens ids, computed layout).

## Approach

Same shape as plan 002: new inputs, same drawing. The access view gets a resolver that
emits `diagram.Spec`, so the lint, both themes, the icon set and the draw.io renderer
apply unchanged. Lenses decorate the spec a topology view already resolves to.

```
model.yaml (memberships, grants) ─▶ internal/access ─┬─▶ resolve.Access ─▶ diagram.Spec ─▶ lint ─▶ render
                                                     └─▶ resolve.Lens (topology spec + levels)
```

Existing systems are untouched: no field is required, and a system without grants
renders byte for byte as before.

## Packages

| Package | Change |
|---|---|
| `internal/model` | `Membership`, `Grant`, `Auth`; `View.Focus`, `View.State`, `View.Lenses`; view type `access` (computed); checks of FR-001 and the view fields |
| `schemas/v1alpha1` | model: `memberships`, `grants`; view: type `access`, `focus`, `state`, `lenses` |
| `internal/access` | `Graph` from a model and a state; `Reach(subject)`, `Reachers(resource)`, `Escalations()`; levels ordered read < write < admin; break-glass apart |
| `internal/diagram` | `Node.Level`, `Node.Dim`, `Node.Removed`; `Zone.Level`, `Zone.Dim`; edge kinds `member`, `grant-read`, `grant-write`, `grant-admin`, `grant-breakglass`, `escalation` |
| `internal/house` | edge styles and legends for the new kinds, `Dim` colours per theme, level badge geometry; a test that dim text keeps 3:1 on the background |
| `internal/resolve` | `access.go` (layers, slots, lanes, ports, cards), `lens.go` |
| `cmd/noodle` | render the new edge kinds, hollow diamond, level badge, dim and removed styles; `-lens`; access legend without steps and ports |

## Access view layout, computed

- **Graph.** Nodes are the elements met while walking from the focus. Edges: membership
  (subject to group), grant through a mechanism (subject to each `via`, then `via` to the
  resource), direct grant (subject to resource), escalation (derived, subject to the
  final resource).
- **Layers.** Longest path from the sources. An edge that skips layers gets a
  pass-through slot in each layer it crosses, so it never runs through a box.
- **Order in a layer.** Two barycentre sweeps, then ties by id, so a run is deterministic.
- **Corridors.** Between two layers, each edge gets its own vertical lane (12 px apart),
  then runs horizontally into its target. Exits and entries are spread on the node's
  right and left sides, 26 px apart, and the node grows to fit them. Labels sit on the
  last horizontal segment, clear of every lane.
- **Cards** under the graph: legend, reach summary, escalation paths.

## Risks

- **Crossings.** A barycentre heuristic does not minimise them; draw.io bridges them and
  the lint accepts perpendicular crossings. Collinear runs are prevented by lanes.
- **Label length.** A long scope (`write · project-*, namespace a`) needs a wide
  corridor; its width is computed from the longest label in it.
- **Cycles.** A membership cycle is a check error; a grant cycle is reported by the
  resolver with the elements involved.
