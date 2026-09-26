# Noodle · Roadmap

Each phase ends with something usable. Order can change, but a phase never starts
before the contracts it depends on are frozen.

| Phase | Level | Deliverable | Done when |
|---|---|---|---|
| **0** ✅ | · | Repository, vision, architecture, ADRs, v0 generator and skill migrated from the CNP experiment; Claude Code plugin + marketplace, Nix package | `/plugin install noodle@noodle` works on a fresh machine |
| **1** | L0 | Contracts `v1alpha1`: model, view, layout as separate files, with JSON Schemas; lint rules moved into `internal/lint`; CNP runtime example migrated | one model, two views (nominal path, OIDC), lint green |
| **2** | L0 | Own SVG renderer from the same geometry; animated flow dashes via CSS, which also work when the SVG is shown as an image | the example renders to SVG without draw.io and animates in a README |
| **2b** | L6 | Read-only MCP server over the model: list elements, find a path, explain a connection, validate | an agent answers "how does a request reach app X?" from the model alone |
| **3** | L1 | Web viewer: scenario selector, focus and dim, tags as overlays, search, element docs, export of the current view | the example is explorable in a browser and exports what is on screen |
| **4** | L2 | Flows with IcePanel-style step types, playback, auto-play tours, self-contained HTML export, public redaction profile | a 60-second auto-played tour of CNP on a static page |
| **5** | L3 | Model diff, visual diff colours, per-PR static preview, GitHub Action | a PR shows added and removed elements through a link |
| **6** | L4 | Adapters: manifests (Helm/Kustomize), `kubectl` read-only, Hubble flows, Terraform state; provenance; drift report; auto-layout with ELK | an unknown cluster yields a draft map in minutes, with a drift report |
| **7** | L5 | Live overlays: Argo CD health, alerts, Hubble drops; deep links; path connectivity checks | a failing pod turns red on the map with links to logs and metrics |
| **8** | L6 | MCP with live context, wired into HolmesGPT or Claude; improvement proposals as model patches | an incident investigation starts from the topology |

## Why this order

- **Contracts before features.** Multi-path views, the AI database, discovery and PR
  diffs all need model, view and layout to be separate. That is phase 1.
- **SVG renderer before the viewer.** The viewer is that renderer plus interaction. It
  also removes draw.io from the critical path.
- **Read-only MCP early (2b).** Once the model is clean it costs little, and it gives AI
  agents value long before the live cockpit.
- **Discovery after the viewer.** Discovered models need auto-layout and a way to look
  at them. Both exist by then.
- **Live and actions last.** They need infrastructure access, storage and an
  authorization model.

## Next up: phase 1

1. Write ADR-0007: file layout of model, views and layouts, and how they reference each other.
2. JSON Schemas in `schemas/v1alpha1/`.
3. Split `examples/cnp-runtime` into `model.yaml`, `views/runtime.yaml`, `layouts/runtime.yaml`.
4. Move lint rules into a package with one rule per file and table-driven tests.
5. Add the OIDC flow as a second view, drawn as a sequence (ADR-0004).
