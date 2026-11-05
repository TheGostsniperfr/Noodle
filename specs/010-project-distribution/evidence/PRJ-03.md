# Evidence · PRJ-03

- Feature: PRJ-03
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: db81c22
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Each rule has a case with its rule id | PASS | `go test ./tools/speccheck/ -run Skills -v` exits 0, also uncached with `-count=1`. `--- PASS` for `name_not_folder`, `name_uppercase`, `name_too_long`, `description_empty`, `description_too_long`, `unknown_key`, `valid`, `claude_name_under_heading`, and for `claude_name_outside_heading/<literal>` over all six literals. `skills_test.go:44-46`: `require.Len(findings, 1)`, `Contains(path)`, `Contains(rule+":")`. `skills_test.go:85-88` does the same with `claude-only-name:`. The `valid` and `claude_name_under_heading` cases assert 0 findings. |
| 2 | `all` fails on a broken skill | PASS | `go test ./tools/speccheck/ -run 'TestRun_Exits.*/broken_skill' -v` prints `--- PASS: TestRun_ExitsWithFindingsThroughTheCLI/broken_skill`. It runs `run([... "all"])` and asserts `ErrorIs(err, errFindings)` with the exact stderr line `skills/demo/SKILL.md: description-empty: ...` (`speccheck_test.go:412-426`). |
| 3 | The repository's skills pass | PASS | `make spec-check` runs `go run ./tools/speccheck all` and exits 0. |
| 4 | Claude Code-only names sit under their heading | PASS | The literal grep matches only lines 61, 62 and 63. Headings: `58:## In Claude Code` and `65:## Semantics`. All matches sit between line 58 and the next `##` heading. The fenced block at lines 37-46 holds no `#` lines that the heading grep would pick up. |

## Verify step

    $ go test ./tools/speccheck/ -run Skills -v
    --- PASS: TestSkills_ChecksEachRule (0.00s)
        --- PASS: TestSkills_ChecksEachRule/name_not_folder
        --- PASS: TestSkills_ChecksEachRule/description_too_long
        --- PASS: TestSkills_ChecksEachRule/unknown_key
        --- PASS: TestSkills_ChecksEachRule/description_empty
        --- PASS: TestSkills_ChecksEachRule/name_too_long
        --- PASS: TestSkills_ChecksEachRule/valid
        --- PASS: TestSkills_ChecksEachRule/name_uppercase
    --- PASS: TestSkills_FailsAllForABrokenProcessSkill/process_skill_broken
    --- PASS: TestSkills_KeepsClaudeCodeNamesUnderTheirHeading/claude_name_outside_heading/{noodle:fact-finder,noodle:diagram-reviewer,noodle:icon-curator,the_Agent_tool,Task_tool,/plugin}
    --- PASS: TestSkills_KeepsClaudeCodeNamesUnderTheirHeading/claude_name_under_heading
    PASS
    ok  github.com/TheGostsniperfr/Noodle/tools/speccheck  0.135s   (-count=1)

    $ make spec-check
    == spec-check
    go run ./tools/speccheck all
    exit=0

    $ gofmt -l tools/speccheck internal/skillrules   (no output)
    $ go vet ./tools/speccheck/ ./internal/skillrules/
    vet-ok

Reverting the feature would make these tests fail. Every `ChecksEachRule` case calls `Repo.CheckSkills` on a fixture and requires exactly one finding with its rule id, so removing any rule from `skillrules.Check` fails its case. The `broken_skill` CLI subtest fails if `CheckSkills` is taken out of `all` (`main.go:56`). Moving the `noodle:*` names back into the Workflow section makes `make spec-check` fail.

Scope: commits `a154bce` and `fbc9076` touch only `internal/skillrules/`, `tools/speccheck/`, `skills/noodle-diagram/SKILL.md`, `tasks.md` and `PROGRESS.md`. These are the packages the plan names. No lint rule was relaxed (C04). The `features.json` entry for PRJ-03 is unchanged and still says `passes: false` (C12).

## Findings

These findings break no criterion.

- `internal/skillrules/skillrules.go:150-154`: the Claude Code-only check matches substrings. So `/plugin` would also flag host-neutral text such as `.claude-plugin/plugin.json` or a URL path. Matching on word boundaries, or treating `/plugin` as a command only at the start of a token, would remove the false positive.
- `internal/skillrules/skillrules.go:83`: a missing `name` is reported under `name-format`, which hides the real cause. The fix is a dedicated message or rule, or saying so in FR-004's rule list.
- `internal/skillrules`: the package has no tests of its own (`[no test files]`). Its tests live in `tools/speccheck`. `cmd/noodle` also uses it (PRJ-04 #7), so some unit tests in the package itself would keep it covered if speccheck changes.
- `skills/noodle-diagram/SKILL.md:13`: the reflow in `fbc9076` left one line of about 115 characters, while the rest of the file wraps at about 90. This is cosmetic.
- `skills/noodle-diagram/SKILL.md:40,48-50`: the `-o` comment and the `#71` `-s` scale note came in through other features' commits, not PRJ-03. They are correct, but they sit outside this feature's tasks.
- Contract follow-up for PRJ-03 #1, the table-driven `claude_name_outside_heading` over the six literals: this is done (`skills_test.go:73`).
