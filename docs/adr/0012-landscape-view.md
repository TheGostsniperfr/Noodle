# ADR-0012 · A landscape view shows the technology stack

- Status: Accepted
- Date: 2026-09-28
- Extends: ADR-0007 (view types), ADR-0009 (`type`)

## Context

A platform team presents its technology stack as a grid of logos grouped by concern:
runtime, delivery, platform services, governance, with a side column for cross-cutting
concerns such as observability. The DockAir team drew one with an image generator: wrong
logos, no house style, and it drifts from what is deployed. Every product on it already
is, or should be, an element of the architecture model. Some are not runtime components
(Terraform, Renovate, Linear): the model has no kind for them.

## Decision

- **A new element kind `tool`** for things the system is built or run with but that
  receive no traffic: IaC, scanners, bots, planning and chat tools. Tools may have no
  parent and no connections.
- **A new view type `landscape`.** It selects elements and arranges them, it never
  declares a technology:

  ```yaml
  apiVersion: noodle/v1alpha1
  kind: View
  type: landscape
  title: DockAir tech cartography
  bands:
    - id: delivery
      title: Delivery & Operations
      sub: From code to running platform
      color: emerald
      icon: rocket
      flow: true                      # arrows between sections, left to right
      sections:
        - {title: Provisioning, items: [terraform, openstack]}
        - {title: Delivery (GitOps), items: [argocd, argocd-image-updater]}
  side:
    - {title: Observability, color: orange, items: [alloy, loki, prometheus, grafana]}
  labels: {velero: "Velero (Backup & Restore)"}
  ```

  Title, icon, tech and status come from the model; `labels` overrides the text shown.
- **No layout file.** Like sequence views, the geometry is computed: bands stacked in
  order, sections sized by their item count, items wrapped in rows, the side column at a
  fixed width. Positions are not worth reviewing in Git for a grid.
- **Lint.** An item that is not a model element, or appears twice in one view, is an
  error. An item without an icon is a warning. The geometry rules of ADR-0006 apply to
  the resolved grid unchanged.
- `planned` items render as ADR-0008 says, with their target (ADR-0014).

## Consequences

- The stack is derived from the model: an element renamed or retired disappears from the
  landscape with no second edit, and the MCP server can answer "which tool scans images?".
- The resolver emits the same resolved diagram as topology views: zones for bands and
  sections, nodes of a new `tile` shape for items. The renderer, lint and themes are reused.
- `tool` and `landscape` are additions to `v1alpha1`: existing files stay valid, so no
  version bump.
