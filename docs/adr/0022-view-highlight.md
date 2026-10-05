# ADR-0022 · A topology view can highlight one domain

- Status: Accepted
- Date: 2026-10-02
- Extends: ADR-0021

## Context

ADR-0021 dims everything a proposal leaves unchanged. Reviewing an observability
proposal on the whole PAE platform, the reader also wants the existing observability
(Prometheus, Grafana, Alertmanager, Gatus) in full next to what the proposal adds, and the
rest of the platform as context. A diff alone cannot say that: those boxes are unchanged.
The vision lists the same need as "focus" for the viewer (phase 3).

## Decision

- **`highlight: [ids]` on topology views**, with the patterns of `include` (`zone/**` for
  a subtree). It is a reading choice, so it lives in the view, not in the model.
- An element in the highlight is drawn in full. An edge is in full when either end is
  highlighted: the flows into and out of a domain are part of it. Everything else is
  dimmed, at ADR-0021's 30 %.
- **With `state: diff`**, added and removed keep their frame wherever they are; the
  highlight only decides what else stays in full. Without a diff, planned and deprecated
  keep their ADR-0008 style inside the highlight.
- The legend reads "outside the highlight, dimmed" for the faded sample.

## Consequences

- One layout serves a reference view, a diff and a focused diff of the same platform.
- Phase 3's interactive focus can set this field from a click.
- `v1alpha1` addition, no version bump.
