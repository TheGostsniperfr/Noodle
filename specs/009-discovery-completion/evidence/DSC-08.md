# Evidence · DSC-08

- Feature: DSC-08
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 82ef5df
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The transcript reader prints all four measures | PASS | `go test ./tools/tokencost/ -v -count=1` → `ok` (exit 0). `TestRun_PrintsTheCostOfTheFixtureTranscript` compares the full stdout with `expected.txt` using `assert.Equal`. The `grep -c` counts on `expected.txt` are: `^category` 3, `^model` 2, `^turns` 1, `^wall` 1 |
| 2 | The checkpoint rows and transcripts | PASS | The contract's `grep -cE '^[\|] [AC] [\|] \$…/25 [\|] docs/benchmarks/transcripts/…jsonl [\|]$'` returns `2`. `ls` finds both `checkpoint-A.jsonl` and `checkpoint-C.jsonl` |
| 3 | Transcripts hold no tool results (C09) | PASS | `grep -c '"type":"tool_result"'` returns `0` for both files. A looser `grep -c tool_result` also returns `0`. Record types are `assistant` (A: 65, C: 107), `user` (1 each, the prompt only) and `cost-state` (1 each) |
| 4 | The decision follows ADR-0029 | PASS | The file cites ADR-0029 (2 matches). `grep -cE '^Decision: (go\|stop)$'` returns `1`, and the line is `Decision: stop`. $4.76 / $4.34 = 109.7 %, which is above 70 %, so the rule gives `stop`. ADR-0029 has `^status: accepted` = `1` |
| 5 | PROGRESS records cost and recall of both arms (P6a-08) | PASS | The contract's grep on `docs/PROGRESS.md` returns `1`, at line 30: `checkpoint on PAE platform-overview, arm A $4.34 24/25, arm C $4.76 25/25: Decision: stop` |

## Verify step

    $ go test ./tools/tokencost/ -v -count=1
    --- PASS: TestRun_PrintsTheCostOfTheFixtureTranscript (0.00s)
    --- PASS: TestRun_ReadsStdin_WhenNoFileIsGiven (0.00s)
    --- PASS: TestRun_Fails (0.00s)
    --- PASS: TestRun_CountsAMessageOnce_WhateverItsNumberOfContentBlocks (0.00s)
    --- PASS: TestRun_FailsOnTwoArguments (0.00s)
    ok  	github.com/TheGostsniperfr/Noodle/tools/tokencost	0.003s

    $ grep -E '^\| [AC] \||^Decision' docs/benchmarks/discovery-checkpoint.md
    | A | $4.34 | 24/25 | docs/benchmarks/transcripts/checkpoint-A.jsonl |
    | C | $4.76 | 25/25 | docs/benchmarks/transcripts/checkpoint-C.jsonl |
    Decision: stop

    $ go run ./tools/tokencost docs/benchmarks/transcripts/checkpoint-A.jsonl | tail -1
    total weighted=868013 cost=$4.34
    $ go run ./tools/tokencost docs/benchmarks/transcripts/checkpoint-C.jsonl | tail -1
    total weighted=952751 cost=$4.76

I reproduced both costs from the committed transcripts with the committed tool.

I recounted recall myself. I took the from/to pairs of `checkpoint-{A,C}.facts.yaml` and matched them against the 25 rows of `platform-overview-edges.md`:

- **Arm A: 24/25.** The only edge missing is #2 (`majoutes-api-prod → keycloak-prod`). A has the preprod edge, `majoutes-api-preprod → keycloak-preprod`.
- **Arm C: 25/25.** #2 is `prod-majoutes/majoutes-api → prod-majoutes/keycloak`.

Both counts match the table. Each transcript holds a single `sessionId`, which is consistent with one fresh `claude -p` per arm.

## Findings

These findings break no contract criterion:

- **The fixture is independently checkable.** I checked `tools/tokencost/testdata/expected.txt` by hand against `transcript.jsonl`, using weights of 1 / 1.25 / 0.1 / 5:
  - `msg_1` is split across two records and counts once: 2 + 12500 + 10000 + 5000 = 27502.
  - Bash is `msg_2` + `msg_4`: 23501 + 14326 = 37827.
  - Sonnet: 5 + 5000 + 15000 = 20005.
  - Wall time: 10:00:00 to 10:02:05.5 = 2m05s.

  The expected output is not just a snapshot of whatever the code printed. Reverting the reader would fail `TestRun_PrintsTheCostOfTheFixtureTranscript` on the whole-stdout `Equal`.
- **Recall is scored by hand, with no committed script.** The 24/25 and 25/25 figures are correct, but the next run cannot reproduce them mechanically. A small scorer under `docs/benchmarks/` would fix this: it would map facts ids to the 25 ground-truth ids and print `N/25`. This matters for DSC-12, which needs six runs.
- **Dollar costs rest on assumed prices.** `tools/tokencost/transcript.go:22` holds an assumed price per model family. The checkpoint file says so, and gives the CLI's own `total_cost_usd`: A $3.37, C $3.47, so C is 103 % of A. That figure also gives `stop`, so the decision does not depend on the assumed prices.
- **A message with two tools counts toward one category only.** In the fixture, `msg_4` has both Bash and Read and is counted only under `bash`. This choice is undocumented. It moves cost between categories, though not the total, and the per-tool breakdown in the "Measures" table depends on it.
- **Scope is clean.** The diff touches only `docs/benchmarks/`, `docs/PROGRESS.md` and the T09 checkbox in `tasks.md`. `docs/features.json` is unchanged (C04/C12), and `make features` exits clean. No secret values appear in the transcripts or the facts files (C09).
