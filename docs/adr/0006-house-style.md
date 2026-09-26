# ADR-0006 · House style enforced by lint

- Status: Accepted
- Date: 2026-09-26

## Context

AI-generated diagrams tend to be inconsistent and cluttered: overlapping labels, lines
through boxes, arbitrary colours. Readability has to be guaranteed by construction, not
by review.

## Decision

- Palettes are Tailwind CSS v3, dark (400 strokes on slate-950, from Cocoon-AI) and
  light (600 strokes on 50 fills). Text pairs meet WCAG AA. Colour is never the only
  carrier of meaning: icon, dash pattern and legend repeat it.
- Node text follows the C4 triptych: name, [technology], one-line responsibility.
- Official logos for products (CNCF artwork, simple-icons), Kubernetes community icons
  for K8s objects, a generic CRD icon otherwise, as Argo CD does.
- Zones carry an environment or product icon next to their title.
- The lint rejects: a line through a box, label, zone title or port badge; overlapping
  labels; collinear edges; boxes closer than 24 px; zones closer than 60 px; text
  overflow; wrong egress/ingress sides; more than one step on a connection or a reused
  step.
- Crossings are drawn with bridges.

## Consequences

- An AI agent can generate diagrams and rely on the lint to catch layout mistakes.
- Style changes happen in one place, the theme, and apply to every diagram.
