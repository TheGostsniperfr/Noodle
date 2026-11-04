# Contract 009 · Discovery, phase 6a completion

What "done" means for each feature, agreed before implementation (ADR-0024). One failed
criterion fails the feature. Every "Verified by" runs inside the evaluator's allowlist
(`EVALUATOR` in the Makefile). In a table cell, `\|` is the Markdown escape of a pipe:
run every command with a plain `|`.

- **Assumes the recommended options of the five open questions**: Q1 a (Claude runs the
  arms as fresh `claude -p` processes, transcripts committed), Q2 a (`-kustomize` shells
  out to `kustomize`), Q3 a (additive `object` and `field` in `v1alpha1`), Q4 a (rules in
  `<system>/discovery.yaml`), Q5 a (sources kept by the merge, card in batch F). A
  different answer reopens the criteria it touches before implementation: Q1 DSC-08 and
  DSC-12, Q2 DSC-03 #3, Q3 DSC-04, Q4 DSC-10, Q5 DSC-06 #2. The ADRs of Q2, Q3 and Q4 are
  written at the start of wave 1.
- Agreed: pending `make review-contract SPEC=009` (three rounds at most, PROCESS.md).
- Round 1 (2026-11-04): CHANGES, 33 points, all applied.

Test names below are the ones the implementation must use. Each named test is
table-driven and asserts with `Equal` on the whole output it checks (a full reference
list, a full stdout), so extra or duplicate items fail it. The example
`examples/discovery-k8s` has namespaces `prod` and `data`; Deployment `worker` is
deliberately matched by no curated element.

## DSC-01 · Configuration references

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | One fixture per reference kind of FR-001 and FR-002 | `go test ./internal/adapter/k8s/ -run TestReferences -v` | exit 0, with subtests `env-from`, `secret-key`, `config-key`, `mounts-secret`, `mounts-configmap`, `mounts-pvc`, `externalsecret-target`, `externalsecret-store`, `clusterstore-provider`, `store-provider-1password`, `store-provider-vault`, `store-provider-aws`, `argocd-syncs`; each asserts with `Equal` the fragment's full `references` list, `src` lines included |
| 2 | Providers are typed | same | the three `store-provider-*` subtests assert an element of kind `external` with tech `1Password`, `Vault`, `AWS` |
| 3 | No Secret value reaches a fragment, for a fixture and for every input | `go test ./internal/adapter/k8s/ -run 'TestReferences_NoSecretValue\|TestReferences_Property_NoSecretValue' -v` | exit 0: the fixture with `data` and `stringData` yields no value nor its base64 form in the fragment; the property test (`testing/quick`, `MaxCount` ≥ 200) does the same over generated Secrets with random values |

## DSC-02 · Addresses

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The full matrix of FR-003 forms | `go test ./internal/adapter/k8s/ -run TestAddresses -v \| grep -cE -- '--- PASS: TestAddresses/(svc-ns\|svc-ns-svc\|svc-ns-svc-cluster-local)/(port\|noport)/(scheme\|noscheme)/(configmap\|env\|cr)'` | exit 0 for the test run, count `36`; each subtest asserts with `Equal` one connection with `inferred: true`, the port of plan 009's port rule, and the value's `src` line |
| 2 | The port rule without a port in the value | same run, subtest `noport-multiport` | a Service with ports `metrics` 9090 and `http` 8080 yields the connection on 8080 |
| 3 | Unknown hosts are not guessed | same run, subtests `unknown-service` and `external-host` | no connection and no new element |
| 4 | Public hosts map to their entry point | same run, subtests `route-hostname` and `ingress-host` | exactly one inferred connection from the reading workload to the Gateway of the HTTPRoute, resp. to the Ingress controller element |

## DSC-03 · Inputs

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The namespace warning | `go test ./cmd/noodle/ -run TestNamespace -v` | exit 0: subtest `mixed` prints exactly `3 objects name no namespace; pass -namespace` on stderr; subtest `all-named` prints no such line; subtest `namespace-flag` with `-namespace web` prints no such line and gives the 3 objects ids under `k8s:web/` |
| 2 | `-argocd-app` sets the default namespace | same, subtest `argocd-app` | the 3 objects get ids under `k8s:web/` |
| 3 | `-kustomize` renders read-only, passes `-enable-helm`, names the fragment | `go test ./internal/adapter/k8s/ ./cmd/noodle/ -run TestKustomize -v` | exit 0, with a fake `kustomize` on PATH that records its argv: subtest `render` discovers its output and the input directory's file hashes are unchanged; `enable-helm` asserts argv `build <DIR> --enable-helm`; `fragment-name` asserts the file written is `discovered/k8s-shop.yaml`; `absent` (no `kustomize` on PATH) yields one `unrendered` entry with hint `kustomize` |
| 4 | `src` paths are relative | `go test ./internal/adapter/k8s/ -run TestSrcPaths -v`; `grep -rcE '(^\|[ {,-])file: */' examples/discovery-k8s/discovered/` | exit 0, subtests `manifests` and `kustomize` assert that every `src.file` starts with the input directory relative to the working directory, none starts with `/`, none contains `os.TempDir()`; grep gives `0` for every file |

## DSC-04 · Identity and field paths

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The contract: top-level `object`, `field` on references, `src.object` kept | `go test ./internal/model/ -run TestIdentity -v` | exit 0: subtests `schema-accepts` (an item with `object {apiVersion, kind, namespace, name}`, a reference with `field`), `schema-rejects` (an `object` without `kind`, one without `name`), `src-object-unchanged` (a fragment using only the string `src.object` validates), `legacy-fragment` (a fragment committed before this spec loads and marshals back byte-identical) |
| 2 | The adapter fills them everywhere | `go test ./internal/adapter/k8s/ -run TestIdentity -v` | exit 0: subtest `scenario` asserts FR-008's scenario exactly; subtest `every-item` asserts that in the fragment of `examples/discovery-k8s/manifests` 0 items lack `object` and 0 references lack `field` |
| 3 | Existing fixtures are not rewritten | `git diff --name-only --diff-filter=MD origin/main...HEAD -- internal/adapter/k8s/testdata cmd/noodle/testdata` | no output |

## DSC-05 · The example

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Golden from manifests, twice, and from a List | `go test ./cmd/noodle/ -run TestDiscoverExample -v` | exit 0, subtests `manifests-first`, `manifests-second`, `kubectl-list`, each byte-equal to `examples/discovery-k8s/discovered/k8s-manifests.yaml` |
| 2 | Determinism for every input | `go test ./internal/adapter/k8s/ -run TestDiscover_Property_Deterministic -v` | exit 0: `testing/quick`, `MaxCount` ≥ 200, over generated manifest sets in shuffled order, two runs byte-identical |
| 3 | The example has what spec 003's plan lists | `grep -h '^kind:' examples/discovery-k8s/manifests/*.yaml \| sort \| uniq -c`; `grep -c 'matches:' examples/discovery-k8s/model.yaml`; `grep -c 'status: planned' examples/discovery-k8s/model.yaml` | at least Gateway 1, HTTPRoute 2, Deployment 3, StatefulSet 1, Secret 1, ExternalSecret 1, NetworkPolicy 2; matches ≥ `2`; planned = `1` |
| 4 | `make examples` reruns and diffs | `make examples`; `grep -cE 'discover k8s examples/discovery-k8s/manifests' Makefile`; `grep -cE '(diff\|cmp).*examples/discovery-k8s/discovered/k8s-manifests.yaml' Makefile` | exit 0; `1`; `1` |

## DSC-06 · Merge

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | ADR-0015 merge rules | `go test ./internal/graph/ -run TestMerge -v` | exit 0, subtests `fills-empty-field`, `keeps-curated-field`, `maps-connection-to-curated-ids`, `takes-list-whole`, each asserting with `Equal` the merged element or connection |
| 2 | Sources survive the merge | same, subtest `keeps-sources` | the merged element lists the `src` and `object` of each matched item |
| 3 | Commands load the merged graph | `go test ./cmd/noodle/ -run TestLoad_MergesFragments -v` | exit 0: loading `examples/discovery-k8s` through the command loader yields the merged `api` element with port 8080 from the fragment |
| 4 | Curated fields never overridden, for every input | `go test ./internal/graph/ -run TestMerge_Property -v` | exit 0: `testing/quick`, `MaxCount` ≥ 500; the generator sets at least one scalar and one list field on each curated element; the test asserts every curated field is equal after the merge, and that at least one case had a field set by both sides |

## DSC-07 · Drift

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The four drift families | `go test ./internal/graph/ -run TestDrift -v` | exit 0, subtests `unmatched`, `gone`, `planned-absent`, `unmodelled-connection`, each asserting with `Equal` the entry and its `src` |
| 2 | Cross-fragment resolution | `go test ./internal/graph/ -run TestCrossFragment -v` | exit 0: a route in one fragment and its Gateway in another are linked, and no `missing-backend` remains for that route |
| 3 | The example's report, byte for byte | `go test ./cmd/noodle/ -run TestDrift -v` | exit 0: `noodle drift examples/discovery-k8s` stdout equals `examples/discovery-k8s/drift.txt`, and the command exits 0 |
| 4 | The report is not empty and holds FR-011's scenario | `grep -cE '^(unmatched\|gone\|planned\|unmodelled)' examples/discovery-k8s/drift.txt`; `grep -cE 'worker.*manifests/.*:[0-9]+' examples/discovery-k8s/drift.txt` | `4`, each heading followed by at least one entry line; `1`, under `unmatched` |

## DSC-08 · Token cost and checkpoint

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | The transcript reader prints all four measures | `go test ./tools/tokencost/ -v`; `grep -cE '^(category\|model\|turns\|wall)' tools/tokencost/testdata/expected.txt` | exit 0, the test asserting with `Equal` the full stdout on the fixture against `expected.txt`; each of the four prefixes at least once |
| 2 | The checkpoint rows and transcripts | `grep -cE '^\| [AC] \| \$[0-9]+\.[0-9]{2} \| [0-9]+/25 \| docs/benchmarks/transcripts/[^ ]+\.jsonl \|' docs/benchmarks/discovery-checkpoint.md`; `ls` of each transcript path | `2`; each path exists |
| 3 | The decision follows ADR-0018's rule | `cat docs/benchmarks/discovery-checkpoint.md` | the file states the rule as numbers (C's cost at most 70 % of A's, C's edges at least A's), and exactly one line `Decision: go` or `Decision: stop` that agrees with the two rows |
| 4 | PROGRESS records cost and recall of both arms (P6a-08) | `grep -cE 'checkpoint.*A.*\$[0-9.]+.*[0-9]+/25.*C.*\$[0-9.]+.*[0-9]+/25.*(go\|stop)' docs/PROGRESS.md` | `1` |

## DSC-09 · Rules after the checkpoint

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Each rule has its fixtures | `go test ./internal/adapter/k8s/ -run 'TestNetworkPolicy\|TestSecretEndpoint\|TestOperatorRules' -v` | exit 0, subtests `netpol-deny`, `netpol-deny-route`, `netpol-egress`, `netpol-allow`, `secret-endpoint-env`, `secret-endpoint-envfrom`, `secret-endpoint-externalsecret`, `secret-endpoint-ignores-password`, `cnpg-cluster`, `prometheus-alertmanager`, `certmanager-issuer`, `external-dns`, `servicemonitor`, `podmonitor`, each asserting with `Equal` |
| 2 | Denials name their policy | same, `netpol-deny` | the blocked connection has `denied: true` and `enforced_by: k8s:data/networkpolicy/default-deny` |
| 3 | `secret-endpoint` names keys and readers, never values | same, `secret-endpoint-env` | value `db/DB_URL`, `about` equal to `[k8s:prod/deployment/api, k8s:prod/deployment/worker]`, no Secret value in the fragment; `secret-endpoint-ignores-password` yields no entry for key `PASSWORD` |
| 4 | Started only after `Decision: go` (FR-015) | anchor `git log --format=%h -S'Decision: go' --reverse -- docs/benchmarks/discovery-checkpoint.md \| head -1`; then, for the files `netpol.go`, `secretendpoint.go`, `operators.go`, `rules.go`, `refkinds.go` of `internal/adapter/k8s/` and their `_test.go`, `git log --format=%h <anchor>..HEAD -- <files>` and `git log --format=%h -- <files>` | the anchor exists and the two lists are equal; if the file says `Decision: stop` instead, `grep -c '^## Revision' specs/009-discovery-completion/spec.md` ≥ `1` and the full list is empty |

## DSC-10 · Rules file and reference kinds

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | A rule in `<system>/discovery.yaml` maps a kind | `go test ./internal/adapter/k8s/ -run TestRules -v` | exit 0: subtest `vaultsecret` asserts FR-019's scenario (14 references, no `unknown-kind` for `VaultSecret`); `template-ns`, `template-name` and `template-fanout` (N references for N `targetRefs`) assert with `Equal` |
| 2 | A malformed rule is refused | same, subtest `bad-template` | discovery fails with an error matching `^.*discovery\.yaml:[0-9]+: ` followed by the template, asserted with `Regexp` |
| 3 | The six reference kinds of FR-020 | `go test ./internal/adapter/k8s/ -run TestReferenceKinds -v` | exit 0, subtests `owner`, `selects`, `runs-as`, `bound-by`, `scales`, `issues`, each asserting `{from, to, kind, field, src.line}` with `Equal` |
| 4 | Started only after `Decision: go` | DSC-09 #4, which covers `rules.go` and `refkinds.go` | same threshold |

## DSC-11 · Query

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Neighbours, filtered and not | `go test ./cmd/noodle/ -run TestQuery -v` | exit 0: subtest `neighbours-kind` asserts stdout equal to a golden listing exactly `k8s:prod/deployment/api` and `k8s:prod/deployment/worker`, each line with kind `secret-key` and a `file:line`; `neighbours-all` runs without `-kind` against its own golden |
| 2 | Shortest path and no path | same, subtests `path` and `no-path` | `path` stdout equals its golden, each edge with kind and `src`; `no-path` prints exactly `no path` and exits 1 |

## DSC-12 · Skill and benchmark

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Discovery is step 1 of the skill | `grep -nE '^#{2,3} ' skills/noodle-diagram/SKILL.md`; `grep -n 'noodle discover' skills/noodle-diagram/SKILL.md` | the first `noodle discover` line lies between the first and the second workflow step headings |
| 2 | `fact-finder` gets only what is left; the rules file is documented | `grep -n 'unresolved' skills/noodle-diagram/SKILL.md`; `grep -c 'discovery.yaml' skills/noodle-diagram/reference.md`; `grep -c 'matches' skills/noodle-diagram/reference.md` | an `unresolved` line within the fact-finder step; ≥ `1`; ≥ `1` |
| 3 | The skill names the edges discovery never sees | `grep -niE 'never sees' skills/noodle-diagram/SKILL.md skills/noodle-diagram/reference.md` | at least one match listing outside-cluster, CI and secret-value edges |
| 4 | The skill's run reads no manifest (FR-022, P6a-06) | `grep -cE '"name":"Read".*examples/discovery-k8s/manifests/' docs/benchmarks/transcripts/skill-discovery-k8s.jsonl`; `grep -c 'noodle discover' docs/benchmarks/transcripts/skill-discovery-k8s.jsonl` | `0`; ≥ `1` |
| 5 | The benchmark rows and means | `grep -cE '^\| [ABC] \| [12] \| \$[0-9.]+ \| [0-9]+/[0-9]+ \| [0-9]+m[0-9]+s \| docs/benchmarks/transcripts/[^ ]+\.jsonl \|' docs/benchmarks/discovery-tokens.md`; `ls` of each transcript | `6`; each exists; the file states the mean recall of A and C, C ≥ A |
| 6 | The Notion publication | `grep -cE 'notion\.(so\|site)\|app\.notion\.com' docs/benchmarks/discovery-tokens.md`; `grep -c 'Checked on Notion by Brian' docs/PROGRESS.md` | ≥ `1`, with the page's last-edited date copied beside the URL; ≥ `1` |

## Immutable features carried by this spec

Flipped on their own `make evaluate`, whose evidence lives in `specs/003-static-discovery/evidence/`.

| Feature | Graded by |
|---|---|
| P6a-03 | DSC-05 #1 and #4 |
| P6a-04 | `go test ./internal/adapter/k8s/ -run TestExample_ManifestsEqualList -v`, exit 0: the fragment of `examples/discovery-k8s/manifests` equals byte for byte the fragment of `list.yaml`, so P6a-04's own verify (`./internal/adapter/k8s/...`) makes the comparison |
| P6a-05 | DSC-07 #3 and #4 |
| P6a-06 | DSC-12 #1 to #4 |
| P6a-07 | the committed `discovery-tokens.md` (DSC-12 #5) with the Notion URL and its last-edited date, and DSC-12 #6, since the evaluator cannot open Notion |
| P6a-08 | DSC-08 #2 to #4 |
| P6a-09 | DSC-09 #1 and #3 (env, envFrom and ExternalSecret targets) |
| P6a-10 | DSC-09 #1, operator subtests |
