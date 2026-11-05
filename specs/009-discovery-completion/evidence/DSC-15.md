# Evidence · DSC-15

- Feature: DSC-15
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 30d9302
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | The rows and transcripts | PASS | `grep -cE '^[\|] [AC] [\|] \$…checkpoint3-[AC]\.jsonl [\|]$' docs/benchmarks/discovery-checkpoint-3.md` → `2`. `ls` shows `checkpoint3-A.jsonl` (253024 B) and `checkpoint3-C.jsonl` (380006 B). |
| 2 | Transcripts hold no tool results, prompt snapshot or account records | PASS | `grep -cE '"type": *"tool_result"\|"toolUseResult"\|prompt_snapshot\|credential_org\|session-owner' …checkpoint3-*.jsonl` → `checkpoint3-A.jsonl:0`, `checkpoint3-C.jsonl:0` |
| 3 | C got the digest and the honest paragraph | PASS | The `tool_use` lines that match `discovered/` → `0`. `grep -c 'never sees' checkpoint-prompt-c3.md` → `1`. `grep -c 'complete' checkpoint-prompt-c3.md` → `0`. |
| 4 | The decision follows ADR-0029 | PASS | The file cites ADR-0029 (1 match) and ADR-0032 (2 matches). It has exactly one line `Decision: stop`. Recomputed costs: C/A = $4.91 / $3.89 = 126 %, which is above 70 %, so `stop` is the decision the numbers give, even though C's edges (25) are at least A's (23). |
| 5 | PROGRESS records it | PASS | The grep returns `1`. The matching line is `- T29: checkpoint 3 on PAE platform-overview, arm A $3.89 23/25, arm C $4.91 25/25: Decision: stop, final (C at 126 %)…` |

## Verify step

    $ grep -cE '^[|] [AC] [|] \$[0-9]+\.[0-9]{2} [|] [0-9]+/25 [|] docs/benchmarks/transcripts/checkpoint3-[AC]\.jsonl [|]$' docs/benchmarks/discovery-checkpoint-3.md
    2
    $ grep -cE '^Decision: (go|stop)$' docs/benchmarks/discovery-checkpoint-3.md
    1
    $ grep -cE 'checkpoint 3.*platform-overview.*A.*\$[0-9.]+.*[0-9]+/25.*C.*\$[0-9.]+.*[0-9]+/25.*(go|stop)' docs/PROGRESS.md
    1

I recomputed the costs from the committed transcripts instead of trusting the table:

    $ go run ./tools/tokencost docs/benchmarks/transcripts/checkpoint3-A.jsonl
    turns 34
    wall 6m08s
    total weighted=778950 cost=$3.89
    $ go run ./tools/tokencost docs/benchmarks/transcripts/checkpoint3-C.jsonl
    turns 45
    wall 7m24s
    total weighted=981337 cost=$4.91

These match the cost rows and the Measures table: turns 34 and 45, wall times 6m08s and 7m24s.

I checked where each arm's prompt came from:

    # first user record of checkpoint3-C.jsonl == checkpoint-prompt.md + "\n\n" + checkpoint-prompt-c3.md
    prompt equal: True
    # checkpoint3-A.jsonl: 0 occurrences of "never sees" or "digest.txt" in its prompt
    $ git log --format='%h %ad' -- docs/benchmarks/checkpoint-prompt-c3.md
    82fbc2b 2026-11-05 15:05:46 +0100   # before the run commit 30d9302 (15:13:56)

I spot-checked recall against `platform-overview-edges.md`:
- `checkpoint3-A.facts.yaml` has no `alertmanager` at all. Its only Grafana edge goes to `github-raw`, so it has no Grafana → Prometheus edge. That supports A's 23/25 and the two edges it is said to miss.
- `checkpoint3-C.facts.yaml` has `prometheus-stack-grafana → …prometheus` (line 632), `prometheus → …alertmanager` (line 592) and `kubelet → docker` (line 798). That supports C's 25/25 on the edges A disputes.

The gates also pass. `make features` and `make spec-check` print no errors. A secret scan of the four committed transcript and facts files for private keys, AWS, GitHub and GitLab tokens and credentialed Postgres URLs returns `0` for every file.

## Findings

- `docs/benchmarks/discovery-checkpoint-3.md:9-12`: the 23/25 and 25/25 recall figures have no per-edge scoring table, so they cannot be recomputed mechanically. I could only spot-check the disputed edges against the facts files. Adding a 25-row found or missed column per arm would make the recall auditable, as `tokencost` already makes the cost auditable. This breaks no criterion.
- `docs/benchmarks/transcripts/checkpoint3-C.jsonl`: arm C made one `Read` of a Claude Code tool-results spill file under `~/.claude/projects/…/tool-results/`. That file is the arm's own spilled Bash output, not a `discovered/` fragment, so criterion 3 holds. The read does show the run's path layout in the committed transcript, which does no harm.
- Scope: the HEAD commit touches only the checkpoint file, the transcripts, the facts files, PROGRESS and tasks. The prompt change was committed earlier, in its own commit with ADR-0032, before the run. It changes only C's paragraph, as that ADR says. No `features.json` entry is edited for DSC-15; it is added with `passes: false` (C12).
