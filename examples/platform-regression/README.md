# Platform regression fixtures

Three diagrams from a real Kubernetes platform (Talos on OpenStack, Envoy Gateway,
Cilium, CloudNativePG, Vault, Argo CD), anonymized. They were built with the v0 spec and
reviewed by a human, so they are known-good inputs: lint and render must keep passing.

They exercise what the CNP example does not:

| Spec | Exercises |
|---|---|
| `platform-overview.yaml` | 25+ edges on one canvas, zones nested three deep, edges from and to zones, several `against_flow` hairpins, `blocked` edges between groups, lettered background steps, gap badges |
| `app-runtime-view.yaml` | one-application context layout (callers, gateway, namespace, reachable, denied), `blocked` edge to a zone, `link` edges to zones |
| `repo-map.yaml` | a non-network diagram: teams, a source-control tree, consumers; `link` references between repositories |

Boxes that the source project styled as planned or stacked (ADR-0008) render plain
here, since v0 has no field for it. Once B-05 and B-06 land, add `status` and
`multiplicity` to these specs.

```bash
for f in examples/platform-regression/*.yaml; do
  go run ./cmd/noodle -icons examples/platform-regression/icons "$f"
done
```
