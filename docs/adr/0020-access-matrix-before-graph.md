# ADR-0020 · The access matrix is built now, on its own traversal

- Status: Accepted
- Date: 2026-10-02
- Amends: the phase order in `docs/ROADMAP.md`

## Context

The access matrix walks memberships to find what an identity reaches. A graph package
is planned in spec 003 (T09, `internal/graph`), mid-way through phase 6a. The CNP
segregation plan needs its matrix now, and waiting couples a small, finished design to
an unfinished one.

The matrix needs little from a graph: the groups an identity belongs to, transitively,
and the grants they hold. No merge, no fragments, no drift.

## Decision

- **Phase 6b starts now**, alongside 6a. It touches model, resolve, lint and render only,
  never the adapters.
- **`internal/access` owns access semantics**: groups of a subject, its grants, its
  strongest level per resource in a state. It keeps its own adjacency lists. When
  `internal/graph` lands, its traversal may replace the local one behind the same
  functions.
- **The matrix lays itself out**, like landscape and catalog: no layout file.

## Consequences

- Two traversals may exist for a while; spec 003's T09 carries a note to merge them.
- Existing views and examples render byte for byte the same: every new field is
  optional.
