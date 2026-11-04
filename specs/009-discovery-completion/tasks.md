# Tasks 009 · Discovery, phase 6a completion

Plan: [`plan.md`](plan.md). One task per commit; each ends with `make check` green. Tick
the box in the same commit. Tasks in one wave do not depend on each other, except where
a task names another. The five questions are answered (2026-11-04) and ADR-0027,
ADR-0028 and ADR-0029 accepted. Each task's first commit applies the contract's
follow-ups for its features.

## Wave 1 · References, addresses, inputs

- [x] **T01** Contract: optional `object` on discovered items, schema and loader,
  additive (FR-008, ADR-0028). Flip **DSC-04** after `make evaluate FEATURE=DSC-04`.
- [ ] **T02** Configuration references: `envFrom`, `secretKeyRef`, `configMapKeyRef`,
  Secret, ConfigMap and PVC volumes, no Secret value (FR-001); ExternalSecret, stores,
  providers, Argo CD Applications (FR-002). Spec 003 T05. Flip **DSC-01**.
- [ ] **T03** Addresses in ConfigMaps, env and CR specs to known Services, inferred
  (FR-003); public URL hosts to the Gateway or Ingress serving them (FR-004). Spec 003
  T06. Flip **DSC-02**.
- [ ] **T04** Inputs: namespace warning and `-argocd-app` (FR-005), `-kustomize` and
  `-enable-helm`, the failure when `kustomize` is missing and the rerun hint without the
  flag, `kustomize` in the flake dev shell (FR-006, ADR-0027), relative `src` paths
  (FR-007). Flip **DSC-03**.

## Wave 1 · Merge, drift, example, cost

- [ ] **T05** `tools/tokencost` with a trimmed transcript fixture (FR-013). Spec 003 T12.
- [ ] **T06** `examples/discovery-k8s`, its golden fragment and `list.yaml`, with an
  unmatched Deployment `worker`; `make examples` reruns discovery and diffs the golden
  file; `TestExample_ManifestsEqualList` in the adapter package (FR-009). Spec 003 T08.
  Flip **DSC-05**, then **P6a-03** and **P6a-04** after their own `make evaluate`.
- [ ] **T07** `internal/graph`: merge per ADR-0015, `src` and `object` kept on merged
  elements, lookup, neighbours (FR-010). Spec 003 T09. Flip **DSC-06**.
- [ ] **T08** `noodle drift` with `drift.txt` in the example; references resolved across
  fragments of one `discovered/` (FR-011, FR-012). Spec 003 T10. Flip **DSC-07**, then
  **P6a-05**.

## Wave 2 · Checkpoint

- [ ] **T09** Arms A and C on PAE `platform-overview`, cost from `tools/tokencost`, recall
  against the 25 in-cluster edges, decision recorded in `docs/PROGRESS.md` and
  `docs/benchmarks/discovery-checkpoint.md` (FR-014). Spec 003 T14. Needs T02, T03, T05.
  Flip **DSC-08**, then **P6a-08**.

## Wave 3 · Only after `Decision: go` (FR-015)

- [ ] **T10** NetworkPolicy pairs, `denied` and `enforced_by` (FR-016). Spec 003 T07.
- [ ] **T11** `secret-endpoint` entries (FR-017). Spec 003 T15. Flip **P6a-09**.
- [ ] **T12** Operator rules with one fixture each, ServiceMonitor and PodMonitor
  included (FR-018). Spec 003 T16. Flip **P6a-10**. With T10 and T11, and FR-015
  respected, flip **DSC-09**.
- [ ] **T13** Discovery rules file `<system>/discovery.yaml` and templated fields, with
  an ADR fixing the file format, rules and remembered resolutions (FR-019).
- [ ] **T14** Reference kinds `owner`, `selects`, `runs-as`, `bound-by`, `scales`,
  `issues` (FR-020). With T13, flip **DSC-10**.

## Wave 4 · Consumers

- [ ] **T15** `noodle query neighbours | path` over the merged graph (FR-021). Flip
  **DSC-11**.
- [ ] **T16** Skill step 1 runs discovery first; `matches` and the rules file in
  `reference.md`; the skill's run on the example committed as a transcript (FR-022).
  Spec 003 T11. Flip **P6a-06**.
- [ ] **T17** Benchmark of arms A, B and C, two runs each, committed and on Notion
  (FR-023). Spec 003 T13. Flip **P6a-07**. With T16, flip **DSC-12**.

## Convergence

Tasks the evaluator found missing, appended here; never edit the waves above.
