# ADR-0004 · Topology and sequence are separate views

- Status: Accepted
- Date: 2026-09-26

## Context

Numbering an OIDC login on the CNP topology put several numbers on one arrow and the
same number on several arrows. Three options were considered: decimal hop numbers
(4.1, 4.2), duplicated parallel arrows for reused connections, and splitting topology
from sequence. Decimals still leave several numbers on one arrow. Duplicated arrows break
ADR-0003, since one connection would be drawn twice.

## Decision

- A **topology** view numbers only its nominal path: **one step per connection, each
  step used once**. The lint enforces it.
- Flows that reuse connections or go back and forth (auth, token exchange, retries,
  sagas) are **sequence** views. Their connections stay on the topology, unnumbered, with
  a label pointing to the sequence.
- Numbers `[1]…[n]` are the request path (data plane). Letters `[A]…` are background
  work (control plane).

## Consequences

- Every number on a topology is findable in one glance.
- The renderer needs a sequence layout (phase 1 onwards). IcePanel-style step types
  (alternate, parallel, go-to-flow) live in views, not in the model.
