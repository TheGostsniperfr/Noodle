# Progress

Handoff between sessions. The newest entry comes first. Update it at the end of every
session: what changed, what is open, where the next session starts. Keep entries short;
decisions go in ADRs, acceptance in `features.json`.

## Start here

1. `make init`: fix anything red before new work. Commands: `docs/COMMANDS.md`; process: `docs/PROCESS.md`.
2. Read the current spec: [`specs/003-static-discovery/spec.md`](../specs/003-static-discovery/spec.md)
   and ADR-0017, ADR-0018. Backlog: the [project board](https://github.com/users/TheGostsniperfr/projects/5), issue #42.
3. Next action: spec 003 T05 (secrets and references, Argo CD Applications), then T06,
   T12 and the T14 checkpoint before anything else ([`tasks.md`](../specs/003-static-discovery/tasks.md)).
   Phase 1 still has spec 002 T03 (issue #43, in `dockair-docs`).
   Measure on real repos with `make discover-eval`.
4. Spec 007 (process, #108) is done on `feat/spec-process`; PR pending. Then pilot the
   process on #79 (`noodle drift`, P6a-05) as spec 008, full mode.

## Open threads outside this repo

| Where | What | Status |
|---|---|---|
| `~/Documents/aepita/ing2/3-Istor/cnp-docs` | CNP runtime specs and diagrams (`architecture/`), the old generator `tools/archgen` and skill `.claude/skills/cnp-diagram` | **uncommitted**. The generator and skill are superseded by this plugin. Proposal: keep `architecture/specs` + diagrams, delete `tools/archgen` and the local skill, render with the plugin. |
| CNP platform | Security gaps found while mapping: G1 no Cilium tenant isolation in code, G2 no groups-claim authorization in SecurityPolicy, G3 plain HTTP in cluster including Vault (`tls_disable`), G4 tunnels bound to Envoy's hashed Service name, G5 one tunnel per app on `cloudflared:latest`, G6 token call probably hairpins through Cloudflare | documented in the diagram only. Deserve issues in the CNP repos. G1 and G2 first. |
| `~/.dotfiles` | noodle plugin installed declaratively (commit `9bf7e57`) | pushed. `nixos-rebuild switch` still to run. |

## Log

### 2026-11-05 · Spec 008 convergence

- LNT-01, 02, 03, 04, 06, 08 pass. T14: #90 reproduced from agent-playbook's real view on
  c2cf233 and does not clip (drawio, PNG and SVG exports); fixture added, no resolver
  change. T15: LNT-05 contract aligned with ADR-0011 step numbers. LNT-05 and 07 pass:
  spec 008 done, all eight features graded by the evaluator.

### 2026-11-05 · Lint robustness, convergence (spec 008)

- T13: the zone title obstacle includes the icon offset of a wrapped sub, so it equals the drawn block for icon zones.
- T12: LNT-05 sequence tests assert the thresholds' labels after the step number, with a real empty-text message over a bare connection.
- T11: LNT-01 #1 holds exactly: the 420 px legend now has one finding, the `link` row; step and port rows use the free room to the card edge. Notes and every drawn card row are covered by tests.
- T10: LNT-04 shared-port test now asserts 0 findings; two ports on the same point are reported.

### 2026-11-04 · Lint robustness (spec 008)

- T01: `house.Header` and `house.CardLayout` own the header, card and legend geometry; `render.go` draws from them, byte-equal on the frozen specs of `cmd/noodle/testdata/cardlayout`. Every example's `.drawio` is now committed and `TestExamples_DrawioCurrent` keeps it current.
- T02: `card_text.go` measures every card row, legend row and note against the cell the renderer gives it, and card height. Four example legends were too narrow (cnp-runtime, platform-overview, app-runtime-view, repo-map): widened, neighbours shifted. Flips of LNT-01, B-07, B-26 wait for the evaluator.
- T03: `header_overlap.go` reports a zone, node, card or note over the header block as drawn (`house.HeaderBox`). No example needed a fix. Flips of LNT-02 and B-25 wait for the evaluator.
- T04: `ports.go` compares the distinct badges of a node pairwise, 2 px inflated; `testdata/ports-73.yaml` is the report's repro. No example needed a fix. LNT-04 flip waits for the evaluator.
- T05: empty label plus badges renders the badges alone; views gain `hide_annotations` and `hidden_annotations` (schema, model, ADR-0026), and `model.Check` fails an annotation no view target draws. `examples/regression-78` is the fixture; oidc-login acknowledges G1 and G2. LNT-05 flip waits for the evaluator.
- T06: `canvas_bounds.go` reports a zone, node, card or note past any canvas edge. The #90 landscape does not overrun the canvas at this commit (fixture `landscape-90` reaches the width cap and stays inside), so the resolver is unchanged; see the report. LNT-07 flip waits for the evaluator.
- T07: node text is measured at `house.ExportCharWidthEm` (0.602) plus `house.FitsWithMargin`, in the lint, `resolve.boxWidth` and the landscape tile width. Fixture `examples/regression-72`. cnp-runtime, platform-overview and app-runtime-view boxes widened (nodes shifted to keep their edges); sequence boxes and landscape tiles grow by themselves. LNT-03 flip waits for the evaluator.
- T08: a zone sub that does not fit the title line inside the zone wraps under the label in a column of at most 240 px; `house.ZoneTitleBox` covers exactly those lines and the renderer draws them. Fixture `examples/regression-87` (rejected by the lint on main, so its before image comes from main's renderer with the lint off). No other example changed. LNT-06, B-20 flips wait for the evaluator.
- T09: every example whose drawing changed commits its `.drawio` and re-exported PNGs, with a before image from main's renderer in `specs/008-lint-robustness/visual/`. The fixes themselves landed with the rule that needed each (T02 legends, T05 oidc-login, T07 boxes). LNT-08 flip waits for the evaluator.
### 2026-11-04 · Discovery completion, spec 009 wave 1

- T01: optional `object {apiVersion, kind, namespace, name}` on discovered elements, connections and references (ADR-0028), schema and loader; the adapter fills it. Flip of DSC-04 waits for `make evaluate`.
- T02: references `env-from`, `secret-key`, `config-key`, `mounts` from workloads (one per workload, target and kind), ExternalSecret to target and store, stores to `k8s:_external/<provider>` elements (1Password, Vault, AWS; others `unknown-provider`), Argo CD `syncs`; no Secret value read. DSC-01 flip waits for `make evaluate`.
- T03: inferred connections from `<svc>.<ns>[.svc[.cluster.local]][:port]` in env values, ConfigMaps read by a workload (per key for `configMapKeyRef`) and custom resource specs, only to Services of the input; public URL hosts of HTTPRoutes and Ingresses to their Gateway or controller. Secrets are never scanned. DSC-02 flip waits for `make evaluate`.
- T04: stderr warning `N objects name no namespace; pass -namespace` on a mixed input, `-argocd-app FILE`, `-kustomize DIR [-enable-helm]` (fake `kustomize` in tests, fails naming `nix shell nixpkgs#kustomize`), `-o DIR/` names the fragment `<adapter>-<source>.yaml`, relative `src.file`, `kustomize` in the dev shell. OPEN: the `unrendered` hint is still the short code `kustomize`, not the rerun command (DSC-03 #3 `hint-without-flag`): the fragment schema restricts `hint` to `^[a-z0-9-]+$`, and relaxing it is a contract change ADR-0028 does not cover. Needs Brian.
- T05: `tools/tokencost` reads a Claude Code transcript (one turn per API message id) and prints turns, wall time, weighted tokens and dollars per model and per first tool; prices are an assumption per model family (opus 5, sonnet 3, haiku 1 $/M input), to confirm before the T09 checkpoint.
- T06: `examples/discovery-k8s` (one manifests file, `list.yaml` mirroring its lines, curated `model.yaml`, golden `discovered/k8s-manifests.yaml`), `make examples` diffs a rerun against the golden file, `TestDiscoverExample`, `TestExample_ManifestsEqualList`, `TestDiscover_Property_Deterministic`, `TestIdentity/every-item`. DSC-05, P6a-03 and P6a-04 flips wait for `make evaluate`.
- T07: `internal/graph` (`Merge`, `Lookup`, `Neighbours`, sources and objects kept per merged element and connection) and the command loader `loadSystem`, which render now uses; a discovered connection between matched ends is mapped onto curated ids, its port kept only if the curated end has it. DSC-06 flip waits for `make evaluate`.
- T08: `noodle drift DIR` (unmatched, gone, planned, unmodelled; sorted; exit 0; error without fragments), cross-fragment resolution of `missing-backend` for a Gateway or Service another fragment defines, `examples/discovery-k8s/drift.txt` diffed by `make examples`. Namespace zones are not reported as unmatched. DSC-07 and P6a-05 flips wait for `make evaluate`. Wave 1 ends here: T09 is the paid checkpoint, not started.
- T04 follow-up: the `unrendered` hint stays `kustomize`; `noodle discover` prints `<DIR>: kustomization found, rerun: noodle discover k8s -kustomize <DIR> (needs kustomize: nix shell nixpkgs#kustomize)` on stderr (subtest `hint-without-flag`, per the coordinator decision in ffa8c40).
- Convergence T18: `takes-list-whole` asserts the whole merged element (DSC-06 NEEDS_WORK #1).
- Convergence T19: the property test counts only fields set on both sides with different values, excluding ID, Kind and Matches, so its guard can fail (DSC-06 #4).
- Convergence T20: `keeps-sources` asserts through `Lookup("api")` (`Node.Sources`). DSC-06 ready for a new `make evaluate`.
### 2026-11-05 · Spec 010 convergence (PRJ-06)

- Evaluator NEEDS_WORK on PRJ-06: tasks T16 to T18 below.
- T16: VISION names architecture as code and living documentation in lowercase.
- T17: tutorial view drops its legend card (the generated Gx line was false for a view with no gaps); PNG re-rendered and looked at.

### 2026-11-04 · Spec 010 implementation (project and distribution)

- T01: Apache-2.0 `LICENSE`, `NOTICE`, and `license` in plugin.json, flake.nix and SKILL.md.
- T02: icon attribution page (45 rows, source and license) and `go test ./assets/` that fails per missing row, URL or license.
- T04: skill wording: Claude Code-only names moved under an `In Claude Code` section; the facts step names the `fact-finder` subagent.
- T03: `speccheck skills` (run by `all`) checks every SKILL.md against the Agent Skills rules; rules live in `internal/skillrules` so `noodle setup` tests can reuse them.
- T05: CONTRIBUTING, SECURITY, CODE_OF_CONDUCT (Contributor Covenant 2.1), idea and bug issue templates, README links.
- T06: `noodle version` prints `main.version`; the flake sets it from the flake revision.
- T08: Codex subagents page reread (required keys name, description, developer_instructions; sandbox_mode is ignored in agent files), recorded in plan.md; `content.go` embeds `skills` and `agents`; the wrapper hash covers them.
- T09: `noodle setup opencode|codex -project|-user` writes the skill to `.agents/skills/` and one agent file per host, idempotent, golden-file tests.
- T10: setup marker line on every written file (replaced on rerun, foreign files kept and exit 1); README install section per host: Claude Code, OpenCode, Codex, Nix.
- T11: user guide in `docs/guide/` (tutorial on `examples/tutorial`, four how-tos, explanation); `noodle help [-markdown]`, `make cli-doc`, `cli.md` and the test that keeps it current.
- T12: README positioning (architecture as code, "designed to stay in sync" until P6a-05), VISION "Where it sits", and the Discovery section of the skill reference (#88); goldens refreshed for the skill text.
- T13: goreleaser config (4 platforms, checksums), release workflow on `v*` tags, git-cliff config, `make changelog` and `make release-snapshot` (evaluator allowlist, recorded in ADR-0024), plugin `version` 0.1.0, first CHANGELOG.md. No tag pushed.
- T14: `bin/noodle` downloads the release of the plugin's version when no noodle, Go or Nix is found, verifies SHA-256 against the release checksums, refuses a mismatch; Go test with stub curl, uname, go, nix.
- T15: process skills `issue-writer`, `release-notes`, `contrib-check` in `.agents/skills/`, symlinked from `.claude/skills/`, checked by `make spec-check`.

### 2026-11-04 · Engineering process (spec 007, ADR-0024)

- Backlog: 15 issues created (#93 to #107), #16 and #18 rewritten, #95 and #96 sub-issues
  of #69, everything on the board; drafts in the artifact "Noodle B-42 to B-58".
- Process: constitution (C01 to C15), EARS specs in full or quick mode, MADR 4 ADRs with
  a generated index, contracts, a read-only `evaluator` agent, an evidence gate on
  `features.json`. `tools/speccheck` enforces it; `Makefile` is the only entry point,
  documented in `docs/COMMANDS.md`. `check-features.sh` ported to Go and removed;
  `scripts/init.sh` is now `exec make init` because old verify steps cite it.
- Contract reviewed in five rounds, all applied; reviews now capped at three (PROCESS.md).
- Every feature evaluated by the read-only evaluator; DEV-01 to DEV-05 pass. Its
  NEEDS_WORK verdicts found real bugs, fixed as Convergence tasks T06 to T14: contract
  criteria matched any mention of an id, an unfilled template verdict passed the gate, a
  misspelled BASE skipped the history rules, an Evaluator look-alike was accepted.
  One finding was wrong (set -e already stops a failed locate) and is recorded as such.
- PR #109 opened for spec 007 (closes #108).
- ADR-0025: the backlog is worked in thematic batches and waves, all in full mode, on
  `plan/wave-1`. Wave 1 specs being written in parallel worktrees: 008 lint robustness,
  009 discovery 6a completion (takes over spec 003's open tasks and #79), 010 project and
  distribution. Questionnaire answered; specs 008 to 010 agreed (ADR-0026 to ADR-0030).
- Stopped on the usage limit. Spec 010 (worktree agent-ab81ef575d825ffd7): implemented
  except T07; PRJ-01, 03, 04, 07, 08 PASS (evidence committed, not flipped yet), PRJ-05
  NEEDS_WORK (SECURITY private channel check) and PRJ-06 NEEDS_WORK (VISION terms in
  lower case). Spec 009 (agent-a555890757e8a7270): wave 1 T01 to T08 done; its 10
  evaluations were stopped mid-run, rerun them. Spec 008 (agent-ad00c897fbfc7dbf9): the
  Sonnet implementer was stopped mid-run, check its last commit before resuming.
  Note: drop Co-Authored-By from commits (Brian's git rule wins over the harness).

### 2026-10-25 · Multi-line node desc

- The lint measured `desc` as one line, so the CNP AI views (`desc` with `<br>`, drawn as
  several lines by draw.io) failed on width and height. `Node.Lines()` now splits `desc` on
  `<br>`; two lint tests lock it. Not a roadmap item: found while updating `cnp-docs`.

### 2026-10-05 · Badge-only step labels

- B-28: `labels: {id: ""}` already left a numbered edge with its badge; now without a
  trailing space, locked by a test and documented in the skill. Used on the DockAir
  messaging topology, where fifteen edges share one corridor.

### 2026-10-05 · Messaging primitives (spec 006)

- ADR-0023: `shape: pipe` for queues, topics and streams, and `routes` in the model for
  what a broker forwards (bindings, subscriptions, dead-letter links), drawn as their own
  edge kind. Asked by the DockAir RabbitMQ and Clock diagrams.
- The pipe is draw.io's horizontal cylinder (`direct_data`): `cylinder3` with a
  `direction` rotates entry and exit points, so edges landed on the wrong side.
- Spec and tasks 006, phase 1c in the roadmap; P1c-01 to P1c-04 pass. Every existing
  example renders byte-identical to main in both themes.
- `examples/messaging` in init.sh and CI. Spec 003 stays the next action.

### 2026-10-02 · Topology diff for proposals (spec 005)

- ADR-0021: `state: diff` on topology views and `status`/`target` on connections and
  references, inherited from their ends. Asked by Brian for the PAE observability
  proposal: what a plan adds and removes must stand out over the dimmed platform.
- Spec and tasks 005, phase 5a in the roadmap; features P5a-01 to P5a-04 pass.
- Added: change-coloured frame (matrix green), `+ target` pill, `+` label. Removed: red
  dotted frame, struck title, `−`. Unchanged: 30 % opacity, drawn first. Port badges
  stay lit when a changed edge lands on them. Legend lists the three in place of
  planned and deprecated.
- Every existing example renders byte-identical to main in both themes.
- `examples/topology-diff`: a proposal and its reference view over one model, in
  init.sh and CI. Next: use it for the PAE observability proposal.
- ADR-0022: `highlight: [ids]` on topology views, with include's patterns. Highlighted
  elements and every edge touching them stay in full, the rest is dimmed; a diff keeps
  its added and removed frames. Asked by Brian to show the whole observability domain of
  the PAE platform, existing and proposed. `examples/topology-diff` view `focus`, P5a-05
  passes; existing examples still byte-identical to main. Not committed yet.

### 2026-10-02 · Access matrix (spec 004)

- ADR-0019 and ADR-0020: grants in the model, one `matrix` view, built now on its own
  traversal. The access graph and lenses of #53 are set aside on `feat/access-views`,
  after the CNP owner found the graph unreadable and the matrix clear.
- Spec, plan and tasks 004; features P6b-01 to P6b-05. Spec 003 stays the current work.
- T01: memberships, grants, auth; view type `matrix` with `identities`, `resources`,
  `state`; `target` allowed on deprecated items (the phase that removes them) (P6b-01).
- T02: `internal/access`: groups, grants, levels per state, grant ids, phase of a
  change (P6b-02).
- T03: `resolve.Matrix`, `diagram.Matrix`, matrix text lint; cells carry level, change
  (added, changed, removed, rescoped), phase and counts (P6b-03).
- T04: render cells, frames, tags, hatches, counts and legend; level and change
  colours in both themes, AA contrast tested. On the CNP model: rescoped only when no
  grant at the top level survives, a grant's own phase before its membership's (P6b-04).
- T05: `examples/cnp-access` (the CNP model) in init.sh and CI; skill reference, SKILL
  recipe and README (P6b-05). Phase 6b features pass; it closes when cnp-docs publishes
  the matrix.
- Review: column groups framed down to the last row like zones, uppercase group titles,
  legend in a card like a topology's.

### 2026-10-01 · Spec 003 blind spots and checkpoint

- ADR-0018 and spec 003 update from the PAE recall check: `secret-endpoint` (T15),
  operator rules (T16), the fragment as a floor in the skill (T11), and a checkpoint
  (T14, P6a-08) after T05, T06 and T12: discovery continues only if the fragment run
  costs clearly less and finds as many edges as the agent alone.

### 2026-09-30 · Token cost of a diagram, fact-finder agent

- `noodle discover -namespace NS`: objects that name none land there, as `kubectl apply
  -n` and an Argo CD destination do. `generic-api` sets no namespace on its Deployment
  and Service: without it prod and preprod collided in `default`.
- Recall check against Brian's PAE `platform-overview` (31 edges, 25 inside the cluster),
  rendered as Argo CD does (release name, destination namespace, three value files):
  T04 finds 5 (the Gateway routes); T05/T06 should add ~6 (Grafana datasources, ESO to
  1Password, Argo CD to GitLab, cert-manager issuers, OIDC URL). Blind spot: ~9 edges
  are credential-implied (DB, S3, SMTP addresses live in 1Password, only secret key
  names are in Git), and 4 come from operator CRs or controller args (Prometheus to
  Alertmanager, external-dns `--provider=aws`, image registries, SecurityPolicy OIDC).
- Found in PAE: `sre/{gatus,longhorn,prometheus-stack}/prod/charts/secure-route-0.1.0.tgz`
  still say `gateway-infra`, fixed in the source by `c486f8b` (2026-05-29): if Argo CD
  uses the committed tgz, those routes attach to no Gateway.

- Spec 003 T04: Gateway (listener ports), HTTPRoute, GRPCRoute, TLSRoute, TCPRoute and
  Ingress as elements (desc: hosts), IngressClass under `k8s:_cluster/`. Connections:
  Gateway or IngressClass to Service on the Service port name, Service to each selected
  workload per target port. References: parentRef, backendRef, Ingress backend and class.
  `unresolved`: `unmatched-selector` (operator-made pods: CNPG, Prometheus) and
  `missing-backend` for a Service or a Gateway not in the input (hint `service`,
  `gateway`, `not-a-service`). Edge ids are a 10-hex hash: readable ids repeating from and
  to cost ~20 % of the fragment.
- Measured, prod: PAE argocd 93 elements, 36 connections, ~9.8k tokens; DockAir 74
  elements, 22 connections, ~8.3k tokens. DockAir's 20 `missing-backend` are real: its
  routes render from the overlays while the apps come from `generic-app-chart` through
  Argo CD, so the rendered input lacks them. `discover-eval.sh` now passes relative paths,
  as users do: absolute temp paths in every `src` inflated the count by ~10 %.

- ADR-0017 proposed: yaml.v3 decoding without apimachinery (keeps `src` lines),
  deterministic host resolvers, `unresolved` in the fragment with token guards, agent
  answers remembered as rules. Accepted after review.
- Spec 003 T03: `internal/adapter/k8s` reads files, directories, stdin, multi-document,
  `List` and JSON with yaml.v3; a repeated key keeps its last value (a rendered Keycloak
  chart has one); skips owned objects, generated kinds, Helm test and delete hooks, CRDs
  and duplicates; walks skip hidden directories and name Helm charts and Kustomize
  directories as `unrendered`. Elements: workloads (tech with the short image),
  Services, namespaces as zones; no title, the id holds the name. Unknown kinds: one
  `unresolved` line per kind with a count. `Stats` and a summary line on stderr.
- `scripts/discover-eval.sh` renders real repositories on a copy and compares. Prod,
  inventory only (connections come in T04): PAE argocd ~277k tokens of manifests
  without CRD schemas to a ~6.8k fragment, 87 elements; majoutes ~35k to ~0.8k; DockAir
  ~78k to ~3.4k. The input is rendered output, larger than the sources an agent reads:
  T13 measures against the agent. Findings: majoutes charts set no namespace (Argo CD
  does, T05 fixes it); useful unknown kinds are operator CRs (CNPG Cluster, Prometheus,
  ImageUpdater, SecurityPolicy).

- Spec 003 T02 (P6a-02): `internal/adapter` (interface, explicit `Registry`, `Source`,
  `Run` stamps provenance, sorts by id and rejects ids without the adapter prefix);
  `model.EncodeFragment` writes one flow item per line; `noodle discover ADAPTER
  [PATH…|-] [-o] [-ref] [-observed-at]`. No adapter registered yet: k8s comes in T03.

- Spec 003 T01 (P6a-01): `kind: Fragment` contract and schema, `matches` on model
  elements, loader reads `discovered/`. Choices within ADR-0015: discovered id
  `<adapter>:<path>`; `src` is `{file, line}` or `{object}`; connections carry
  `inferred` for FR-005 heuristics; fragment elements take no curated presentation
  (status, shape, color) nor `matches`. A discovered id is matched by one element at most.
- Backlog moved from Notion to the [GitHub project board](https://github.com/users/TheGostsniperfr/projects/5),
  referenced in AGENTS.md, README, the SessionStart hook and the PR template (`Closes #`).
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
