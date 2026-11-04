# Evidence · DSC-05

- Feature: DSC-05
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 05ed45d
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Golden from manifests, twice, and from a List | PASS | `go test ./cmd/noodle/ -run TestDiscoverExample -v -count=1` exited 0. `--- PASS: TestDiscoverExample/manifests-first`, `.../manifests-second` and `.../kubectl-list` passed. Each one runs `assert.Equal(t, string(golden), stdout)` against `examples/discovery-k8s/discovered/k8s-manifests.yaml`. |
| 2 | Determinism for every input | PASS | `go test ./internal/adapter/k8s/ -run TestDiscover_Property_Deterministic -v -count=1` printed `--- PASS (1.70s)`. The test uses `quick.Check` with `&quick.Config{MaxCount: 200}`. It builds 2 to 13 documents from 7 generators, applies `r.Shuffle` to their order, and requires two runs to be byte-identical with no error. |
| 3 | The example has what spec 003's plan lists | PASS | Kind counts: Gateway 1, HTTPRoute 2, Deployment 3, StatefulSet 1, Secret 1, ExternalSecret 1, NetworkPolicy 2. `matches:` = 5. `status: planned` = 1. |
| 4 | `make examples` reruns and diffs | PASS | `make examples` exited 0 and ran `go run ./cmd/noodle discover k8s examples/discovery-k8s/manifests -observed-at 2026-11-04T00:00:00Z \| diff - examples/discovery-k8s/discovered/k8s-manifests.yaml`. Both Makefile greps returned `1`. |

## Verify step

    $ go test ./cmd/noodle/ -run 'TestDiscoverExample' -v -count=1
    --- PASS: TestDiscoverExample (0.04s)
        --- PASS: TestDiscoverExample/manifests-first (0.01s)
        --- PASS: TestDiscoverExample/manifests-second (0.01s)
        --- PASS: TestDiscoverExample/kubectl-list (0.01s)
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.049s

    $ make examples; echo exit=$?
    go run ./cmd/noodle discover k8s examples/discovery-k8s/manifests -observed-at 2026-11-04T00:00:00Z | diff - examples/discovery-k8s/discovered/k8s-manifests.yaml
    k8s: read 15 objects, ~1.0k tokens; skipped 0; fragment 13 elements, 8 connections, 9 references, 0 unresolved, ~2.1k tokens (unresolved ~0)
    exit=0

    $ go test ./internal/adapter/k8s/ -run 'TestDiscover_Property_Deterministic|TestExample_ManifestsEqualList' -v -count=1
    --- PASS: TestExample_ManifestsEqualList (0.02s)
    --- PASS: TestDiscover_Property_Deterministic (1.70s)
    ok  	github.com/TheGostsniperfr/Noodle/internal/adapter/k8s	1.740s

Other checks:
- `gofmt -l` on both packages printed nothing, and `go vet` passed.
- `go run ./tools/speccheck locate DSC-05` printed `specs/009-discovery-completion`.

Tests that would catch the feature being reverted:
- `TestDiscoverExample` compares full stdout with the golden file. It fails if the example, the golden file or the List path changes.
- `kubectl-list` stages `list.yaml` at the manifests path. If List unwrapping were dropped, the output would have no items and would differ from the golden file.
- `TestExample_ManifestsEqualList` checks that `list.yaml` equals the generator output, and that the fragment from the List equals the fragment from the manifests.

Scope: commit df1dec0 only touches the example, the Makefile, two test files, COMMANDS.md, PROGRESS.md and tasks.md.

Constitution:
- C04 holds: no lint rule was relaxed.
- C12 holds: `docs/features.json` only gains new entries on this branch, and DSC-05 is still `passes: false`.

The feature draws nothing, so there is no PNG to inspect.

## Findings

- `examples/discovery-k8s/list.yaml` is not real `kubectl get -o yaml` output. `listFromManifests` in `internal/adapter/k8s/example_test.go:32` generates it so that each item starts on the same line as its manifest document, and turns `---` into `  # ---` comment lines. Real kubectl output has `status`, `managedFields`, server-set metadata and no comments. The List path is exercised, but only in its most favourable shape. This breaks no criterion. A second fixture with typical server-side fields would make FR-009's "as `kubectl get -o yaml` prints them" claim honest.
- `TestDiscover_Property_Deterministic` (`internal/adapter/k8s/example_test.go:121`) shuffles once and then runs the same input twice. It catches map-iteration nondeterminism. It does not show that the output is independent of input order: `src` lines would differ, so that claim cannot hold anyway. This is consistent with the threshold as written. A note in the test saying order-independence is out of scope would avoid misreading.
- Neither NetworkPolicy shows up in the golden fragment, which is expected until DSC-09 (wave 3). When DSC-09 lands, the golden file and `drift.txt` will need regenerating, and DSC-05 #1 will flag that.
- The worktree has an untracked `specs/009-discovery-completion/evidence/P6a-04.md`. It is outside DSC-05 and is reported only so it is not committed by accident.
