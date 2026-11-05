# Evidence · DSC-03

- Feature: DSC-03
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 05ed45d
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The namespace warning | PASS | `go test ./cmd/noodle/ -run TestNamespace -v` gives `--- PASS: TestNamespace` with subtests `mixed`, `all-named` and `namespace-flag`. `mixed` checks that the exact line `3 objects name no namespace; pass -namespace` is one of the stderr lines. `namespace-flag` checks the full element id list with `ElementsMatch`: `k8s:web/service/{a,b,c}`. Rerun by hand on stdin, stderr is the summary line plus `3 objects name no namespace; pass -namespace` |
| 2 | `-argocd-app` sets the default namespace, silently | PASS | Subtest `argocd-app` in the same run passes. The full id list has `k8s:web/service/{a,b,c}`, and `assert.NotContains(stderr, "name no namespace")` holds |
| 3 | `-kustomize` renders read-only, passes `-enable-helm` only when asked, names the fragment | PASS | `go test ./internal/adapter/k8s/ ./cmd/noodle/ -run TestKustomize -v` exits 0. k8s: `render` (argv `[build apps/shop]`, the file hashes of `apps/shop` are equal before and after), `enable-helm` (argv `[build apps/shop --enable-helm]`), `missing-tool`. cmd: `fragment-name` (`discovered/k8s-shop.yaml`, one entry in the directory), `missing-tool` (the error contains `kustomize` and `nix shell nixpkgs#kustomize`, and there is no `discovered/` dir), `hint-without-flag` (`Equal` on `[{unrendered overlay kustomize}]`, and stderr has the exact line `overlay: kustomization found, rerun: noodle discover k8s -kustomize overlay (needs kustomize: nix shell nixpkgs#kustomize)`) |
| 3b | The tool is in the dev shell | PASS | `grep -cE '^\s*packages = with pkgs; \[[^]]*\bkustomize\b' flake.nix` gives `1` (line 27) |
| 4 | `src` paths are relative | PASS | `go test ./internal/adapter/k8s/ -run TestSrcPaths -v` gives `--- PASS: TestSrcPaths/manifests` and `--- PASS: TestSrcPaths/kustomize`. Both pass absolute inputs and assert the prefix `manifests/`, resp. `apps/shop/`, no leading `/` and no `os.TempDir()`. `grep -rcE '(^\|[ {,-])file: *"?/' examples/discovery-k8s/discovered/` gives `k8s-manifests.yaml:0` |
| 5 | The decision is recorded | PASS | `grep -c '^status: accepted' docs/adr/0027-discovery-may-call-kustomize.md` gives `1`. `make spec-check` gives exit 0 |

## Verify step

    $ go test ./internal/adapter/k8s/ ./cmd/noodle/ -run 'TestNamespace|TestKustomize|TestSrcPaths' -count=1
    ok  	github.com/TheGostsniperfr/Noodle/internal/adapter/k8s	0.066s
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.037s

    $ go test ./internal/... ./cmd/... -count=1   # no regression
    ok  	.../internal/adapter/k8s	2.582s
    ok  	.../cmd/noodle	0.407s
    (all other packages ok)
    $ go vet ./cmd/... ./internal/...; gofmt -l cmd internal
    (no output, exit 0)

I checked whether each test would still pass with its code reverted. None would:
- Remove the warning in `discover.go`, and `mixed` fails.
- Remove `ArgoNamespace`, and `argocd-app` fails.
- Remove `kustomizeBuild`, and `render`, `enable-helm` and `missing-tool` fail.
- Remove `adapter.RelativePath` in `decode.go`, and `TestSrcPaths/manifests` fails, because it passes an absolute `Paths` entry.

The `docs/features.json` diff has no `-` line, so no existing entry was edited (C12). No lint rule changed (C04).

## Findings

- **Possible write into DIR with `-enable-helm`.** In `internal/adapter/k8s/kustomize.go:24-27`, real `kustomize build DIR --enable-helm` pulls charts into `DIR/charts/` by default (`helmGlobals.chartHome`). That would break FR-006's "without writing to DIR" and ADR-0027's read-only promise (C09). I did not run it: it needs helm and the network. The `enable-helm` subtest uses a fake tool and does not check hashes. No criterion fails, because criterion #3 requires argv to be exactly `build <DIR> --enable-helm`. Two possible fixes:
  - Render a copy of DIR made in a temp directory, then map `src` back to DIR.
  - Pass a temp `--helm-...` chart home. This needs a contract amendment to #3's argv rule.

  Either way, add a hash check to `enable-helm`.
- **`mixed` checks one stderr line, not the whole stderr.** In `cmd/noodle/discover_inputs_test.go:59`, `mixed` uses `assert.Contains` on the stderr lines. The contract preamble says tests assert with `Equal` on the whole output. The warning line itself is matched exactly, so #1 holds. But a duplicated warning line would not fail this test. Asserting `Equal` on the stderr lines after the summary line would close that gap.
- **`namespace.go` was never created.** `specs/009-discovery-completion/plan.md:35` puts FR-005 in it. The logic is in `decode.go` (counts) and `cmd/noodle/discover.go` (the warning), which is a small drift from the plan. Either update the plan or move the code.
- **Inputs outside the working directory.** `internal/adapter/adapter.go` `RelativePath` gives `../…` for an input outside the working directory. It falls back to the path as given, possibly absolute, when `Abs`, `Getwd` or `Rel` fails. FR-007 is met for the tested cases. No test covers an input outside the working directory.
