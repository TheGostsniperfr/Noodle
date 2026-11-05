# Evidence · DSC-07

- Feature: DSC-07
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 05ed45d
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The four drift families, sorted | PASS | `go test ./internal/graph/ -run TestDrift -v`: exit 0. It runs the subtests `unmatched`, `gone`, `planned-absent`, `unmodelled-connection` and `sorted`. Each one uses `assert.Equal` on the full `[]graph.Entry`, including `Src`. `sorted` gives every family its entries in reverse order: `zz`/`aa`, `b-gone`(z,a)/`a-gone`, `z-plan`/`a-plan`, and flow `y->w` before `x->w`. It then checks the sorted output with `Equal`. |
| 2 | Cross-fragment resolution | PASS | `go test ./internal/graph/ -run TestCrossFragment -v`: `--- PASS: TestCrossFragment`. The route in `k8s-apps` and the Gateway in `k8s-gateway-infra` are linked by a `parentRef` edge, checked with `Equal` through `Neighbours`. `g.Unresolved` holds only the ghost Service, so no `missing-backend` remains for the route. |
| 3 | The example's report, byte for byte | PASS | `go test ./cmd/noodle/ -run TestDrift -v`: `--- PASS: TestDrift`. I also ran it myself: `go run ./cmd/noodle drift examples/discovery-k8s` gives `exit=0`, and `diff` against `drift.txt` shows no difference. `make examples` exits 0 and includes `drift ... \| diff - examples/discovery-k8s/drift.txt`. |
| 4 | The report is not empty and holds FR-011's scenario | PASS | `grep -cE '^(unmatched\|gone\|planned\|unmodelled)'` gives `4`. The headings are at lines 1, 6, 8 and 10, and each is followed by at least one indented entry (4, 1, 1 and 3 entries). `grep -cE 'worker.*manifests/.*:[0-9]+'` gives `1`, on line 3 (`k8s:prod/deployment/worker  examples/discovery-k8s/manifests/platform.yaml:82`), which is under `unmatched (4)`. |

## Verify step

    $ go test ./internal/graph/ ./cmd/noodle/ -run 'TestDrift|TestCrossFragment' -count=1
    ok  	github.com/TheGostsniperfr/Noodle/internal/graph	0.004s
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.012s

    $ go run ./cmd/noodle drift examples/discovery-k8s > /tmp/drift.out; echo "exit=$?"; diff /tmp/drift.out examples/discovery-k8s/drift.txt && echo IDENTICAL
    exit=0
    IDENTICAL

    $ make examples
    go run ./cmd/noodle drift examples/discovery-k8s | diff - examples/discovery-k8s/drift.txt
    exit=0

    $ go vet ./internal/graph/ ./cmd/noodle/ && gofmt -l internal/graph cmd/noodle
    VET_OK

## Findings

These are minor and none breaks a contract criterion.

- **Would the tests catch a revert?** Yes:
  - Removing `Drift()` breaks the build.
  - Removing the `sort.Slice` loop at `internal/graph/drift.go:59` fails `sorted`, because `Gone` and `Planned` are filled in model order and arrive in reverse. `Unmatched` and `Unmodelled` come from map iteration and would become flaky.
  - Removing `settleAcrossFragments` leaves `g.Unresolved` empty, so `TestCrossFragment`'s `Equal` on `[ghost]` fails.
  - `cmd/noodle` `TestDrift` is a golden byte comparison.
- `internal/graph/drift_test.go:113` (`TestCrossFragment`) builds the `Unresolved` entries by hand. It never runs the k8s adapter on two manifest sets. I read `internal/adapter/k8s/routes.go:60,174`: the adapter emits `id(ns, "Gateway", name)` with hint `gateway`/`service`, which matches the `kinds` map at `drift.go:89`, so the real path should work. The only example ships a single fragment (`discovered/k8s-manifests.yaml`), though, so no end-to-end test covers FR-012's scenario (`k8s-apps.yaml` + `k8s-gateway-infra.yaml`). Fix: add a two-fragment fixture through `noodle discover`, or split the example into two fragments.
- `internal/graph/drift.go:11`: `gone` and `planned` entries cite `model.yaml` with no line number. FR-011 only asks for "their `src`", so this is within contract. The PROCESS.md example (`model.yaml:12`) suggests readers will expect a line, and `model.Element` has no line to give one.
- `internal/graph/drift.go:92`: a `missing-backend` with hint `not-a-service` is never resolved across fragments, even when the target is defined. That is defensible, but the spec does not mention it.
- Scope: commit `0ff3438` only touches the drift command, `internal/graph`, the golden file, one Makefile line, PROGRESS and tasks. The large `origin/main...HEAD` diff comes from earlier spec 007/009 work that is not part of this feature. Constitution:
  - C12: `features.json` DSC-07 is unchanged since `1e71a86`.
  - C04: no lint rule touched.
  - C15: gofmt and vet are clean.
  - C13: not visual, so no PNG to review.
