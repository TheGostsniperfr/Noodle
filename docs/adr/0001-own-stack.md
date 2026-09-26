# ADR-0001 · Build our own stack instead of adopting LikeC4

- Status: Accepted
- Date: 2026-09-26

## Context

LikeC4 covers much of the vision: model/view split, dynamic views with a sequence mode,
an interactive viewer, an MCP server, draw.io round-trip. IcePanel and Ilograph cover the
presentation side. The v0 experiment on the CNP platform showed that what we value most
is precise control over semantics and rendering: port badges on borders, egress/ingress
sides, one step per connection, bridges on crossings, a lint that refuses unreadable
output. Later levels need discovery from infra, drift reports and live overlays.

## Decision

Build noodle's own model, lint, layout and renderers. Borrow ideas explicitly from
IcePanel (flows, tags, drafts), Ilograph (perspectives, focus, HTML export) and LikeC4
(flow-control blocks, MCP over the model). See `docs/INSPIRATIONS.md`.

## Consequences

- Full control over look and semantics; every rule can be enforced.
- Months of work before the viewer matches LikeC4's interactivity. Phases are ordered
  so each one is usable on its own.
- We own auto-layout integration (ELK) and the viewer.
