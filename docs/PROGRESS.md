# Progress

Handoff between sessions. The newest entry comes first. Update it at the end of every
session: what changed, what is open, where the next session starts. Keep entries short;
decisions go in ADRs, acceptance in `features.json`.

## Start here

1. `./scripts/init.sh`: fix anything red before new work.
2. Read the current spec: [`specs/001-model-view-layout-split/spec.md`](../specs/001-model-view-layout-split/spec.md).
3. Next action: spec 003, task T01 in [`tasks.md`](../specs/003-static-discovery/tasks.md).
   Phases 1 and 1b are closed.

## Open threads outside this repo

| Where | What | Status |
|---|---|---|
| `~/Documents/aepita/ing2/3-Istor/cnp-docs` | CNP diagrams as one noodle system `architecture/cnp`: platform overview, app runtime view, request path, OIDC login; archgen and the local skill removed | branch `feat/add-noodle-schema` pushed (1bfed51), for Brian to review. |
| CNP platform | Security gaps found while mapping: G1 no Cilium tenant isolation in code, G2 no groups-claim authorization in SecurityPolicy, G3 plain HTTP in cluster including Vault (`tls_disable`), G4 tunnels bound to Envoy's hashed Service name, G5 one tunnel per app on `cloudflared:latest`, G6 token call probably hairpins through Cloudflare | documented in the diagram only. Deserve issues in the CNP repos. G1 and G2 first. |
| CNP platform | **G7**: the CMP `/api` route has no OIDC policy and the backend decodes JWTs with `verify_signature`, `verify_aud` and `verify_exp` off (`CMP/backend/app/services/keycloak_service.py:476`, `routers/account.py:45`): a forged token is accepted | found 2026-09-30, in the diagram only. Deserves an issue first. |
| `~/.dotfiles` | noodle plugin installed declaratively (commit `9bf7e57`) | pushed. `nixos-rebuild switch` still to run. |

## Log

### 2026-09-30 · Token cost of a diagram, fact-finder agent

- Measured the PAE diagram session (29/09, 137 turns, noodle dev excluded), cost weighted
  at cache read ×0.1, cache write ×1.25, output ×5: discovery ~41 % (50 `cat`/`grep`
  turns, each rereading the context), geometry ~20 % (45 % of diagram output tokens),
  lint and render ~11 %, PNG review ~2 %, fixed overhead ~21 %.
- Static discovery (phase 6 adapters) would save ~35-40 %, plus auto-layout ~60-70 %.
- Agreed with Brian: spec 003 (phase 6a) drafted with ADR-0015 (fragments, `matches`,
  facets reserved, files now and a graph DB only when L5 history or scale needs it) and
  ADR-0016 (k8s discovery before the viewer, own Go adapter on k8s libraries; existing
  tools map k8s objects, not client-to-server connections). Auto-layout is the next spec.
- Spec 001 T13: CNP runtime layout on named lanes (tunnels, connectors, tenant bus,
  secrets) with `@px` exits; draw.io output byte-identical in both themes.
- Spec 001 T16: `noodle migrate SPEC DIR` converts a v0 file to a v1alpha1 system and
  refuses to write unless it resolves to the same geometry. The three platform-regression
  fixtures migrated, draw.io byte-identical in both themes; v0 inputs kept in
  `internal/migrate/testdata`. Agreed with Brian: T17 keeps the v0 reader behind
  `migrate` only, for PAE, DockAir and cnp-docs.
- Spec 001 T17 and spec 002 T13: the CLI renders systems only (a `.yaml` argument points
  to `noodle migrate`); CNP and docs v0 files moved to `internal/migrate/testdata`, docs
  diagram migrated (byte-identical). Skill `reference.md` rewritten for `v1alpha1`
  (model, four view types, layout with lanes, a minimal system rendered and looked at),
  `SKILL.md` workflow on `noodle render`, recipes for tech stack and catalogue; README
  status, usage and example list. The minimal example exposed two lint gaps: B-25
  (header overlap) and B-26 (card taller than its box).
- `bin/noodle` rebuild hash now covers `internal/`: plugin users had a stale binary after
  any change there.
- B-06: `multiplicity` renders as two dashed copies 10 px up and right behind the box;
  the lint keeps their footprint clear (`house.NodeFootprint`), the legend lists each
  text. `repo-map` fixture stacked. Unblocks deleting DockAir's post-processor (002 T03).
- Spec 002 T03: dockair-docs !20 merged. Its three diagrams are v1alpha1 systems,
  `postprocess-drawio.py` and its side files deleted, PNGs compressed with pngquant
  (`PNG_QUALITY`): draw.io 30.2.6 exports PNGs 3 to 4 times larger than before.
  Phases 1 and 1b closed. Still on v0: PAE and cnp-docs (`noodle migrate` when needed).
- Benchmark brief in Notion: "Benchmark · Token cost of a diagram", to paste in the PR.
- fact-finder trial on PAE: right on routes, GitOps and secrets; missed the edge forwarder
  and credential-implied connections, prompt fixed. Its "no Image Updater" was the old
  system still on disk; Brian's diagram shows the target, a `planned` case.
- Stopgap shipped: `agents/fact-finder.md` (Sonnet, read-only) runs skill step 1 in its
  own context and returns a `v1alpha1`-shaped YAML inventory with a source on every fact.

### 2026-09-28 · Spec 002 drafted: landscape and catalog views

- Brian agreed to a phase 1b before phase 1 closes: tech stack (landscape) and service
  catalogue (catalog) views, both derived from the model, computed grid, no layout file.
- ADR-0012 (landscape view, element kind `tool`), ADR-0013 (`offerings` in the model,
  catalog view), ADR-0014 (`target` on planned items), accepted by Brian. Spec, plan and tasks in `specs/002-landscape-and-catalog/`, features
  P1b-01 to P1b-07. First consumer: the DockAir Sprint 0 deck.
- Order after acceptance: T01 to T13 of spec 002 (planned rendering first, B-05), then the
  sequence tasks T14, T15 of spec 001 with a DockAir delivery sequence, then T13, T16, T17.
- Branch `feat/landscape-catalog-views`.
- T01: `target` in the model, `System.Status` inherits status and target from zones,
  both carried into the resolved diagram; check rejects a target on a non-planned element.
- T02: `planned` renders hatched and dashed with a target pill on the top border, `deprecated`
  dotted with the title struck, zones too; legend lines for the statuses used. Checked on the
  DockAir platform diagram, dark and light. P1b-01, P1b-02 and B-05 pass.
- T03 deferred until after T12: the DockAir script also draws `.stacked` copies (B-06), so it
  cannot be deleted yet.
- T04: element kind `tool` (external palette), view type `landscape` with `bands`, `side`,
  `width`; checks for unknown, zone or duplicate items and element labels.
- T05: `resolve.Landscape` and `resolve.View` (dispatch by type). Items are plain nodes, not
  a new tile shape, so the node text and spacing lint apply as is. Sections stretch to fill
  their line with tiles centred; width fits the widest band (max 2400 px) unless set;
  sections inherit the band colour; flow arrows are `diagram.Arrow`, no semantics.
- T06: lint findings carry `Warn`; warnings print as `warn:` and never block. First one: a
  landscape item without icon.
- T07: `examples/landscape` (bands, side column, flow, planned with and without target, a
  tool), rendered dark and light, in `init.sh` and CI. P1b-03 to P1b-05 pass.
- T08: `offerings` in the model (ids share the model id space), view type `catalog` with
  `include` of offering ids and `columns` 1 to 4; checks on `backed_by`, status, target.
- T09: `resolve.Catalog`, `diagram.Offering` with text pre-wrapped by `house.Wrap` so resolver,
  lint and renderer agree; card: icon and title, summary, you get, how to request (tinted
  box), backed-by logos; rows share the tallest height. Lint: offering text overflow, overlap.
- T10: `examples/catalog` with a planned offering, dark and light, in `init.sh` and CI. P1b-06 passes.
- T11: 20 logos in `assets/icons` (simple-icons with brand colour, CNCF artwork, project repos for
  dagster, loki, grafana-alloy cropped to their mark); `.dark` variants for sqlalchemy and trivy.
  Not found: checkov (text tile, lint warning), alertmanager (no official mark, the model uses
  the prometheus logo).
- Planned boxes and cards: hatch at 50 % and text at 75 %, after the diagram-reviewer found
  17 planned tiles hard to read at slide scale.
- T12: DockAir system in `dockair-docs/docs/02-architecture/systems/dockair` (47 elements, 9
  offerings), views `tech-stack` (width 2900) and `service-catalog` (3 columns). Reviewer's one
  blocking finding fixed and checked live: app databases run PostgreSQL 18.4, BI 17.7. Other
  fixes: no flow arrows on delivery (the order was not true), collaboration moved to the
  developer platform, TLS at the SiOps edge. P1b-07 passes. Not committed in dockair-docs.
- `-slide` render flag (both CLI forms): drawing only, no header, cards or notes, canvas
  cropped to the content. Lint still runs on the full diagram. Asked by Brian: pasted in a
  slide, the full export wasted most of the space on frame and legend.
- Next: T13 (skill), then T03 once B-06 lands, then spec 001 T14 (sequence).

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
  P1-03 and P1-04 pass.
- T11: lint rules in `internal/lint`, one per file, each with a failing case against a
  clean spec; house style (themes, metrics, badge and label boxes) in `internal/house`.
  draw.io output byte-identical on every spec; lint messages identical, in the same order,
  on two deliberately broken specs (27 findings). P1-05 passes.
- ADR-0011: a sequence step is a message (`from`, `to`, `over` a model connection, either
  direction), a `reply` to an earlier message (dashed), or a `note`. A message against
  its connection needs a text. Gaps show on participants and messages as in topologies.
- T14, T15: sequence layout computed (columns widen until each label fits its span) and
  drawn with the topology's boxes and palette. Micro example `examples/sequence-basics`;
  CNP OIDC login as `views/oidc-login.yaml`, 28 messages over the runtime connections.
  P1-06 passes: phase 1's "one model, two views" holds. T16, T17 remain.

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
