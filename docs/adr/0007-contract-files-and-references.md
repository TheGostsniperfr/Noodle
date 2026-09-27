# ADR-0007 · Contract files and how they reference each other

- Status: Accepted
- Date: 2026-09-27

## Context

ADR-0002 splits the v0 spec into Model, View and Layout, but leaves open where the files
live, what goes in which, and how they point at each other. Field use on a second
platform added one constraint: absolute edge points were the main cost of every layout
change (backlog B-10, B-11, B-12). Freezing `v1alpha1` on absolute coordinates would
break it again soon after.

## Decision

**A system is a directory.** `model.yaml` (kind `Model`), `views/<id>.yaml` (kind `View`),
`layouts/<view-id>.yaml` (kind `Layout`). A view's model is the `model.yaml` of its
directory: no `model:` field, no reference across directories. File stems are ids.
Every file carries `apiVersion: noodle/v1alpha1` and `kind`.

**Model: what exists.** No coordinates, no step numbers.
- `elements`: `id, kind, parent, title, tech, desc, icon, shape, ports, tags, status,
  multiplicity` (ADR-0008). Zones are elements of kind `region` or `group`, with `color`
  and `sub`.
- `connections`: `id, from, to, port` (a port name on the target), `protocol, verb,
  kind: flow | auth | tunnel | async`, and `denied: true` for a connection a policy
  refuses. A refusal is a fact the MCP server must see, so it is not a view overlay.
- `references`: `id, from, to, kind` (`parentRef`, `envFrom`, …). They replace v0 `link`
  edges (ADR-0003).
- `annotations`: gaps with `targets`. Element badges derive from them.

**View: what is shown.** `kind: topology | sequence`, title, subtitle, meta, `include`
(ids, or `zone/**` for a subtree), `steps` (ordered connection ids: numbers for the
request path, letters for background work, ADR-0004), label overrides per connection,
cards, notes. Sequence views may list `participants` to fix column order.

**Layout: where it is drawn.** Topology views only. Sequence views are computed from
step order: participants in order of first appearance, one row per step.
- `canvas: {width, height}`
- `lanes: {name: {x} | {y}}`, reusable in edge routes.
- `elements: {id: {x, y, w, h}}`, positions relative to the parent zone.
- `edges: {id: {from, to, waypoints, label_at, label_offset, against_flow}}`. Endpoints
  are `element.side[@ratio]` (`envoy.left`, `app.right@60%`); a waypoint is `[x, y]` or
  `lane:<name>`. `against_flow` is a drawing exception (ADR-0003), so it lives here.
- `cards` and `notes` positions for the view's cards and notes.

**References are checked.** An id in a view or layout that the model lacks is a lint
error naming the file and the id. A topology layout that omits an included element is
an error.

Rejected: a single multi-document YAML (large diffs, adapters rewrite mixed files);
layouts beside views (regenerated geometry mixed with reviewed files); `blocked` as a
view overlay; absolute-only coordinates; row/column auto layout in phase 1 (a layout
engine, which is ELK's job in phase 6).

## Consequences

- Reviewers read `model.yaml` and `views/`; `layouts/` churns without hiding facts.
- A discovery adapter writes `model.yaml` only.
- Sequence views need no layout, so their lint is structural only.
- `blocked` and `link` leave the edge kind vocabulary. B-10, B-11 and relative positions
  are phase 1 work; B-12 keeps only row/column layout.
