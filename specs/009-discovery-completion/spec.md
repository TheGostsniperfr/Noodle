# Spec 009 · Discovery, phase 6a completion

- Status: **Draft**, five questions open (Clarifications)
- Mode: **full**
- Phase: 6a, the current phase (ADR-0016), batch B of wave 1 (ADR-0025). Takes over the
  open tasks T05 to T16 of spec 003, which stays as written (C10).
- Issue: #42 · Batch items: #75, #76, #79, #80, #81, #16, #25
- Decisions it builds on: ADR-0003, ADR-0005, ADR-0010, ADR-0015, ADR-0016, ADR-0017,
  ADR-0018, ADR-0024, ADR-0025
- Invariants it touches: C02, C06, C07, C08, C09, C12

## Why

Phase 6a promised that an agent drawing a platform reads a fragment instead of the
manifests. On the first real use (Arffornia field report, 2026-10-31) the fragment held
8 connections, all Service to its own Deployment, where the runtime view needed about
30: the rest lived in ConfigMaps, env values and operator CRs. `matches` was validated
but unused, so the session wrote a 15-line drift script. Objects without
`metadata.namespace` landed in `default` and broke `matches` on the next run. Every
Kustomize directory was reported `unrendered`. And no reference carries enough identity
to jump from a node back to the YAML document that defines it.

## User scenarios

1. **Platform engineer, first diagram of a GitOps repository.** Runs `noodle discover
   k8s` on rendered manifests, gets workloads, Services, routes, and the connections
   that config values imply, each with the file and line that proves it. The skill reads
   the fragment, asks only about the `unresolved` entries, and writes the curated model.
2. **Same engineer, a month later.** Reruns discovery; `noodle drift` lists the workload
   nobody drew, the matched Deployment that is gone, and the connection discovery found
   but the model lacks.
3. **Reviewer asking "who reads this Secret?".** Runs `noodle query <system> neighbours
   k8s:prod/secret/db -kind secret-key` and gets every workload, with no agent.
4. **Team with an in-house CRD.** Writes a rule for `VaultSecret` once; the next
   discovery turns its objects into elements and references instead of `unknown-kind`.

## Requirements

### Wave 1 · References, addresses, inputs

- **FR-001** WHEN a workload reads configuration through `envFrom`, `secretKeyRef`,
  `configMapKeyRef`, or a Secret, ConfigMap or PersistentVolumeClaim volume, THE SYSTEM
  SHALL emit one reference per use, of kind `env-from`, `secret-key`, `config-key` or
  `mounts`, with its `src`, and SHALL never write a Secret value into the fragment.
  *Scenario:* WHEN Deployment `api` mounts Secret `db` and reads key `PASSWORD` from it
  THEN the fragment holds `mounts` and `secret-key` references from `k8s:prod/deployment/api`
  to `k8s:prod/secret/db`, and no line of the fragment contains the key's value.
- **FR-002** WHEN the input holds an ExternalSecret, a SecretStore or ClusterSecretStore,
  or an Argo CD Application, THE SYSTEM SHALL emit references ExternalSecret to its target
  Secret and to its store, store to an `external` provider element (1Password, Vault,
  AWS), and Application to its destination namespace (`syncs`).
  *Scenario:* WHEN ExternalSecret `db` names store `onepassword` THEN the fragment has
  references `db → k8s:prod/secret/db`, `db → k8s:prod/secretstore/onepassword`, and that
  store references an external element whose tech is `1Password`.
- **FR-003** WHEN a string value of a ConfigMap, an env var or a custom resource spec
  holds `<svc>.<ns>.svc[.cluster.local][:port]` or `<svc>.<ns>[:port]`, optionally with a
  scheme, naming a Service of the input, THE SYSTEM SHALL emit a connection from the
  reading workload to that Service, marked `inferred`, on the matching Service port, with
  the `src` of the value.
  *Scenario:* WHEN Deployment `web` has env `DB_HOST=postgres.data.svc:5432` and Service
  `postgres` in `data` listens on 5432 THEN the fragment holds an inferred connection
  `web → postgres` on port 5432 citing the env line.
- **FR-004** WHEN such a value holds a URL whose host matches a `hostnames` entry of an
  HTTPRoute or an Ingress host of the input, THE SYSTEM SHALL emit an inferred connection
  from the reading workload to the Gateway or Ingress controller that serves that host.
  *Scenario:* WHEN `api` has env `PUBLIC_URL=https://shop.example.com/` and an HTTPRoute
  on Gateway `shared` lists `shop.example.com` THEN an inferred connection `api → shared`
  is emitted, which shows the hairpin through the edge.
- **FR-005** IF an input mixes objects that name a namespace with objects that name
  none, and neither `-namespace` nor `-argocd-app` is given, THEN THE SYSTEM SHALL print to stderr
  `N objects name no namespace; pass -namespace`; and WHERE `-argocd-app FILE` is given,
  THE SYSTEM SHALL take the default namespace from that Application's
  `spec.destination.namespace`.
  *Scenario:* WHEN a Helm output has 3 objects without namespace and 2 in `web` THEN
  discovery prints `3 objects name no namespace; pass -namespace`; WHEN rerun with
  `-argocd-app app.yaml` naming `web` THEN the 3 objects get ids under `k8s:web/`.
- **FR-006** WHERE `-kustomize DIR` is given and `kustomize` is on PATH, THE SYSTEM SHALL
  run `kustomize build DIR` (with `--enable-helm` when `-enable-helm` is given) without
  writing to `DIR`, discover its output, and name the fragment after the directory; IF
  `kustomize` is absent THEN THE SYSTEM SHALL keep reporting the directory as
  `unrendered` with its hint. (Option of question 2.)
  *Scenario:* WHEN `noodle discover k8s -kustomize apps/shop` runs with kustomize on PATH
  THEN the fragment `k8s-shop.yaml` holds the rendered objects and `git status` of
  `apps/shop` is unchanged.
- **FR-007** THE SYSTEM SHALL write every `src.file` relative to the directory discovery
  was run from, and never an absolute or temporary path.
  *Scenario:* WHEN discovery reads a rendered temp file for `-kustomize apps/shop` THEN
  every `src.file` of the fragment starts with `apps/shop/` or names the Kustomization,
  and none starts with `/`.
- **FR-008** THE SYSTEM SHALL give every discovered item a top-level `object` identity
  `{apiVersion, kind, namespace, name}`, required `apiVersion`, `kind` and `name`, beside
  its `src`, whose existing string `src.object` stays valid and unchanged; and every
  reference the field path that created it, `field`
  (`spec.template.spec.volumes[2].secret.secretName`). (Option of question 3.)
  *Scenario:* WHEN the Deployment `api` mounts Secret `db` through its third volume THEN
  its `mounts` reference has `field: spec.template.spec.volumes[2].secret.secretName` and
  the element `k8s:prod/deployment/api` has `object: {apiVersion: apps/v1, kind:
  Deployment, namespace: prod, name: api}`.

### Wave 1 · Merge, drift, example, cost

- **FR-009** THE SYSTEM SHALL ship `examples/discovery-k8s` as planned in spec 003 (a
  gateway, two apps with a route each, a database reached through a Secret, an
  ExternalSecret, a default-deny NetworkPolicy with one allow, a curated model with
  `matches` and one planned element), whose golden fragment discovery reproduces byte
  for byte, from the manifests and from the same objects wrapped as a `kubectl get -o
  yaml` List.
  *Scenario:* WHEN discovery runs twice on `examples/discovery-k8s/manifests` and once on
  `list.yaml` THEN the three outputs equal `discovered/k8s-manifests.yaml`.
- **FR-010** WHEN a command loads a system with fragments, THE SYSTEM SHALL merge them in
  `internal/graph` per ADR-0015: fill fields the curated element leaves empty, never
  override one it sets, map discovered connections between matched elements onto curated
  ids, and keep the `src` and `object` of every matched item on the merged element.
  *Scenario:* WHEN curated `api` sets no ports and matches `k8s:prod/service/api`
  listening on 8080 THEN the merged `api` lists port 8080 and its sources include the
  Service's file and line; WHEN curated `api` sets tech `Go` THEN the merge keeps `Go`.
- **FR-011** WHEN `noodle drift DIR` runs, THE SYSTEM SHALL report, sorted and with their
  `src`: discovered objects no curated element matches, matched ids no fragment contains
  any more, `planned` elements still absent, and discovered connections between matched
  elements that the model lacks; and SHALL exit 0, drift being a report (ADR-0015).
  *Scenario:* WHEN Deployment `worker` is in the manifests but matched by nothing THEN
  `noodle drift examples/discovery-k8s` lists it under unmatched with its manifest line,
  and the whole report equals `examples/discovery-k8s/drift.txt`.
- **FR-012** WHEN a reference's target is in another fragment of the same `discovered/`
  directory, THE SYSTEM SHALL resolve it there and not report it as `missing-backend`.
  *Scenario:* WHEN `k8s-apps.yaml` holds an HTTPRoute whose parent Gateway is in
  `k8s-gateway-infra.yaml` THEN the merged graph links them and drift reports no
  `missing-backend` for that route.
- **FR-013** THE SYSTEM SHALL provide `tools/tokencost`, which reads a Claude Code
  transcript and prints its weighted cost per tool category and per model, its turns and
  wall time.
  *Scenario:* WHEN `go run ./tools/tokencost` reads the trimmed fixture transcript THEN
  it prints the totals stored in the fixture's expected output.

### Wave 2 · Checkpoint

- **FR-014** WHEN FR-001, FR-003 and FR-013 are done, THE SYSTEM SHALL be measured on
  the PAE `platform-overview` in two fresh sessions, agent alone (arm A) and agent given
  the fragment (arm C), and the cost, the recall against the 25 in-cluster edges and the
  go or stop decision SHALL be recorded in `docs/PROGRESS.md` and
  `docs/benchmarks/discovery-checkpoint.md` (ADR-0018). (Who runs it: question 1.)
  *Scenario:* WHEN both arms have run THEN the benchmark file lists, per arm, cost in
  dollars, edges found out of 25, and one line `Decision: go` or `Decision: stop`.

### Wave 3 · Only after a recorded "go"

- **FR-015** WHILE the checkpoint has not recorded `Decision: go`, THE SYSTEM SHALL keep
  FR-016 to FR-020 unimplemented; WHEN it records `stop`, THE SYSTEM SHALL revise this
  spec instead, under a `## Revision` heading (ADR-0018).
  *Scenario:* WHEN `docs/benchmarks/discovery-checkpoint.md` says `Decision: stop` THEN
  no commit after it touches the files of FR-016 to FR-020, and the spec gains a revision
  note.
- **FR-016** WHEN a NetworkPolicy is in force, THE SYSTEM SHALL resolve its ingress and
  egress rules to allowed pairs, and mark `denied` with `enforced_by` the policy any
  connection of FR-003 or of the routes that the policy blocks (ADR-0010).
  *Scenario:* WHEN `examples/discovery-k8s` has a default-deny in `data` and one allow
  from `api` THEN an inferred connection `web → postgres` is `denied` with `enforced_by:
  k8s:data/networkpolicy/default-deny`, and `api → postgres` is not.
- **FR-017** WHEN a workload reads a Secret key whose name holds `HOST`, `URL`, `URI`,
  `ENDPOINT`, `DSN`, `JDBC`, `ADDR` or `SERVER`, THE SYSTEM SHALL emit one
  `secret-endpoint` unresolved entry per Secret, naming the matching keys and every
  reading workload, and never the values (ADR-0018).
  *Scenario:* WHEN `api` and `worker` read `DB_URL` from Secret `db` THEN one entry
  `secret-endpoint` has value `db/DB_URL` and `about` both workloads.
- **FR-018** WHEN the input holds a CloudNativePG `Cluster`, a Prometheus and
  Alertmanager pair, cert-manager issuers, an external-dns Deployment, or a
  ServiceMonitor or PodMonitor, THE SYSTEM SHALL apply that operator's rule: a database
  element and its `-rw`, `-ro`, `-r` Services; Prometheus to Alertmanager; issuer to its
  ACME server or DNS provider; external-dns to its provider; the monitor's scrape
  connection from Prometheus to the selected Service; each rule with one fixture.
  *Scenario:* WHEN a `Cluster` named `shop-db` is discovered THEN the fragment holds a
  database element `shop-db` and Services `shop-db-rw`, `shop-db-ro`, `shop-db-r`
  referencing it.
- **FR-019** WHERE the system holds a discovery rules file, `<system>/discovery.yaml`
  (option of question 4), THE SYSTEM SHALL
  map a kind to an element and to references with templated fields (`{ns}`,
  `{metadata.name}`, `{spec.targetRefs[*].name}`), apply it on every discovery, and stop
  counting that kind as `unknown-kind`.
  *Scenario:* WHEN a rule maps `VaultSecret` to a reference from its target Secret to the
  Vault element THEN the 14 `VaultSecret` objects of the input yield 14 references and no
  `unknown-kind` entry for `VaultSecret`.
- **FR-020** WHEN the input holds ownerReferences, PodDisruptionBudgets,
  ServiceAccounts with RoleBindings, HorizontalPodAutoscalers or KEDA ScaledObjects, or
  cert-manager Certificates, THE SYSTEM SHALL emit the references `owner`, `selects`,
  `runs-as` and `bound-by`, `scales`, and `issues`, each with `src` and field path
  (issue #16).
  *Scenario:* WHEN HPA `api` targets Deployment `api` THEN the fragment holds a `scales`
  reference `k8s:prod/horizontalpodautoscaler/api → k8s:prod/deployment/api` with
  `field: spec.scaleTargetRef.name`.

### Wave 4 · Consumers

- **FR-021** WHEN `noodle query DIR neighbours ID [-kind K]` or `noodle query DIR path A
  B` runs, THE SYSTEM SHALL answer from the merged graph, with no model in the loop:
  every node one edge from `ID` (of kind `K`), or one shortest path from `A` to `B`, each
  edge with its kind and `src`; IF no path exists THEN it SHALL print `no path` and exit 1.
  *Scenario:* WHEN `noodle query examples/discovery-k8s neighbours k8s:prod/secret/db
  -kind secret-key` runs THEN it lists exactly `k8s:prod/deployment/api` and
  `k8s:prod/deployment/worker`, the workloads that read a key of that Secret.
- **FR-022** WHEN manifests exist for a system, THE SYSTEM SHALL make the skill run
  discovery first, give `fact-finder` only the `unresolved` entries and non-Kubernetes
  sources, document `matches` and the rules file in `reference.md`, and name the edge
  families discovery never sees (ADR-0018).
  *Scenario:* WHEN the skill builds a diagram of `examples/discovery-k8s` THEN its
  transcript reads no manifest file directly.
- **FR-023** THE SYSTEM SHALL be benchmarked on PAE with arms A, B (`fact-finder`) and C,
  two runs each, cost and accuracy against the validated `platform-overview` spec,
  results committed in `docs/benchmarks/discovery-tokens.md` and published on the
  Notion page "Benchmark · Token cost of a diagram".
  *Scenario:* WHEN the six runs are done THEN the benchmark file has one row per run with
  cost, connection recall and wall time, and arm C's recall is at least arm A's.

## Properties

- For every input, discovery run twice gives byte-identical fragments.
- For every merge, a field the curated element sets is unchanged by any fragment.
- For every fragment, no line contains the value of a Secret key of the input.

## Out of scope

- The sources card that draws provenance on a diagram (#25 second half): batch F,
  presentation (question 5). This spec keeps the data.
- Live `kubectl` discovery through client-go, Terraform, Hubble flows (roadmap phase 6).
- CI failing on drift (ADR-0015: a team decides later).
- `noodle check` rules built on these facts (#95, #96, #97): batch C, wave 2.

## Clarifications

| # | Question | Options (recommended first) | Answer | ADR |
|---|---|---|---|---|
| 1 | Who runs the checkpoint (FR-014) and the benchmark (FR-023) sessions on PAE, and what does ADR-0018's "costs clearly less" mean in numbers? | a) Claude runs each arm as a fresh `claude -p` process on a local PAE checkout, transcripts committed without tool results; "clearly less" is C at most 70 % of A's cost, in a new ADR; b) Brian runs them by hand, same threshold; c) replace PAE by `examples/discovery-k8s`, same threshold | open | |
| 2 | Kustomize (#81) means running an external tool, which ADR-0016 and ADR-0017 kept out. | a) shell out to `kustomize build` behind `-kustomize`, read-only, `unrendered` when absent, in a new ADR; b) keep no external tool: stay `unrendered`, the skill documents the loop; c) vendor the kustomize Go API | open | |
| 3 | `object` and field paths (FR-008) change the Fragment contract (C06). | a) additive optional fields in `v1alpha1`, with an ADR, no version bump; b) wait for the `v1alpha2` bump of #104; c) `object` now, field paths and git origins later | open | |
| 4 | Where do discovery rules (#80) and remembered resolutions (ADR-0017) live? `discovered/` is regenerated and uncommitted (ADR-0015). | a) one committed `<system>/discovery.yaml` holding both; b) `discovered/rules.yaml` as the issue says, committed by exception; c) two committed files, `rules.yaml` and `resolutions.yaml` | open | |
| 5 | #25 asks for a `ref` field and a sources card on the diagram. | a) this spec keeps every `src` through the merge (FR-010); the card moves to batch F; b) the card is built here too; c) drop the card, `noodle query` is enough | open | |

## Success criteria

Features `DSC-*` of phase `6a2` pass their verify step, graded in `contract.md`, and the
immutable features P6a-03 to P6a-10 they carry are flipped on their own evaluation.
