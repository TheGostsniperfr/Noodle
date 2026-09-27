# ADR-0011 · A sequence step is a message over a model connection

- Status: Accepted
- Date: 2026-09-27
- Refines: ADR-0004 and ADR-0007 (view `steps`)

## Context

ADR-0007 made view `steps` a list of connection ids. That fits a topology, where each
connection carries at most one step (ADR-0004). It does not fit a sequence: messages
travel both ways over one connection. A response goes back to the client, and through a
tunnel the edge sends requests to the connector that opened it. Spec 001 (FR-008) also
asks for dashed responses in sequence views.

## Decision

In a topology view, `steps` stays a list of connection ids. In a sequence view, each step
is one of:

```yaml
- {id: get, from: user, to: cf-edge, over: e-1-browser-edge, text: "GET app"}
- {reply: get, text: "200"}
- {note: envoy-shared-gateway, text: "no session cookie"}
```

- A **message** names `from`, `to` and `over`, a model connection whose two ends are
  `from` and `to`, in either direction. A sequence cannot invent a path: a message with
  no connection means the model is incomplete. `from` and `to` restate what `over`
  implies so the file reads as a story; `model.Check` rejects any mismatch.
- A **reply** points at an earlier message by `id` and travels back over the same
  connection. It is drawn dashed. Going against a connection's direction is not a
  reply: a request relayed through a tunnel is solid.
- A **note** is something a participant does on its own. It is not a message, so there
  are no arrows from a participant to itself.
- `text` defaults to the connection's `verb · protocol`. Messages and replies are
  numbered 1..n; notes are not.
- Participants are the elements in order of first appearance, unless `participants`
  fixes the order.

Alternatives, loops and parallel blocks are phase 4 and would be new step keys.

## Consequences

- A connection renamed or removed in the model breaks every sequence that uses it, in
  `model.Check`, not in review.
- Sequence files are more verbose than a list of ids. They are read more than written,
  and every repeated fact is checked.
