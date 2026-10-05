# Spec 006 · Messaging primitives

- Status: **Done** (`tasks.md`)
- Phase: 1c, an addition to the L0 vocabulary asked by the first message broker diagram
- Decisions it builds on: ADR-0003, ADR-0006, ADR-0007, ADR-0021, ADR-0023

## Why

A message broker diagram must show where messages wait and how the broker routes them.
Today a queue and an exchange look the same, and a binding is drawn as a grey object
reference although every message travels along it.

## User scenario

**Platform engineer documenting a RabbitMQ topology.** Declares exchanges as boxes and
queues with `shape: pipe`, then one route per binding with its key in `match`, and the
dead-letter link of each queue. Applications keep their connections to the exchanges
(publish) and to their queues (consume). The reader sees at once who opens what, where
messages wait, and which key brings them there.

## Requirements

- **FR-001** `shape: pipe` accepted by the checks and the model schema.
- **FR-002** `routes` in the model with `id, from, to, kind, match, status, target`:
  ids unique across the model, ends existing components, status rules of ADR-0021,
  never a step. Schema updated.
- **FR-003** The resolver turns a route into a `route` edge labelled `kind · match`.
- **FR-004** The renderer draws the pipe and the route edge in both themes and explains
  each in the legend when it is used. The text lint accounts for the pipe's open end.
- **FR-005** Existing examples render byte-identical. `examples/messaging` renders in
  `init.sh` and CI.

## Out of scope

Messages over a route in sequence views. Routes from discovered `Binding` objects.
Ports inherited from a zone.
