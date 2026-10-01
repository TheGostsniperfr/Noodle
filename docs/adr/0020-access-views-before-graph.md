# ADR-0020 · Access views are built now, on their own traversal

- Status: Accepted
- Date: 2026-10-01
- Amends: ADR-0019 (dependency on `internal/graph`), the phase order in `docs/ROADMAP.md`

## Context

ADR-0019 placed phase 6b after spec 003 T09, because the access view walks a graph and
`internal/graph` was planned there. The CNP segregation plan needs the before and after
pictures now, while spec 003 is mid-way (T05 next, a checkpoint at T14 that may stop
further discovery work). Waiting couples a finished design to an unfinished one.

Looking closer, the access view needs much less than T09: no merge, no fragments, no
drift. It needs reachability over memberships and grants, a few dozen edges per system.

## Decision

- **Phase 6b starts now**, alongside 6a. Spec 003 stays the current work for discovery;
  6b touches model, resolve and render only, never the adapters.
- **`internal/access` owns access semantics**: effective access of a subject, who
  reaches a resource, derived escalations. It reads `model.Model` and keeps its own
  adjacency lists. When `internal/graph` lands, its traversal may replace the local one
  behind the same functions; nothing outside `internal/access` changes.
- **Lens entries carry an id**: `lenses: [{id, subject, state}]`, so `-lens <id>` picks
  one and two lenses on the same subject (current, target) never clash.
- **The access view lays itself out**: layers by longest path, pass-through slots for
  edges that skip a layer, one lane per edge in each corridor, ports spread on node
  sides. The existing lint runs unchanged on the result, so no rule is relaxed.

## Consequences

- Two traversals may exist for a while. The duplication is a few functions and is
  removed when T09 is done; spec 003's plan gets a note.
- The roadmap lists 6b after 6a with both open.
- A maintainer reviewing 6b sees no change to adapters, fragments or existing views:
  every new field is optional and existing examples render byte for byte the same.
