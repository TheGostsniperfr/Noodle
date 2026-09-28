# ADR-0012 · Platform offerings live in the model; a catalog view shows them

- Status: Accepted
- Date: 2026-09-28
- Extends: ADR-0007 (model content, view types)

## Context

A platform team sells a service catalogue: a database per app, an app as a service,
dependency updates, hardened base images. Each offering has a promise ("one PostgreSQL
per app, primary + replica"), a way to request it ("`database.enabled: true` in
app-configs") and components behind it. Today this lives in prose docs and slides. It is
exactly what an agent needs to answer "how do I get a database?", and it references
elements the model already has.

## Decision

- **Model: a top-level `offerings` list.** Not `services`, which reads as a Kubernetes
  Service.

  ```yaml
  offerings:
    - id: dbaas
      title: Database as a Service
      icon: cnpg
      summary: One PostgreSQL per app, primary + replica
      provides: [RW and RO endpoints, credentials injected, BI replica]
      request: "database.enabled: true in app-configs"
      backed_by: [cnpg-operator, vault]
      owner: socle
      status: planned            # optional, ADR-0008 values, target from ADR-0013
  ```

  `backed_by` ids must exist in the model. `owner` is free text.
- **A new view type `catalog`**: `include` lists offering ids in display order,
  `columns` sets the grid (2 or 3). No layout file: the grid is computed.
- **Rendering.** One card per offering: icon and title, summary, the `provides` list,
  the `request` line in monospace, a row of `backed_by` logos taken from the model.
  A planned offering is hatched like a planned element.

## Consequences

- The catalogue cannot promise a component that is not in the model: the reference check
  catches it.
- Offerings are not elements: they have no geometry in topology views and no connections.
  A later topology overlay could highlight what backs one offering.
- `offerings` and `catalog` are additions to `v1alpha1`: no version bump.
