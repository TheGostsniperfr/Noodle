I've finished checking: every criterion reproduces, and I graded the edges myself. Here is the evidence file.

# Evidence · DSC-14

- Feature: DSC-14
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 2aecda6
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The rows and transcripts | PASS | The row grep on `discovery-checkpoint-2.md` gives `2`. `ls` lists both `transcripts/checkpoint2-A.jsonl` and `checkpoint2-C.jsonl`. `go run ./tools/tokencost` on each transcript gives `cost=$4.82` for A and `cost=$3.04` for C, the same as the rows. I matched each arm's `facts.yaml` connections against `platform-overview-edges.md`. A finds all 25 edges (17 `grafana -> prometheus`, 18 `prometheus -> alertmanager`, 21 `talos-node -> docker-hub`). C finds 22 and misses #17, #18 and #21. `argocd-repo-server -> docker-hub` does not count for #21. |
| 2 | Transcripts hold no tool results, prompt snapshot or account records | PASS | The grep gives `checkpoint2-A.jsonl:0` and `checkpoint2-C.jsonl:0`. The record types in A are `assistant` 82, `user` 1, `cost-state` 1. C has the same three types. |
| 3 | Arm C got the digest, not the fragments | PASS | The `tool_use` lines of C matching `discovered/` count `0`. Seven of C's `tool_use` lines name `digest.txt`. C's user prompt carries the `checkpoint-prompt-c2.md` paragraph. `run-arm.sh` round 2 writes `noodle digest` output to `$work/digest.txt` and does not copy the fragments. |
| 4 | The decision follows ADR-0029 | PASS | `grep -cE '^Decision: (go\|stop)$'` gives `1`, and that line is `Decision: stop`. The file cites ADR-0029 once and ADR-0031 once. C costs 3.04/4.82 = 63 % of A, so the cost half is met (2.39/3.66 = 65 % on the CLI's `totalCostUSD`). But 22 < 25 edges, so the rule gives `stop`. The decision is correct. |
| 5 | PROGRESS records it | PASS | The grep gives `1`, on `docs/PROGRESS.md:30`: `T23: checkpoint 2 on PAE platform-overview, arm A $4.82 25/25, arm C $3.04 22/25 with the digest: Decision: stop …` |

## Verify step

    $ grep -cE '^[|] [AC] [|] \$[0-9]+\.[0-9]{2} [|] [0-9]+/25 [|] docs/benchmarks/transcripts/checkpoint2-[AC]\.jsonl [|]$' docs/benchmarks/discovery-checkpoint-2.md
    2
    $ grep -E '^Decision' docs/benchmarks/discovery-checkpoint-2.md
    Decision: stop
    $ go run ./tools/tokencost docs/benchmarks/transcripts/checkpoint2-A.jsonl | tail -1
    total weighted=963021 cost=$4.82
    $ go run ./tools/tokencost docs/benchmarks/transcripts/checkpoint2-C.jsonl | tail -1
    total weighted=607929 cost=$3.04
    $ grep -nE 'checkpoint 2.*platform-overview' docs/PROGRESS.md
    30:- T23: checkpoint 2 on PAE platform-overview, arm A $4.82 25/25, arm C $3.04 22/25 with the digest: Decision: stop ...
    $ make features
    go run ./tools/speccheck features -base HEAD    (exit 0)

## Findings

None of these breaks a criterion.

- `docs/benchmarks/checkpoint-prompt-c2.md:2-3`: the C prompt tells the agent the digest "is complete for the connections between components inside the cluster". That is false, because the digest has no operator-implied edges or chart-value datasources. The prompt claim probably cost C edges #17 and #18. The checkpoint file says so honestly. Still, round 2 measures a digest plus a wrong claim, not the digest alone. If there is a round 3, it should drop the completeness claim, or note under the C arm that the claim was part of it.
- `docs/benchmarks/discovery-checkpoint-2.md:3`: the run says it used "the T09 snapshot", but no snapshot hash or commit is recorded. I cannot check from the repository that rounds 1 and 2 ran on the same input. A fix is to record the snapshot commit or a tree hash in the Protocol section.
- `docs/benchmarks/discovery-checkpoint-2.md:46-50`: arm A moved about 10 % between two identical runs. That is large next to the 7-point margin under the 70 % threshold. It does not change this decision, since recall decides it. A future `go` decided by cost alone would need more than one run per arm.
- The working tree holds an untracked `specs/009-discovery-completion/evidence/DSC-13.md`. It is not part of DSC-14 and does not affect this grade. It should be committed or removed before the PR.
