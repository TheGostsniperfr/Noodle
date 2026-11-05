# Evidence · PRJ-04

- Feature: PRJ-04
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 9adb43a
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | OpenCode project setup | PASS | `go test ./cmd/noodle/ -run Setup_OpenCode -v` exit 0, `--- PASS: TestSetup_OpenCode/project`. The test checks `stdout == wroteLines(openCodeWant)`, which is the five listed paths in order. It also checks that the tree equals `testdata/setup/opencode` (setup_test.go:118-119). A manual run of the built binary printed the same five `wrote` lines. |
| 2 | Codex project setup in the documented format | PASS | `-run Setup_Codex` exit 0, `--- PASS: TestSetup_Codex/project`, with stdout and golden asserted (setup_test.go:141-142). The grep finds `2:name = "fact-finder"`, `3:description = ...` and `4:developer_instructions = """`. I also parsed all three golden TOMLs with Python `tomllib`: each has exactly the keys `description`, `developer_instructions` and `name`. |
| 2b | Written agents carry the real instructions | PASS | `-run 'Setup_(OpenCode\|Codex)/instructions'` gives `--- PASS` for both parents and for `diagram-reviewer`, `fact-finder` and `icon-curator` under each. The test reads `agents/<name>.md`, strips the frontmatter and runs `assert.Contains` on it. I also compared the TOML-decoded `developer_instructions` with the source body: `True` for all three. |
| 3 | `-user` scope | PASS | `-run Setup_User` exit 0, with subtests `opencode` and `codex`. They set `HOME` to `t.TempDir()`, call `setupCommand` (the real `os.UserHomeDir` path) and assert `~/.agents/skills/...`, `~/.config/opencode/agents/fact-finder.md` and `~/.codex/agents/fact-finder.toml`, with 5 files in total. A manual `HOME=$T/h noodle setup codex -user` printed five `wrote ~/...` lines and exited 0. |
| 4 | Idempotent per host | PASS | `-run Setup_Idempotent` exit 0, with subtests `opencode` and `codex`. The test sets every mtime to 2000-01-01, reruns setup, and asserts empty stdout, equal bytes and unchanged mtime (setup_test.go:183-199). A manual second run printed nothing and exited 0. |
| 5 | Foreign files are kept | PASS | `-run Setup_KeepsForeign` exit 0, with 3 subtests. Each asserts the bytes are unchanged and checks stderr for `kept <path>: it was not written by noodle setup`. Exit code 1 is asserted by `TestSetup_ExitsOneAndKeepsAForeignFile` (passes under `-run Setup`). I also saw it by hand: a foreign `.codex/agents/fact-finder.toml` stayed `mine`, `foreign exit=1`. See Findings. |
| 6 | Read-only agents deny edits | PASS | The grep finds `mode: subagent` and `edit: deny` in `fact-finder.md` (lines 3 and 5) and `diagram-reviewer.md` (lines 3 and 5). `icon-curator.md` has only `3:mode: subagent`. |
| 7 | The written skill still passes the skill rules | PASS | `-run Setup_OpenCode/skill_rules` gives `--- PASS`. It runs `assert.Empty(skillrules.Check(...))` on the written SKILL.md. This is the same `skillrules.Check` that `tools/speccheck/skills.go` calls. |
| 8 | Formats are sourced | PASS | The grep on plan.md finds `opencode.ai/docs` (line 27), `developers.openai.com/codex` (line 28) and `developer_instructions` (line 35). |
| 9 | README installs per host | PASS | `## Install` is at line 23. Below it are `### Claude Code` (29), `### OpenCode` (48), `### Codex` (62) and `### Nix` (73). The next `## ` is `## Usage` at line 82. |

## Verify step

    $ go test ./cmd/noodle/ -run Setup -v -count=1
    --- PASS: TestSetup_OpenCode (0.00s)
        --- PASS: TestSetup_OpenCode/project (0.00s)
        --- PASS: TestSetup_OpenCode/instructions (0.00s)
        --- PASS: TestSetup_OpenCode/skill_rules (0.00s)
    --- PASS: TestSetup_Codex (0.00s)
        --- PASS: TestSetup_Codex/project (0.00s)
        --- PASS: TestSetup_Codex/instructions (0.00s)
    --- PASS: TestSetup_User (0.00s)
    --- PASS: TestSetup_Idempotent (0.00s)
    --- PASS: TestSetup_MarkerDoesNotBreakTheSkillRules (0.00s)
    --- PASS: TestSetup_RejectsBadArguments (0.00s)
    --- PASS: TestSetup_ReadsFlagsAndDirInAnyOrder (0.00s)
    --- PASS: TestSetup_KeepsForeign (0.00s)
    --- PASS: TestSetup_ReplacesAFileItWroteWhenTheVersionChanges (0.00s)
    --- PASS: TestSetup_ExitsOneAndKeepsAForeignFile (0.21s)
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.222s
    EXIT=0

    $ noodle setup codex -project q   # q/.codex/agents/fact-finder.toml holds "mine"
    kept .codex/agents/fact-finder.toml: it was not written by noodle setup
    setup: 1 file(s) kept
    foreign exit=1

Would the tests catch a revert? I checked one test per requirement:
- **FR-005:** `TestSetup_OpenCode/project` compares the written tree with committed goldens. It would fail if the frontmatter or paths changed, or if setup were removed.
- **FR-006:** `TestSetup_Codex/instructions` compares the source agent bodies with the written TOML. It would fail if `developer_instructions` were dropped or emptied.
- **FR-007:** `TestSetup_KeepsForeign` asserts that a foreign file keeps its exact bytes. It would fail if the marker check in setup.go:138 were removed.

Scope: the PRJ-04 commits are ce718ea, d63f954 and a79a00d. They touch `content.go`, `cmd/noodle/setup*.go`, the goldens, `cmd/noodle/main.go`, `README.md`, `internal/skillrules` (+7 lines) and spec/PROGRESS bookkeeping. All of these are in the plan's file table. I found no edits to features.json entries and no lint relaxation (C04, C12). There is no visual change, so C13 does not apply.

## Findings

- `contract.md` PRJ-04 #5: the "Verified by" (`-run Setup_KeepsForeign`) does not select the test that asserts exit code 1. The threshold's exit-1 part is proven by `TestSetup_ExitsOneAndKeepsAForeignFile` and by my manual run, not by the listed command. Fix either:
  - rename that test to `TestSetup_KeepsForeign_ExitsOne`, or
  - add `TestSetup_ExitsOne...` to the criterion's `-run` pattern.
- `cmd/noodle/setup.go:190`: `markerRe` is multiline and matches the marker on any line. A user's own file that quotes the marker line anywhere in its body would be treated as noodle's and overwritten. Anchoring the match to the position where `markMarkdown`/`codexAgent` put it would close this. FR-007 says "lacks the marker line", so this breaks no criterion.
- `cmd/noodle/setup.go:254-259`: the Codex agents for `fact-finder` and `diagram-reviewer` are read-only only through their instructions. No sandbox setting is written. This is deliberate and recorded in plan.md lines 34-39, because Codex does not let an agent file override the parent's sandbox. It is noted here so the README or the how-to does not promise enforced read-only on Codex.
