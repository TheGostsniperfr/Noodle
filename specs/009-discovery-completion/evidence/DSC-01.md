# Evidence · DSC-01

- Feature: DSC-01
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 05ed45d
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | One fixture per reference kind of FR-001 and FR-002 | PASS | `go test ./internal/adapter/k8s/ -run TestReferences -v -count=1` exits 0. All 13 required subtests appear as `--- PASS`: `env-from`, `secret-key`, `config-key`, `mounts-secret`, `mounts-configmap`, `mounts-pvc`, `externalsecret-target`, `externalsecret-store`, `clusterstore-provider`, `store-provider-1password`, `store-provider-vault`, `store-provider-aws`, `argocd-syncs`. Each runs `assert.Equal(t, tt.want, refLines(f))` (refs_test.go:176). That compares the full sorted reference list as `from -> to kind line`, so src lines are included. Two extra subtests also pass: `mounts-projected-and-init-container` and `two-keys-of-one-secret-are-one-reference`. |
| 2 | Providers are typed | PASS | Same run. The three `store-provider-*` subtests compare the full map of `external` elements with `Equal` (refs_test.go:154-162,184): `{Kind: "external", Tech: "1Password"}`, `"Vault"` and `"AWS"`. |
| 3 | No Secret value reaches a fragment, for a fixture and for every input | PASS | `go test ./internal/adapter/k8s/ -run 'TestReferences_NoSecretValue\|TestReferences_Property_NoSecretValue' -v` gives `--- PASS: TestReferences_NoSecretValue` and `--- PASS: TestReferences_Property_NoSecretValue (0.11s)`. The fixture has both `data` and `stringData`. The test checks the raw and base64 forms of each value with `NotContains` against the `EncodeFragment` output. The property test uses `quick.Config{MaxCount: 200}` (refs_test.go:275). |

## Verify step

    $ go test ./internal/adapter/k8s/ -run TestReferences -v -count=1
    --- PASS: TestReferences (0.00s)
        --- PASS: TestReferences/env-from (0.00s)
        --- PASS: TestReferences/secret-key (0.00s)
        --- PASS: TestReferences/config-key (0.00s)
        --- PASS: TestReferences/mounts-secret (0.00s)
        --- PASS: TestReferences/mounts-configmap (0.00s)
        --- PASS: TestReferences/mounts-pvc (0.00s)
        --- PASS: TestReferences/mounts-projected-and-init-container (0.00s)
        --- PASS: TestReferences/two-keys-of-one-secret-are-one-reference (0.00s)
        --- PASS: TestReferences/externalsecret-target (0.00s)
        --- PASS: TestReferences/externalsecret-store (0.00s)
        --- PASS: TestReferences/clusterstore-provider (0.00s)
        --- PASS: TestReferences/store-provider-1password (0.00s)
        --- PASS: TestReferences/store-provider-vault (0.00s)
        --- PASS: TestReferences/store-provider-aws (0.00s)
        --- PASS: TestReferences/argocd-syncs (0.00s)
    --- PASS: TestReferences_UnknownProviderIsReported_NotDrawn (0.00s)
    --- PASS: TestReferences_NoSecretValue (0.00s)
    --- PASS: TestReferences_Property_NoSecretValue (0.11s)
    ok  	github.com/TheGostsniperfr/Noodle/internal/adapter/k8s	0.125s

    $ go vet ./internal/adapter/k8s/ && go test ./internal/adapter/... ./cmd/noodle/ ./internal/model/ -count=1
    ok  	.../internal/adapter	0.020s
    ok  	.../internal/adapter/k8s	6.658s
    ok  	.../cmd/noodle	0.736s
    ok  	.../internal/model	0.826s

    $ make spec-check; make features
    (both exit 0, no reported problems)

End to end: `examples/discovery-k8s/discovered/k8s-manifests.yaml` contains `secret-key`, `target`, `store` and `provider` references with `object` and relative `src`. The manifest Secret value `cGxhY2Vob2xkZXI=` does not appear in the fragment.

## Tests that would catch a revert

- FR-001: `TestReferences/secret-key` and `mounts-secret` use `Equal` on the full list. If `configRefs` were removed, the list would be empty and both would fail. `TestReferences_NoSecretValue` also expects exactly the `mounts` and `secret-key` references.
- FR-002: `externalsecret-target`, `externalsecret-store`, `store-provider-*` and `argocd-syncs` use `Equal` on references and provider elements. If `externalSecrets` or `applications` were removed, they would fail.
- No Secret value: both tests read the encoded fragment. If an emitter wrote `data` or `stringData`, `NotContains` would fail.

## Scope and constitution

- Commit 1ada04d touches only `internal/adapter/k8s/` (argocd.go, decode.go, externalsecrets.go, k8s.go, nodes.go, refs.go, refs_test.go), plus `docs/PROGRESS.md` and `tasks.md`.
- The diff to `docs/features.json` only adds entries. It edits or deletes none, and DSC-01 still has `passes: false` (C12).
- No lint rule is relaxed (C04).

## Findings

These findings break no contract criterion.

- **Dedup vs. "one per use"** (`internal/adapter/k8s/connections.go:50-60`, `refs.go:17`): `graph.refer` deduplicates on (from, to, kind). Two `secretKeyRef` uses of one Secret therefore give a single reference, and subtest `two-keys-of-one-secret-are-one-reference` locks that in. FR-001 says "one reference per use". The behaviour follows from the content-hashed reference id. Fix: change FR-001's wording to "one reference per workload, target and kind", or record the choice in an ADR, so spec and code agree.
- **Fallback src line** (`internal/adapter/k8s/externalsecrets.go:21-22`): when an ExternalSecret has no `spec.target`, `firstLine` returns 0. The `target` reference then cites line 1 of the object, which is the `apiVersion` line (subtest `externalsecret-store`, `... target 1`). Citing the `metadata.name` line would point the reader at the name the target defaults to.
- **`src.file` not asserted here** (`internal/adapter/k8s/refs_test.go:22-29`): `refLines` compares from, to, kind and line, but not `src.file`. That meets "src lines included", and DSC-03's `TestSrcPaths` covers the file. Still, a wrong `file` on configuration references would not fail `TestReferences`.
