# ADR-0003 · An arrow is a connection

- Status: Accepted
- Date: 2026-09-26

## Context

Architecture diagrams mix requests, responses, data flow and configuration links in the
same kind of arrow, which makes them useless for network or security review.

## Decision

- An arrow is a **connection**, from the side that opens it (client) to the side that
  listens (server). Firewall and NetworkPolicy reasoning reads straight off the diagram.
- **Requests only.** A response that changes the path (302, 401, 500, retry) is a note
  attached to the component that sends it.
- The **listening port** is a badge straddling the server's border. Client ports are
  ephemeral and not drawn. The label says `verb · protocol`.
- **Egress leaves on the right or bottom, ingress enters on the left or top.** The only
  exception is `against_flow`, for outbound connections that run against the reading
  direction: tunnels, agents calling home, hairpins.
- **References** (parentRef, targetRef, envFrom, secret name) are a separate dotted kind.
  They are not traffic.

## Consequences

- Traffic carried back over a connection opened the other way, such as a tunnel, is
  explained in the label, not drawn as a reverse arrow.
- The lint can check sides, ports and kinds mechanically.
