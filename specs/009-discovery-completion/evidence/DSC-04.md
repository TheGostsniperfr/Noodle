# Evidence · DSC-04

- Feature: DSC-04
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 05ed45d
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The contract: optional top-level `object`, `src.object` kept | PASS | `go test ./internal/model/ -run TestIdentity -v`: `--- PASS: TestIdentity`, with subtests `schema-accepts` (5 cases: with object, cluster-scoped without namespace, without object, connection, reference), `schema-rejects/{no_apiVersion,no_kind,no_name}`, `src-object-unchanged` and `legacy-fragment`. `legacy-fragment` validates the fixture, runs load→encode→load with `assert.Equal(first, second)`, asserts no top-level `object` key and checks `bytes.Equal(encoded, reencoded)`. `ok internal/model` |
| 2 | The legacy fixture predates this spec | PASS | `git log --format=%h origin/main -- internal/model/testdata/minimal/discovered/k8s-manifests.yaml` → `f611cde` |
| 3 | The adapter fills them where an object stands behind the item | PASS | `go test ./internal/adapter/k8s/ -run TestIdentity -v`: `--- PASS: TestIdentity/{scenario,declared-namespace-and-cluster-scoped,edges-carry-the-object-that-created-them,connection-object,every-item}`. `scenario` checks FR-008's api/namespace pair with `Equal`. `every-item` runs on `examples/discovery-k8s/manifests` and checks with `Equal` that the elements without `object` are exactly `[k8s:_external/onepassword, k8s:data/namespace/data, k8s:prod/namespace/prod]`. It also asserts that every connection and every reference has a non-nil `Object`. `connection-object` checks with `Equal` that the inferred connection carries ConfigMap `prod/settings`. No `field` key: no struct has a `yaml:"field"` tag, the schema items are `additionalProperties: false`, and grep on the golden fragment gives `0` |
| 4 | Existing fixtures are not rewritten | PASS | `git diff --name-only --diff-filter=MD origin/main...HEAD -- internal/adapter/k8s/testdata cmd/noodle/testdata internal/model/testdata` → no output, rc=0 |
| 5 | The decision is recorded (C06) | PASS | `grep -c '^status: accepted' docs/adr/0028-fragments-carry-object-identity.md` → `1`; `make spec-check` → exit=0 |
| 6 | No version bump | PASS | `ls schemas/` → `v1alpha1`; `grep -c 'noodle/v1alpha1' schemas/v1alpha1/fragment.schema.json` → `1`. The only apiVersion value in that file is `"const": "noodle/v1alpha1"` (line 10) |

## Verify step

    $ go test ./internal/model/ ./internal/adapter/k8s/ -run 'TestIdentity' -v -count=1
    --- PASS: TestIdentity (0.01s)
        --- PASS: TestIdentity/src-object-unchanged (0.00s)
        --- PASS: TestIdentity/schema-rejects (0.00s)
        --- PASS: TestIdentity/schema-accepts (0.00s)
        --- PASS: TestIdentity/legacy-fragment (0.01s)
    ok  	github.com/TheGostsniperfr/Noodle/internal/model	0.032s
    --- PASS: TestIdentity (0.00s)
        --- PASS: TestIdentity/scenario (0.00s)
        --- PASS: TestIdentity/declared-namespace-and-cluster-scoped (0.00s)
        --- PASS: TestIdentity/edges-carry-the-object-that-created-them (0.00s)
        --- PASS: TestIdentity/connection-object (0.00s)
        --- PASS: TestIdentity/every-item (0.00s)
    ok  	github.com/TheGostsniperfr/Noodle/internal/adapter/k8s	0.013s

    "existing golden fragments are unchanged":
    $ git diff --name-status origin/main...HEAD | grep -iE 'golden|discovered/|testdata'
    A	examples/discovery-k8s/discovered/k8s-manifests.yaml
    A	tools/tokencost/testdata/expected.txt
    A	tools/tokencost/testdata/transcript.jsonl
    (additions only; no existing golden modified or deleted)

    $ go test ./internal/model/ ./internal/adapter/k8s/ ./cmd/noodle/ -count=1
    ok  	.../internal/model	0.282s
    ok  	.../internal/adapter/k8s	1.522s
    ok  	.../cmd/noodle	0.221s

## Findings

- Would the tests catch a revert? Yes for each requirement:
  - Without `object` in the schema, `schema-accepts` fails, because every item def is `additionalProperties: false`.
  - If the `required` list is dropped, `schema-rejects` fails.
  - If `o.identity()` is not wired in `elements.go`, `routes.go` or `connections.go`, `scenario`, `every-item` and `connection-object` fail on `Equal`.
  - If an implied namespace or provider gets an object by mistake, the `Equal` in `every-item` fails.
- Minor, breaks no criterion: in `internal/adapter/k8s/identity_test.go:116-138`, `every-item` does not assert the absence of a `field` key itself. It holds by construction for now: `internal/model/model.go:189-225` has no such field, the schema rejects extra properties, and the golden fragment has `0` matches. An explicit `NotContains` on the encoded fragment would make it hold as a test.
- Minor: the synthesized Ingress controller element named by FR-008 is not in the `every-item` example. The nil-object path it uses (`declared[ns].identity()` / `identity()` on a nil `*object`) is the same one the namespaces go through, and that path is tested.
- Scope: T01 (`89887eb`) touches only `internal/model`, `internal/adapter/k8s`, `schemas/v1alpha1/fragment.schema.json`, `tasks.md` and `PROGRESS.md`, which is within plan 009. No existing `features.json` line was removed. DSC-04 is still `passes: false`. No lint rule was relaxed (C04, C12).
