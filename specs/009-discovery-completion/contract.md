# Contract 009 · Discovery, phase 6a completion

What "done" means for each feature, agreed before implementation (ADR-0024). One failed
criterion fails the feature. Every "Verified by" runs inside the evaluator's allowlist
(`EVALUATOR` in the Makefile). In a table cell, `\|` is the Markdown escape of a pipe:
run every command with a plain `|`.

- Agreed: pending `make review-contract SPEC=009` (three rounds at most, PROCESS.md).

Test names below are the ones the implementation must use; each named test is table-driven
and asserts exact outputs (`Equal`), so a passing test proves its threshold.

## DSC-01 · Configuration references

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | One fixture per reference kind of FR-001 and FR-002 | `go test ./internal/adapter/k8s/ -run TestReferences -v` | exit 0, with subtests `env-from`, `secret-key`, `config-key`, `mounts-secret`, `mounts-configmap`, `mounts-pvc`, `externalsecret-target`, `externalsecret-store`, `store-provider`, `argocd-syncs`, each asserting the exact reference `{from, to, kind}` and its `src` line |
| 2 | No Secret value reaches a fragment | `go test ./internal/adapter/k8s/ -run TestReferences_NoSecretValue -v` | exit 0: a fixture whose Secret holds `data` and `stringData` values yields a fragment in which none of those values, nor their base64 decodings, appears |
| 3 | The provider element is typed | same as #1 | `store-provider` asserts an element of kind `external` with tech `1Password` for a `onepassword` provider |

## DSC-02 · Addresses

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Every DNS form of FR-003 resolves to a known Service, inferred, on its port | `go test ./internal/adapter/k8s/ -run TestAddresses -v` | exit 0, with subtests for `svc.ns`, `svc.ns.svc`, `svc.ns.svc.cluster.local`, each with and without `:port` and `http://`, from a ConfigMap, an env var and a CR spec string; each asserts one connection with `inferred: true`, the Service port, and the value's `src` line |
| 2 | Unknown hosts are not guessed | same | the subtests `unknown-service` and `external-host` assert no connection and no new element |
| 3 | A public URL host maps to its Gateway | same | the subtest `route-hostname` asserts exactly one inferred connection from the reading workload to the Gateway of the HTTPRoute that lists the host |

## DSC-03 · Inputs

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The namespace warning | `go test ./cmd/noodle/ -run TestNamespace -v` | exit 0: an input with 3 objects without namespace and 2 in `web` prints exactly `3 objects name no namespace; pass -namespace` on stderr; an input where every object names one prints no such line |
| 2 | `-argocd-app` sets the default namespace | same | the subtest `argocd-app` asserts ids under `k8s:web/` for the 3 objects |
| 3 | `-kustomize` renders read-only, and absence is reported | `go test ./internal/adapter/k8s/ -run TestKustomize -v` | exit 0: with a fake `kustomize` on PATH (set by the test) the rendered objects are discovered and the input directory's file hashes are unchanged; with no `kustomize` on PATH the directory is one `unrendered` entry with hint `kustomize` |
| 4 | `src` paths are relative | `go test ./internal/adapter/k8s/ -run TestSrcPaths -v`; `grep -rc '^ *file: /' examples/discovery-k8s/discovered/` | exit 0; `0` for every file |

## DSC-04 · Identity and field paths

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | `object` and `field` are in the contract | `go test ./internal/model/ -run TestIdentity -v` | exit 0: the fragment schema accepts an item with `object` and a reference with `field`, and rejects an `object` missing `kind` or `name` |
| 2 | The adapter fills them | `go test ./internal/adapter/k8s/ -run TestIdentity -v` | exit 0: FR-008's scenario holds exactly, `field: spec.template.spec.volumes[2].secret.secretName` |
| 3 | Fragments without the fields are unchanged | `go test ./internal/adapter/k8s/ ./cmd/noodle/` | exit 0, golden fragments of existing tests unchanged in the diff of the feature's commits (`git diff origin/main...HEAD --stat -- internal/adapter/k8s/testdata` lists only added files) |

## DSC-05 · The example

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Golden from manifests, twice, and from a List | `go test ./cmd/noodle/ -run TestDiscoverExample -v` | exit 0, with subtests `manifests-first`, `manifests-second`, `kubectl-list`, each asserting byte equality with `examples/discovery-k8s/discovered/k8s-manifests.yaml` |
| 2 | The example has what spec 003's plan lists | `ls examples/discovery-k8s/manifests`; `grep -c 'matches:' examples/discovery-k8s/model.yaml`; `grep -c 'status: planned' examples/discovery-k8s/model.yaml` | manifests include a Gateway, two HTTPRoutes, a database StatefulSet, a Secret, an ExternalSecret and a NetworkPolicy; at least `2` matches; exactly `1` planned |
| 3 | `make examples` checks the golden file | `make examples` | exit 0, and `grep -n 'discovery-k8s' Makefile` shows the rerun and its diff in the `examples` recipe |

## DSC-06 · Merge

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | ADR-0015 merge rules | `go test ./internal/graph/ -run TestMerge -v` | exit 0, with subtests `fills-empty-field`, `keeps-curated-field`, `maps-connection-to-curated-ids`, `takes-list-whole`, each asserting the exact merged element or connection |
| 2 | Sources survive the merge | same | the subtest `keeps-sources` asserts the merged element lists the `src` and `object` of each matched item |
| 3 | Curated fields are never overridden, for every input | `go test ./internal/graph/ -run TestMerge_Property -v` | exit 0: a property test over generated models and fragments checks that every field set on a curated element is equal before and after the merge |

## DSC-07 · Drift

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The four drift families | `go test ./internal/graph/ -run TestDrift -v` | exit 0, with subtests `unmatched`, `gone`, `planned-absent`, `unmodelled-connection`, each asserting the exact entry with its `src` |
| 2 | Cross-fragment resolution | `go test ./internal/graph/ -run TestCrossFragment -v` | exit 0: a route in one fragment and its Gateway in another are linked, and no `missing-backend` remains for that route |
| 3 | The example's report | `go test ./cmd/noodle/ -run TestDrift -v` | exit 0: `noodle drift examples/discovery-k8s` output equals `examples/discovery-k8s/drift.txt` byte for byte, and the command exits 0 |

## DSC-08 · Token cost and checkpoint

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The transcript reader | `go test ./tools/tokencost/ -v` | exit 0: the trimmed fixture yields exactly the totals of its expected-output file |
| 2 | The checkpoint is recorded | `cat docs/benchmarks/discovery-checkpoint.md`; `grep -cE '^Decision: (go\|stop)$' docs/benchmarks/discovery-checkpoint.md` | one row per arm A and C with cost in dollars, edges found out of 25 and the transcript path; `1` |
| 3 | PROGRESS has the same decision | `grep -n 'checkpoint' docs/PROGRESS.md` | a line under the run's date naming the decision of #2 |

## DSC-09 · Rules after the checkpoint

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Each rule has its fixture | `go test ./internal/adapter/k8s/ -run 'TestNetworkPolicy\|TestSecretEndpoint\|TestOperatorRules' -v` | exit 0, with subtests `netpol-deny`, `netpol-allow`, `secret-endpoint`, `cnpg-cluster`, `prometheus-alertmanager`, `certmanager-issuer`, `external-dns`, `servicemonitor`, `podmonitor`, each asserting exact items |
| 2 | Denials name their policy | same, subtest `netpol-deny` | the blocked connection has `denied: true` and `enforced_by: k8s:data/networkpolicy/default-deny` |
| 3 | No value of a Secret in `secret-endpoint` | same, subtest `secret-endpoint` | the entry's value is exactly `db/DB_URL` and the fragment holds no Secret value |
| 4 | Started only after `Decision: go` | `git log --format=%h -1 -- docs/benchmarks/discovery-checkpoint.md`, then `git log --format=%h <that sha>..HEAD -- internal/adapter/k8s/netpol.go internal/adapter/k8s/secretendpoint.go internal/adapter/k8s/operators.go` and `git log --format=%h -- internal/adapter/k8s/netpol.go internal/adapter/k8s/secretendpoint.go internal/adapter/k8s/operators.go` | the checkpoint file says `Decision: go`; the two commit lists are equal |

## DSC-10 · Rules file and reference kinds

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | A rule maps a kind | `go test ./internal/adapter/k8s/ -run TestRules -v` | exit 0: FR-019's scenario holds exactly, 14 references and no `unknown-kind` entry for `VaultSecret` |
| 2 | A malformed rule is refused | same | the subtest `bad-template` fails discovery with a message naming the rule file and line |
| 3 | The six reference kinds of FR-020 | `go test ./internal/adapter/k8s/ -run TestReferenceKinds -v` | exit 0, with subtests `owner`, `selects`, `runs-as`, `bound-by`, `scales`, `issues`, each asserting the exact reference and its `field` |

## DSC-11 · Query

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Neighbours, filtered by kind | `go test ./cmd/noodle/ -run TestQuery -v` | exit 0: FR-021's scenario output equals the test's golden text |
| 2 | Shortest path | `go test ./internal/graph/ -run TestQuery_Path -v` | exit 0: a path query on the example returns one shortest path, each edge with its kind and `src`; an unreachable pair returns `no path` and exit 1 |

## DSC-12 · Skill and benchmark

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Discovery first in the skill | `grep -n 'noodle discover' skills/noodle-diagram/SKILL.md`; `grep -c 'matches' skills/noodle-diagram/reference.md` | a match inside the skill's first workflow step; at least `1` |
| 2 | The edge families discovery never sees are named | `grep -n -i 'never sees' skills/noodle-diagram/SKILL.md skills/noodle-diagram/reference.md` | at least one match listing outside-cluster, CI and secret-value edges |
| 3 | The benchmark is committed | `cat docs/benchmarks/discovery-tokens.md` | six rows (A, B, C, two runs each) with cost, recall and wall time; arm C's mean recall is at least arm A's |
| 4 | The publication is linked | `grep -c 'notion.so\|notion.site\|app.notion.com' docs/benchmarks/discovery-tokens.md` | at least `1`; the page itself is checked by Brian at the flip, outside the evaluator's reach |
