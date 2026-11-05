# Plan 009 · Discovery, phase 6a completion

Spec: [`spec.md`](spec.md). Decisions: ADR-0015 (fragments, `matches`, storage), ADR-0016,
ADR-0017 (decoding, `unresolved`, remembered resolutions), ADR-0018 (checkpoint, operator
rules), ADR-0025 (batches), ADR-0027 (kustomize), ADR-0028 (object identity), ADR-0029
(checkpoint threshold). The rules file format of question 4 gets its ADR in T13.

Phase `6a` in `docs/features.json` now points at this spec, so `make evaluate
FEATURE=P6a-0N` reads this contract and its evidence lands in
`specs/009-discovery-completion/evidence/`, in full mode.

## Approach

Spec 003's shape stands: adapters write a `Fragment`, `internal/graph` merges it with the
curated model, everything downstream reads the merged graph. This spec fills the rules
spec 003 left open, adds what the Arffornia field report found missing, and puts the
ADR-0018 checkpoint between the cheap rules (wave 1) and the costly ones (wave 3).

```
manifests ─┐  -kustomize DIR ─▶ kustomize build (read-only, ADR-0027)
helm tpl  ─┼─▶ adapter/k8s ─▶ discovered/k8s-*.yaml ─┐
kubectl   ─┘     rules: refs, addresses, operators,   ├─ graph.Merge(model, fragments)
                 <system>/discovery.yaml               │     ├─ noodle drift
                                                       │     ├─ noodle query neighbours | path
                                                       │     └─ skill, MCP later
```

## Packages and files

| Path | Change |
|---|---|
| `internal/model` | discovered items gain an optional `object` (ADR-0028); `LoadFragments` for `digest` (ADR-0031) |
| `schemas/v1alpha1/` | `fragment.schema.json`: optional `object` (ADR-0028), `secret-endpoint` in the `unresolved` kinds (ADR-0018) |
| `flake.nix` | `kustomize` in the dev shell (ADR-0027) |
| `internal/adapter/k8s` | one rule per file: `refs.go` (FR-001), `refkinds.go` (FR-020), `externalsecrets.go` and `argocd.go` (FR-002), `addresses.go` (FR-003, FR-004), `namespace.go` (FR-005), `kustomize.go` (FR-006), `netpol.go` (FR-016), `secretendpoint.go` (FR-017), `operators.go` (FR-018), `rules.go` (FR-019); `src` paths relative (FR-007) |
| `internal/graph` | new: `Merge`, `Lookup`, `Neighbours`, `Path`, `Drift`; cross-fragment index (FR-012) |
| `cmd/noodle` | `discover` flags `-argocd-app`, `-kustomize`, `-enable-helm`; new `drift`, `query` and `digest` (FR-024, ADR-0031) subcommands |
| `tools/tokencost` | new: transcript reader (FR-013) |
| `examples/discovery-k8s` | manifests, `list.yaml`, `model.yaml`, golden `discovered/k8s-manifests.yaml`, `drift.txt`, `digest.txt` |
| `docs/benchmarks/` | `discovery-checkpoint.md`, `discovery-checkpoint-2.md`, `render-pae.sh`, `run-arm.sh`, the prompts and the ground truth, `discovery-tokens.md`, and `transcripts/*.jsonl` of every arm and of the skill run |
| `skills/noodle-diagram` | step 1 runs discovery first; `reference.md` documents `matches` and the rules file |
| `Makefile`, `docs/COMMANDS.md` | `make examples` reruns discovery on the example and diffs the golden file |

`internal/access` (spec 004) keeps its own traversal (ADR-0020); `internal/graph` does
not replace it in this spec.

## Resolution rules added to spec 003's

- **Addresses.** Values are scanned, never evaluated: ConfigMap `data`, env `value`, and
  every string leaf of a custom resource `spec`. A match must name a Service of the input
  (FR-003) or a route hostname of the input (FR-004); anything else stays out, as in
  spec 003.
- **Ports of an address.** A value's port picks the Service port with that number; a
  value with no port uses the Service's only port, or, with several, the port named
  `http`, else the first in spec order.
- **Namespace.** Fallback order: the object's namespace, `-namespace`, `-argocd-app`'s
  destination, then `default` with the FR-005 warning.
- **Kustomize.** `-kustomize DIR` looks `kustomize` up on PATH first and fails before
  any work when it is absent, with `kustomize not found: nix shell nixpkgs#kustomize, or
  nix develop in this repository`. Without the flag, a walked directory holding
  `kustomization.yaml` gets the `unrendered` hint `noodle discover k8s -kustomize <DIR>`.
- **Paths.** `noodle query path` is a breadth-first search; among equal-length paths it
  takes, at each step, the next id first in lexicographic order, so the answer is stable.
- **Transcripts.** Committed transcripts keep usage records and `tool_use` inputs and
  drop every `tool_result`, which may hold a manifest's Secret values (C09).
- **Merge.** Curated fields win; a list field the curated element leaves empty is taken
  whole from the fragment, never merged item by item.

## Alternatives set aside

A graph database behind `internal/graph`: ADR-0015 revisits storage only when loading
becomes slow. A typed Kubernetes client: ADR-0017 keeps yaml.v3 decoding.

## Risks

| Risk | Mitigation |
|---|---|
| Address scanning finds false connections | limited to Services and route hosts of the input, marked `inferred`, sourced to the line; drift and the reviewer can reject them |
| The checkpoint says stop after wave 1 | FR-015: wave 3 is not started; the spec is revised, wave 1 and the merge still ship |
| `kustomize` output differs from Argo CD's | the flag passes Argo CD's build options; the hint names the version found |
| The Fragment contract change breaks existing fragments | one additive optional field (ADR-0028); existing golden fragments regenerate identically |
