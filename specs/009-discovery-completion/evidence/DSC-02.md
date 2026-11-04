# Evidence · DSC-02

- Feature: DSC-02
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 05ed45d
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The test passes | PASS | `go test ./internal/adapter/k8s/ -run TestAddresses -v -count=1` exits 0 and ends with `ok  github.com/TheGostsniperfr/Noodle/internal/adapter/k8s 0.021s` |
| 2 | The full matrix of FR-003 forms | PASS | The `grep -cE -- '--- PASS: TestAddresses/(svc-ns\|svc-ns-svc\|svc-ns-svc-cluster-local)/(port\|noport)/(scheme\|noscheme)/(configmap\|env\|cr)'` count prints `36`. In `addresses_test.go`, each subtest runs `assert.Equal(t, []string{h.want}, inferredLines(f))`. `inferredLines` keeps only connections with `c.Inferred` set and prints `from -> to :port line object`. So each subtest checks exactly one inferred connection, on port `5432`, at the value's line: 5 for the ConfigMap, 12 for env, 6 for the CR. |
| 3 | The port rule without a port in the value | PASS | `--- PASS: TestAddresses/noport-multiport`. The ports are `metrics` 9090 and `http` 8080. The test expects `... -> k8s:data/service/postgres :http 10 Deployment/web`, and `model.go:107` says "Port names a port of To", so `:http` is the 8080 port. A companion case, `noport-multiport-without-http-takes-the-first`, checks the fallback to the first port. |
| 4 | Unknown hosts are not guessed | PASS | `--- PASS: TestAddresses/unknown-service` and `--- PASS: TestAddresses/external-host`. Both assert `assert.Empty(inferredLines(f))` and an `ElementsMatch` on the full list of ids: the input's 4 elements, none added. They cover an unknown Service, the wrong namespace, a port not on the Service, an external FQDN and `postgres.data.example.com`. |
| 5 | Public hosts map to their entry point | PASS | `route-hostname` expects exactly `[web -> k8s:infra/gateway/shared :https 27 Deployment/web]`. `ingress-host` expects exactly `[web -> k8s:_cluster/ingressclass/nginx : 19 Deployment/web]`, which is the controller element `routes.go:124` builds. Both are `Equal` on a one-item list, and both pass. |

## Verify step

    $ go test ./internal/adapter/k8s/ -run TestAddresses -v -count=1
    ...
        --- PASS: TestAddresses/unknown-service (0.00s)
        --- PASS: TestAddresses/external-host (0.00s)
        --- PASS: TestAddresses/route-hostname (0.00s)
        --- PASS: TestAddresses/ingress-host (0.00s)
        --- PASS: TestAddresses/an-unserved-host-yields-nothing (0.00s)
    PASS
    ok  	github.com/TheGostsniperfr/Noodle/internal/adapter/k8s	0.021s

    $ go test ./internal/adapter/k8s/ -run TestAddresses -v -count=1 | grep -cE -- '--- PASS: TestAddresses/(svc-ns|svc-ns-svc|svc-ns-svc-cluster-local)/(port|noport)/(scheme|noscheme)/(configmap|env|cr)'
    36

    $ go vet ./...                      # exit 0
    $ gofmt -l internal cmd tools       # no output
    $ go test ./... -count=1            # every package ok
    $ make examples                     # exit 0; discovery-k8s fragment and drift.txt diff clean
    $ make spec-check                   # exit 0

Would the tests fail if the feature were reverted? Yes. Without `addresses(idx, g, uses)` in `k8s.go`, `inferredLines` returns an empty list. All 36 matrix subtests and both entry-point subtests would then fail their `Equal`. Removing the `http` preference in `addressPort` would fail `noport-multiport`. Removing the Secret exclusion or the ConfigMap reader filter would fail `a-secret-is-never-scanned` and `a-config-key-reads-that-key-only`.

Scope: commit b5936e0 touches only `internal/adapter/k8s/` (`addresses.go`, `connections.go`, `k8s.go`, `routes.go`, tests), plus `tasks.md` and `PROGRESS.md`. This matches plan 009's package table. No lint rule changed, `docs/features.json` is untouched and `DSC-02` is still `passes: false`. No C04 or C12 breach.

## Findings

- `internal/adapter/k8s/addresses.go:113-120`: the Ingress `entryPoint` has no `ports`. The FR-004 hairpin to the Ingress controller is therefore emitted with an empty port (`: 19` in the `ingress-host` expectation). C02 puts the listening port on the server. A fix is to give the IngressClass element its 80/443 ports and resolve them through `entryPortName`, as the Gateway path does. No criterion is broken.
- `internal/adapter/k8s/addresses.go:258-266`: a custom resource whose spec names a Service is emitted as a `backend` element, with `Tech` set to its Kind. `kinds.go:51` still counts that Kind as `unknown-kind`, so the same object is both modelled and reported as unknown. I read this from the code and did not run it. A fix is to skip kinds that `addresses` turned into elements in `unknownKinds`, or to accept this and record it.
- `internal/adapter/k8s/addresses_test.go:16-30`: `inferredLines` checks `Src.Line` but not `Src.File`. A wrong file in a connection's `src` would still pass the matrix. A fix is to add the file to the printed line. No criterion is broken: criterion 2 asks for the `src` line.
