# Evidence · DSC-06

- Feature: DSC-06
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 53ab054
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | ADR-0015 merge rules: subtests `fills-empty-field`, `keeps-curated-field`, `maps-connection-to-curated-ids` and `takes-list-whole`, each asserting with `Equal` the merged element or connection | PASS | `go test -count=1 ./internal/graph/ -run TestMerge -v` gives `ok`, and all four subtests show `--- PASS`. Each one asserts `Equal` on `g.Model.Elements` or `g.Model.Connections` as a whole: merge_test.go:37, :49, :112 and :62. `takes-list-whole` now compares the whole element, with `Tech`, `Ports`, `Tags` and `Matches` (T18, 9705637). This closes the previous finding. |
| 2 | Sources survive the merge (`keeps-sources`) | PASS | `--- PASS: TestMerge/keeps-sources`. merge_test.go:77-82 reads `g.Lookup("api")` and asserts `Equal` on `node.Sources`. That gives two `graph.Source`: `src` manifests/api.yaml:5 with object Deployment, and :24 with object Service. The unmatched `k8s:prod/gone` is left out. |
| 3 | Commands load the merged graph | PASS | `go test -count=1 ./cmd/noodle/ -run TestLoad_MergesFragments -v` gives `--- PASS: .../fills_the_ports_the_curated_element_leaves_empty`. `loadSystem("../../examples/discovery-k8s")` returns `api` with `{http TCP 8080}`. The curated `api` has no ports (model.yaml line 6), and the curated `Tech: "[Deployment]"` is kept. |
| 4 | Curated fields are never overridden, for any input | PASS | `--- PASS: TestMerge_Property_NeverOverridesACuratedField (0.05s)`. The test uses `quick.Config{MaxCount: 500}`. The generator forces a non-empty `Tech` (scalar) and `Ports` (list) on every curated element (merge_test.go:164-169). Reflection compares every non-zero curated field after the merge. `setByBoth` counts only fields other than ID, Kind and Matches that are set on both sides with different values (:200-203), so `assert.Positive` can now fail (T19). |

## Verify step

    $ go test -count=1 ./internal/graph/ -run TestMerge -v
    --- PASS: TestMerge (0.00s)
        --- PASS: TestMerge/fills-empty-field (0.00s)
        --- PASS: TestMerge/keeps-curated-field (0.00s)
        --- PASS: TestMerge/takes-list-whole (0.00s)
        --- PASS: TestMerge/keeps-sources (0.00s)
        --- PASS: TestMerge/maps-connection-to-curated-ids (0.00s)
        --- PASS: TestMerge/keeps-the-port-the-curated-end-has (0.00s)
        --- PASS: TestMerge/does-not-depend-on-the-order-of-fragments (0.00s)
    --- PASS: TestMerge_Property_NeverOverridesACuratedField (0.05s)
    ok  	github.com/TheGostsniperfr/Noodle/internal/graph	0.060s

    $ go test -count=1 ./cmd/noodle/ -run TestLoad_MergesFragments -v
    --- PASS: TestLoad_MergesFragments (0.01s)
        --- PASS: TestLoad_MergesFragments/fills_the_ports_the_curated_element_leaves_empty (0.00s)
        --- PASS: TestLoad_MergesFragments/keeps_the_ports_the_curated_element_sets (0.00s)
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.017s

    $ gofmt -l internal/graph cmd/noodle; go vet ./internal/graph/ ./cmd/noodle/
    (no gofmt output) vet-ok

    $ make features
    go run ./tools/speccheck features -base HEAD   (no violation)

## Findings

- **Reversion check.** I read the tests against graph.go:64-182. Each one would catch the loss of the behaviour it covers:
  - `fills-empty-field` and `TestLoad_MergesFragments` would fail without `fill` (graph.go:99).
  - `keeps-curated-field` and the property test would fail if `fillString` or the list guards overwrote a set field.
  - `takes-list-whole` would fail if ports were merged item by item, because it would get 443 appended.
  - `maps-connection-to-curated-ids` would fail without `mapConnections`.
  - `keeps-sources` would fail if `g.Sources` were not recorded, or if `Lookup` did not expose it.
- **Minor, breaks no criterion.** cmd/noodle/load_test.go:37 checks the loader's sources only with `NotEmpty`. FR-010's scenario says the sources "include the Service's file and line". The unit subtest `keeps-sources` proves that, but the end-to-end path through the example does not check it. One `assert.Contains` on the Service `src` would close the gap.
- **Scope and constitution.** The three convergence commits (9705637, 05acb9c and 53ab054) touch only `internal/graph/merge_test.go`, `tasks.md` and `PROGRESS.md`. The feature code (53fd95b) stays inside `internal/graph` and `cmd/noodle`, as plan 009 lists. In `features.json`, the only changes since the last evaluation are `passes` flips of other features, and DSC-06 is still `false`. No lint rule was relaxed (C04, C12).
