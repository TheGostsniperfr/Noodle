# Evidence · DSC-13

- Feature: DSC-13
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 2aecda6
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The digest of the example is golden | PASS | `go test ./cmd/noodle/ -run TestDigest -v` gives `ok`. The five subtests `folds-service`, `resolves-route`, `keeps-externalsecret-path`, `drops-config-refs` and `keeps-unresolved` pass, and each runs `assert.Equal(t, tt.want, stdout.String())` on the full output. `go run ./cmd/noodle digest examples/discovery-k8s/discovered \| diff - examples/discovery-k8s/digest.txt` gives an empty diff with exit 0. |
| 2 | Config references are left out | PASS | `grep -cE 'configmap/' …digest.txt` returns `0`. `grep -E 'secret/' … \| grep -vcE 'externalsecret\|secret-endpoint'` returns `0`. |
| 3 | Every line has a source | PASS | `grep -vcE '^(#.*)?$\|src [^ ]+:[0-9]+' …digest.txt` returns `0`. |
| 4 | Size bound on PAE (FR-025) | PASS | The grep on `discovery-checkpoint-2.md` returns one line: `Tokens: fragments ~35.0k, digest ~6.4k (18 %, bound 25 %, FR-025)`. 6.4 / 35.0 = 18.3 %, which is under 25 %. Both numbers come from `kTokens` (bytes/4) in `cmd/noodle/discover.go:167`, which is the estimator `noodle discover` uses, as FR-025 requires. |

## Verify step

    $ go test ./cmd/noodle/ -run TestDigest -v
    --- PASS: TestDigest (0.00s)
        --- PASS: TestDigest/keeps-externalsecret-path (0.00s)
        --- PASS: TestDigest/drops-config-refs (0.00s)
        --- PASS: TestDigest/folds-service (0.00s)
        --- PASS: TestDigest/keeps-unresolved (0.00s)
        --- PASS: TestDigest/resolves-route (0.00s)
    --- PASS: TestDigest_FailsOnAnEmptyDirectory (0.00s)
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle

    $ go run ./cmd/noodle digest examples/discovery-k8s/discovered | diff - examples/discovery-k8s/digest.txt
    digest: 1 fragments ~2.1k tokens, digest ~384 tokens
    (no diff, exit 0)

    $ grep -E '^Tokens: fragments ~[0-9.]+k, digest ~[0-9.]+k' docs/benchmarks/discovery-checkpoint-2.md
    Tokens: fragments ~35.0k, digest ~6.4k (18 %, bound 25 %, FR-025)

    $ make examples
    go run ./cmd/noodle digest examples/discovery-k8s/discovered 2>/dev/null | diff - examples/discovery-k8s/digest.txt
    exit=0

    $ go vet ./cmd/noodle/ && gofmt -l cmd/noodle/digest*.go internal/model/load.go
    vet-ok   (no gofmt output for the files in the diff)

Tests checked against a revert of the feature: each subtest compares the whole stdout with `Equal`, so dropping a behaviour fails its own subtest.
- Remove the Service fold and `folds-service` gets `prod/web -> prod/service/api`.
- Remove route resolution and `resolves-route` loses the `edge/gateway/shared -> prod/api  host …` line.
- Remove the ExternalSecret chain and `keeps-externalsecret-path` fails.
- Stop filtering ConfigMap, Secret and PVC references and `drops-config-refs` fails, because the expected output has no line for them.
- Remove the open section and `keeps-unresolved` fails.

Without `digest.go` the package does not compile, so every subtest fails.

## Findings

- **Size bound not reproducible from the repository.** `docs/benchmarks/discovery-checkpoint-2.md:8` is the only record of the PAE size bound. The PAE fragments and the digest's stderr are not committed, so I graded the reported line, not a rerun. The threshold of criterion 4 only asks for the reported line, so this breaks no criterion. Fix: commit the `noodle digest` stderr line of the PAE run, or a hash of the fragment set, next to the report.
- **Possible panic.** At `cmd/noodle/digest.go:231`, a `secret-endpoint` entry reads `u.About[0]` without checking its length. An entry with no `about` would panic instead of returning an error. The adapter always sets `about` today. Fix: guard on `len(u.About) > 0`, or return a wrapped error.
- **Plan file table not updated.** At `specs/009-discovery-completion/plan.md:37`, the `cmd/noodle` row does not list the `digest` subcommand. At line 39, the `examples/discovery-k8s` row does not list `digest.txt`. Line 32, the `internal/model` row, does not list the new `LoadFragments`. ADR-0031 covers the scope, so this is documentation drift and not scope creep. Fix: add them to the table, as T26 did for P6a-09.
- **Pre-existing gofmt failure.** `gofmt -l cmd/noodle` lists `cmd/noodle/setup_test.go`. That file is not in this diff, so it is not this feature's fault. It should still be fixed, because C15 requires gofmt'd Go.
