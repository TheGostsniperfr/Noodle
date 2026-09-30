# Platform regression fixtures

Three diagrams from a real Kubernetes platform (Talos on OpenStack, Envoy Gateway,
Cilium, CloudNativePG, Vault, Argo CD), anonymized. They were built with the v0 spec and
reviewed by a human, so they are known-good inputs: lint and render must keep passing.

They exercise what the CNP example does not:

| System | Exercises |
|---|---|
| `platform-overview/` | 25+ edges on one canvas, zones nested three deep, edges from and to zones, several `against_flow` hairpins, `blocked` edges between groups, lettered background steps, gap badges |
| `app-runtime-view/` | one-application context layout (callers, gateway, namespace, reachable, denied), `blocked` edge to a zone, `link` edges to zones |
| `repo-map/` | a non-network diagram: teams, a source-control tree, consumers; `link` references between repositories |

Boxes that the source project styled as planned or stacked (ADR-0008) render plain
here, since v0 had no field for it. `status` can be added to these models now that
B-05 has landed; `multiplicity` waits for B-06.

```bash
for d in platform-overview app-runtime-view repo-map; do
  go run ./cmd/noodle render examples/platform-regression/$d -icons examples/platform-regression/icons
done
```

Each directory is a system in the v1alpha1 format (ADR-0007), produced by
`noodle migrate` from the original v0 files, which now live in
`internal/migrate/testdata` as the migration's regression inputs.
