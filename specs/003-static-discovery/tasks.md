# Tasks 003 · Static discovery from Kubernetes

Plan: [`plan.md`](plan.md). One task per commit. Each ends with `go vet ./... && go test ./...`
and `./scripts/init.sh` green. Tick the box in the same commit. Run `go mod vendor` after
touching `go.mod`.

## A · Contract (P6a-01)

- [x] **T01** `Fragment`, `Provenance`, `Src` and `Element.Matches` in `internal/model`;
  `fragment.schema.json` and `matches` in the model schema; loader reads `discovered/`.
  Checks: `matches` well-formed and unique. Table-driven tests. Flip **P6a-01**.

## B · Adapter frame (P6a-02)

- [x] **T02** `internal/adapter`: interface, registry, `Source` from paths or stdin.
  `noodle discover <adapter> [paths…|-] [-o]`. A test adapter registered in a test runs
  end to end. Flip **P6a-02**.

## C · Kubernetes adapter (P6a-03, P6a-04)

- [x] **T03** Decode files, directories, stdin, multi-document and `kind: List`; skip
  status, managed fields and generated kinds; build the index. Workloads, Services,
  namespaces as zones, ports.
- [x] **T04** Routes: Gateway, HTTPRoute, GRPCRoute, Ingress; Service to workload by
  selector and target port.
- [ ] **T05** Secrets and references: envFrom, secretKeyRef, volumes, ExternalSecret,
  SecretStore providers, Argo CD Applications. No secret value in the output.
- [ ] **T06** Env and ConfigMap DNS references to known Services, marked inferred.
- [ ] **T07** NetworkPolicy: allowed pairs, `denied` and `enforced_by` on connections a
  policy blocks.
- [ ] **T08** `examples/discovery-k8s` with its golden fragment; `init.sh` and CI rerun
  discovery and diff against it. Manifests and the same objects wrapped as a
  `kubectl get -o yaml` List give the same fragment. Flip **P6a-03**, **P6a-04**.

## C2 · Checkpoint and blind spots (P6a-08, P6a-09, P6a-10), ADR-0018

Order: T05, T06, T12, **T14**; then T15, T16 and T07 only if T14 passes.

- [ ] **T14** Checkpoint on PAE: `platform-overview` built twice in fresh sessions, agent
  alone (arm A) and agent given the fragment by hand (arm C), one run each; cost from
  T12, recall against the 25 in-cluster edges. Go only if C costs clearly less and finds
  as many edges; record the result and the decision in PROGRESS. Flip **P6a-08**.
- [ ] **T15** `secret-endpoint`: workloads reading address-like Secret keys (env,
  envFrom, ExternalSecret targets), one entry per Secret, values never read. Flip
  **P6a-09**.
- [ ] **T16** Operator rules with one fixture each: CloudNativePG `Cluster`, Prometheus
  to Alertmanager, cert-manager issuers, external-dns provider. Flip **P6a-10**.

## D · Merge and drift (P6a-05)

- [ ] **T09** `internal/graph`: merge per ADR-0015, lookup, neighbours. `internal/access`
  (spec 004) has its own traversal until then; reuse or replace it here (ADR-0020).
- [ ] **T10** `Drift` and `noodle drift <system>`: unmatched, gone, planned-and-absent,
  unmodelled connections. Expected report in the example. Flip **P6a-05**.

## E · Skill and benchmark (P6a-06, P6a-07)

- [ ] **T11** Skill step 1 runs discovery first; `fact-finder` gets only what is left;
  `matches` documented in `reference.md`. The skill treats the fragment as a floor and
  names the edge families discovery never sees (ADR-0018). Flip **P6a-06**.
- [ ] **T12** `tools/tokencost`: transcript reader in Go (port of the scratch script
  used for the baseline). Test on a trimmed transcript fixture. Needed by T14, so done
  right after T06.
- [ ] **T13** Benchmark on PAE: arms A, B, C, two runs each; cost and accuracy against
  the validated platform-overview spec; results in Notion and in the PR. Flip **P6a-07**.
