Le test `TestMerge/takes-list-whole` vérifie seulement deux champs avec `Equal` au lieu de l'élément fusionné entier. Le critère #1 n'est donc pas atteint et le verdict est NEEDS_WORK. Les autres critères passent. Comme `docs/CONSTITUTION.md` (C15) l'impose, le fichier de preuve ci-dessous est rédigé en anglais.

# Evidence · DSC-06

- Feature: DSC-06
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: aa5e7af
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | ADR-0015 merge rules: subtests `fills-empty-field`, `keeps-curated-field`, `maps-connection-to-curated-ids`, `takes-list-whole`, each asserting with `Equal` the merged element or connection | FAIL | `go test -count=1 ./internal/graph/ -run TestMerge -v` gives `ok`, and all four subtests are `--- PASS`. Three of them assert `Equal` on the whole merged element or connection list. `takes-list-whole` (internal/graph/merge_test.go:51-62) does not. It asserts `Equal` only on `g.Model.Elements[0].Ports` and `.Tags`, never on the merged element. The threshold is not met. |
| 2 | Sources survive the merge (`keeps-sources`) | PASS | `--- PASS: TestMerge/keeps-sources`. The test asserts `Equal` on `g.Sources["api"]`: two `graph.Source` with `src` (manifests/api.yaml:5 and :24) and `object` (Deployment, then Service). The unmatched id `k8s:prod/gone` is left out. `Lookup` returns these on `Node.Sources`. |
| 3 | Commands load the merged graph | PASS | `go test -count=1 ./cmd/noodle/ -run TestLoad_MergesFragments -v` gives `--- PASS: TestLoad_MergesFragments/fills_the_ports_the_curated_element_leaves_empty`. On `examples/discovery-k8s`, `api` (model.yaml:6, no ports) gets `{http TCP 8080}`. `loadSystem` is used by `drift.go:20` and `split.go:46`, the render path. |
| 4 | Curated fields never overridden, `MaxCount` ≥ 500, scalar and list set, at least one field set by both | PASS | `--- PASS: TestMerge_Property_NeverOverridesACuratedField (0.04s)` with `quick.Config{MaxCount: 500}`. The generator forces `Tech` and `Ports` to be non-empty. It compares every non-zero curated field by reflection, and it asserts `setByBoth > 0`. |

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
    --- PASS: TestMerge_Property_NeverOverridesACuratedField (0.04s)
    ok  	github.com/TheGostsniperfr/Noodle/internal/graph	0.049s

    $ go test -count=1 ./cmd/noodle/ -run TestLoad_MergesFragments -v
    --- PASS: TestLoad_MergesFragments (0.00s)
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.009s

    $ gofmt -l internal/graph cmd/noodle; go vet ./internal/graph/ ./cmd/noodle/
    (no output) vet-ok

    $ make features
    go run ./tools/speccheck features -base HEAD   (no violation)

## Findings

- **Criterion #1 fails (takes-list-whole).** internal/graph/merge_test.go:51-62 asserts `Equal` on `Ports` and `Tags` only. The contract asks for `Equal` on the merged element. The contract preamble says this is so extra or wrong fields fail the test.
  - The gap is real. In this fixture, `fill` also copies `Tech` (`Deployment · api:1`) from the first matched element, the one without ports. The test does not check that, and it would not catch a regression in any other field.
  - **Fix:** replace the two field asserts with one `assert.Equal(t, []model.Element{{ID: "api", Kind: "backend", Tech: …, Ports: …, Tags: []string{"edge"}, Matches: …}}, g.Model.Elements)`.
  - Add this as a Convergence task in tasks.md.
- **Minor, breaks no criterion.** At internal/graph/merge_test.go:192-194, the "set by both sides" counter always counts `Kind` (n=1). The curated `Kind` is always `backend` and the discovered one always `database`, so `assert.Positive(setByBoth)` can never fail. The no-override check is unaffected because `Tech` and `Ports` are compared, but the guard is vacuous. **Fix:** count only fields other than `ID` and `Kind`, or require that `Tech` or `Ports` was set by both.
- **Minor, breaks no criterion.** `keeps-sources` asserts the sources through `g.Sources["api"]`, not on the merged element. FR-010 says "on the merged element". `Lookup` exposes them through `Node.Sources`, so the behaviour is there. Asserting through `g.Lookup("api")` would test it the way commands read it.
- **Reversion check.**
  - `fills-empty-field` and `TestLoad_MergesFragments` would fail if `fill` were removed.
  - `keeps-curated-field` and the property test would fail if `fill` overwrote set fields.
  - `maps-connection-to-curated-ids` would fail if `mapConnections` were removed.
  - `keeps-sources` would fail if `Sources` were not recorded.
  - Each of these tests catches the removal of the behaviour it covers.
- **Scope and constitution.** The diff for this feature (53fd95b) touches `internal/graph`, `cmd/noodle/load.go` and `split.go`, which plan 009 lists. `features.json` has no edited entries (`make features` is clean). No lint rule changed.
