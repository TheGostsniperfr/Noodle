# Progress

Handoff between sessions. The newest entry comes first. Update it at the end of every
session: what changed, what is open, where the next session starts. Keep entries short;
decisions go in ADRs, acceptance in `features.json`.

## Start here

1. `./scripts/init.sh`: fix anything red before new work.
2. Read the current spec: [`specs/001-model-view-layout-split/spec.md`](../specs/001-model-view-layout-split/spec.md).
3. Next action: task T11 in [`tasks.md`](../specs/001-model-view-layout-split/tasks.md),
   then the tasks in order. Sonnet is enough: the decisions are in ADR-0007.

## Open threads outside this repo

| Where | What | Status |
|---|---|---|
| `~/Documents/aepita/ing2/3-Istor/cnp-docs` | CNP runtime specs and diagrams (`architecture/`), the old generator `tools/archgen` and skill `.claude/skills/cnp-diagram` | **uncommitted**. The generator and skill are superseded by this plugin. Proposal: keep `architecture/specs` + diagrams, delete `tools/archgen` and the local skill, render with the plugin. |
| CNP platform | Security gaps found while mapping: G1 no Cilium tenant isolation in code, G2 no groups-claim authorization in SecurityPolicy, G3 plain HTTP in cluster including Vault (`tls_disable`), G4 tunnels bound to Envoy's hashed Service name, G5 one tunnel per app on `cloudflared:latest`, G6 token call probably hairpins through Cloudflare | documented in the diagram only. Deserve issues in the CNP repos. G1 and G2 first. |
| `~/.dotfiles` | noodle plugin installed declaratively (commit `9bf7e57`) | pushed. `nixos-rebuild switch` still to run. |

## Log

### 2026-09-27 · ADR-0007 accepted, phase 1 planned

- Spec 001 open questions settled with Brian: one directory per system (`model.yaml`,
  `views/`, `layouts/`), sequence views computed without a layout, `denied` connections
  in the model, layouts with relative positions, symbolic endpoints and named lanes.
- `plan.md` (resolver in front of the v0 renderer) and `tasks.md` (T01 to T17). P1-01 passes.
- PRs that change the rendering now show a before/after image (AGENTS.md, PR template).
  First docs diagram, drawn with noodle: `docs/diagrams/contract-pipeline`, linted in CI.
- T01: geometry types moved to `internal/diagram`; draw.io output byte-identical on all
  five specs, both themes.
- ADR-0009: a view's perspective is `type`, since `kind` already names the file type.
- T02: `internal/model` types and `LoadSystem(dir)`, strict YAML, header checked first.
- T03: `model.Check` reports file, id and reason for every broken reference or invalid
  value (FR-005); one failing table case per rule.
- T04: JSON Schemas in `schemas/v1alpha1`, validated in `go test` with
  santhosh-tekuri/jsonschema v6 (vendored). P1-02 passes.
- T04b: CNP example written in the split format next to the v0 file, checked by schemas
  and `model.Check`. It forced two contract changes: endpoints line up with their
  neighbour by default and accept `@NNpx` (percentages alone gave 91.667 %), and
  annotations may target connections (G1, G6).
- ADR-0010: `denied` is the intent, `enforced_by` names the elements that enforce it.
  CNP's `x-cross-tenant` has none, which is G1. Lint rule for it: backlog B-24.
- T05 to T09: `internal/resolve` turns model + view + layout into the v0 geometry:
  relative positions, endpoints aligned with their neighbour or offset, lanes, step
  numbers, derived labels and badges. An aligned endpoint that lands beside its box is
  an error. `model.Check` also checks routes exist and run from `from` to `to`.
- T10, T12: `noodle render DIR [-view ID]`. The split CNP example renders pixel for pixel
  like v0 (`magick compare -metric AE` = 0, both themes), linted by init.sh and CI.
  P1-03 and P1-04 pass. Next: P1-05 (lint into `internal/lint`), then the OIDC sequence.

### 2026-09-27 · Field feedback from a second platform

- Three diagrams built with the plugin on a second real platform (platform overview,
  application runtime view, GitLab repository map). Findings turned into backlog B-04 to
  B-18 and ADR-0008: node status and multiplicity.
- Fixed ligatures: labels disable `font-variant-ligatures`, so `<app>-frontend` no longer
  renders as `<app>—frontend` (B-04, testify added as test dependency).
- Skill: grouping and naming rules, check-before-drawing rules, two recipes (application
  runtime view, repository map), all from Brian's review of those diagrams.
- ADR-0008 accepted after review: both fields live in the model, `multiplicity` is one
  text field (`one per team`), and they ship in the `v1alpha1` schemas (P1-02).
- Spec 001: B-10, B-11, B-12 (symbolic endpoints, named lanes, zone-relative layout) are
  layout-contract questions for ADR-0007, not backlog extras.
- Regression fixtures: the three diagrams, anonymized, in `examples/platform-regression`,
  linted by `init.sh` and CI (B-22). Backlog B-19 to B-21: `via` edges, zone sub
  wrapping, label placement off vertical segments. B-23: tiled PNG preview, so agent review sees
  what the downsampled full image hides.

### 2026-09-26 · Session harness automated

- Hooks in `.claude/settings.json`: `SessionStart` injects the handoff context, `Stop` blocks
  ending a turn while changes are missing from this file, `PostToolUse` runs gofmt.
- `scripts/check-features.sh` (in `init.sh` and CI) rejects rewriting a feature's history.
- CI (`.github/workflows/ci.yml`): vet, test, lint the example, features history, vendor sync.
- Plugin agents: `diagram-reviewer` (Opus, read-only), `icon-curator` (Haiku).
- `wrap-up` skill for end of session. Model guidance in AGENTS.md.
- Decided against project permission rules (deny/ask/allow): not useful for a solo repo in
  auto mode. Revisit when someone else contributes.

### 2026-09-26 · Bootstrap

- Built the v0 generator and house style on the CNP platform: Tailwind palettes (dark and
  light), official logos, port badges, bridges, lint for readability and semantics.
- Iterated semantics with Brian's feedback into ADR-0003 (arrow = connection, requests
  only, ports on the server, egress/ingress sides) and ADR-0004 (topology vs sequence, one
  step per connection).
- Chose to build our own stack over LikeC4 (ADR-0001). IcePanel and Ilograph are the
  main feature inspirations (`docs/INSPIRATIONS.md`).
- Wrote the vision (levels L0 to L6), architecture with contract sketches, roadmap.
- Shipped as a Claude Code plugin + marketplace (repo root), with a wrapper that finds or
  builds the binary, icons compiled in, a Nix package. Install tested from GitHub.
- Adopted the session harness: this file, `docs/features.json`, `scripts/init.sh`, and
  spec-first work in `specs/`.
