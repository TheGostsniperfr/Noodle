# Tasks 003 · Static discovery from Kubernetes

Plan: [`plan.md`](plan.md). One task per commit. Each ends with `go vet ./... && go test ./...`
and `./scripts/init.sh` green. Tick the box in the same commit. Run `go mod vendor` after
touching `go.mod`.

## A · Contract (P6a-01)

- [ ] **T01** `Fragment`, `Provenance`, `Src` and `Element.Matches` in `internal/model`;
  `fragment.schema.json` and `matches` in the model schema; loader reads `discovered/`.
  Checks: `matches` well-formed and unique. Table-driven tests. Flip **P6a-01**.

## B · Adapter frame (P6a-02)

- [ ] **T02** `internal/adapter`: interface, registry, `Source` from paths or stdin.
  `noodle discover <adapter> [paths…|-] [-o]`. A test adapter registered in a test runs
  end to end. Flip **P6a-02**.

## C · Kubernetes adapter (P6a-03, P6a-04)

- [ ] **T03** Decode files, directories, stdin, multi-document and `kind: List`; skip
  status, managed fields and generated kinds; build the index. Workloads, Services,
  namespaces as zones, ports.
- [ ] **T04** Routes: Gateway, HTTPRoute, GRPCRoute, Ingress; Service to workload by
  selector and target port.
- [ ] **T05** Secrets and references: envFrom, secretKeyRef, volumes, ExternalSecret,
  SecretStore providers, Argo CD Applications. No secret value in the output.
- [ ] **T06** Env and ConfigMap DNS references to known Services, marked inferred.
- [ ] **T07** NetworkPolicy: allowed pairs, `denied` and `enforced_by` on connections a
  policy blocks.
- [ ] **T08** `examples/discovery-k8s` with its golden fragment; `init.sh` and CI rerun
  discovery and diff against it. Manifests and the same objects wrapped as a
  `kubectl get -o yaml` List give the same fragment. Flip **P6a-03**, **P6a-04**.

## D · Merge and drift (P6a-05)

- [ ] **T09** `internal/graph`: merge per ADR-0015, lookup, neighbours.
- [ ] **T10** `Drift` and `noodle drift <system>`: unmatched, gone, planned-and-absent,
  unmodelled connections. Expected report in the example. Flip **P6a-05**.

## E · Skill and benchmark (P6a-06, P6a-07)

- [ ] **T11** Skill step 1 runs discovery first; `fact-finder` gets only what is left;
  `matches` documented in `reference.md`. Flip **P6a-06**.
- [ ] **T12** `tools/tokencost`: transcript reader in Go (port of the scratch script
  used for the baseline). Test on a trimmed transcript fixture.
- [ ] **T13** Benchmark on PAE: arms A, B, C, two runs each; cost and accuracy against
  the validated platform-overview spec; results in Notion and in the PR. Flip **P6a-07**.
