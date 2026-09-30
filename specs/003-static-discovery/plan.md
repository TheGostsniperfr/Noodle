# Plan 003 · Static discovery from Kubernetes

Spec: [`spec.md`](spec.md). Decisions: ADR-0015 (fragments, `matches`, storage),
ADR-0016 (phase 6a).

## Approach

Adapters are the only new input. They emit a `Fragment`; `internal/graph` merges it with
the curated model; everything downstream (resolve, lint, render) keeps reading the
curated model and never sees discovery.

```
manifests ─┐                                   ┌─ noodle drift  (report)
helm tpl  ─┼─▶ adapter/k8s ─▶ discovered/*.yaml ┤
kubectl   ─┘        ▲                          └─ graph.Merge(model, fragments) ─▶ skill / MCP
          adapter.Registry: k8s now, terraform later
```

## Packages

| Package | Change |
|---|---|
| `schemas/v1alpha1` | `fragment.schema.json`; `matches` on model elements |
| `internal/model` | `Fragment`, `Provenance`, `Src`; `Element.Matches`; loader reads `discovered/` when present; checks: `matches` IDs well-formed, unique across elements |
| `internal/adapter` | `Adapter` interface (`Name`, `Discover(ctx, Source) (*model.Fragment, error)`), `Registry`, `Source` (paths or reader) |
| `internal/adapter/k8s` | decode to `unstructured`, an `index` by kind, namespace and labels; one rule per file: `workloads.go`, `services.go`, `routes.go`, `ingress.go`, `secrets.go`, `externalsecrets.go`, `argocd.go`, `envrefs.go`, `netpol.go`; stable sort of the output |
| `internal/graph` | `Merge`, `Lookup`, `Neighbours`, `Drift`; directed graph with adjacency lists, cycles allowed |
| `cmd/noodle` | `discover` and `drift` subcommands |
| `tools/tokencost` | Go reader of Claude Code transcripts: weighted cost per tool category and per model, turns, wall time |
| `skills/noodle-diagram` | step 1: discover, then `fact-finder` for the rest; `matches` in `reference.md` |

No interface in `internal/graph` yet: its exported functions are the seam. A second
backend adds the interface when it exists.

## Resolution rules

- **Selectors.** Service `spec.selector` against the pod template labels of workloads in
  the same namespace. NetworkPolicy `podSelector` and `namespaceSelector` likewise, with
  namespace labels from Namespace objects when present, `kubernetes.io/metadata.name`
  otherwise.
- **Ports.** A Service port resolves its `targetPort` by number or by container port
  name. The connection carries the Service port; the workload element lists its
  container ports.
- **DNS references.** `<svc>`, `<svc>.<ns>`, `<svc>.<ns>.svc[.cluster.local]`, with an
  optional scheme and port, matched against known Services only. Anything else stays out
  of the fragment: no guessing external hosts.
- **External targets.** A SecretStore provider (1Password, Vault, AWS) and an issuer
  (ACME server) become `external` or `cloud` elements with the provider as tech.
- **Determinism.** IDs from source, maps iterated in sorted order, `observedAt` passed by
  flag in tests.

## Fixture

`examples/discovery-k8s/`: `manifests/` (a gateway, two apps with a route each, a
database reached through a secret, an ExternalSecret, a default-deny NetworkPolicy with
one allow), `model.yaml` with `matches` on a curated subset, one planned element,
`discovered/k8s-manifests.yaml` as the golden file, and the expected drift report.

## Risks

- **CRD versions.** Gateway API and External Secrets exist in several versions; decode
  through `unstructured` and read the few fields needed, not typed clients.
- **Helm output noise.** Hooks, tests and CRDs in `helm template` output: skip hooks with
  `helm.sh/hook: test`, keep migration jobs as workloads.
- **Inferred connections.** Env-based detection is a heuristic; it is marked, sourced and
  limited to known Services, so the drift report and the reviewer can reject it.
- **Benchmark variance.** One run per arm is noise; two runs each, same model, fresh
  sessions.
