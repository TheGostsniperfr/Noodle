# ADR-0023 · Queues are pipes, broker routing is a route

- Status: Accepted
- Date: 2026-10-05
- Extends: ADR-0003, ADR-0007, ADR-0008

## Context

The first message broker drawn with noodle (DockAir: RabbitMQ exchanges, queues and the
simulation clock) hit two gaps in the vocabulary.

- **A queue looks like everything else.** Exchanges and queues are both `kind: bus`. A
  queue stores messages in transit and is owned by its consumer; an exchange stores
  nothing and only routes. The only way to tell them apart was the icon, and a queue
  drawn as a `cylinder` reads as a database.
- **Routing inside the broker had no edge type.** A binding from an exchange to a queue
  is not a connection: no process opens it, there is no port (ADR-0003). Drawn as a
  reference it is grey and dotted, "object reference, not traffic", which hides that
  every message travels along it. The binding key, the part readers need, ended up as a
  free-text `kind`.

## Decision

- **`shape: pipe`** on elements: a cylinder lying down, open end on the right, for a
  queue, a topic or a stream. It keeps the element's kind and colour. An exchange or a
  router stays a box: it holds nothing. Drawn with draw.io's horizontal cylinder
  (`mxgraph.flowchart.direct_data`), not `cylinder3` with a `direction`, because draw.io
  rotates entry and exit points with the direction. The legend explains the pipe when
  one is drawn. The open end takes a fifth of the width; the text lint accounts for it.
- **`routes:` in the model**: `{id, from, to, kind, match, status, target}`. A route is a
  path the broker forwards messages along: a binding, a subscription, a dead-letter link.
  `kind` names the mechanism, `match` what selects the messages (a binding key, a
  filter, a condition). The label reads `kind · match`. Both ends are components, never
  zones. `status` and `target` follow ADR-0021.
- **Drawn as its own edge kind**, `route`: the bus colour, dash-dot, filled arrow.
  Legend: "broker routing, not a connection". Like references, a route ignores the
  egress and ingress sides: it is not a connection.
- **A route is never a step.** Steps number connections (ADR-0004). The hop through the
  broker is told in the steps card.

Rejected: a connection kind `route` (breaks ADR-0003: no client, no port); a known
reference kind drawn differently (magic strings in a free field, and `match` would
stay text); a `shape: hexagon` for exchanges (a box with the product icon already reads
as a service that routes); port badges inherited from the broker zone (one badge per
exchange and queue on the same 5672 is noise: the zone `sub` says it once).

## Consequences

- Additive to `v1alpha1`: existing systems render byte-identical.
- A discovery adapter can later emit routes from `Binding` or `Subscription` objects.
- Sequence views still message over connections only. A message over a route would need
  a decision of its own, when a diagram asks for it.
