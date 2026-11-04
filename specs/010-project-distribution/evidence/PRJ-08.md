# Evidence · PRJ-08

- Feature: PRJ-08
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: db81c22
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Three skills, one source | PASS | `ls .agents/skills/` lists `contrib-check`, `issue-writer` and `release-notes`, each with a `SKILL.md`. `ls -l .claude/skills/` shows `contrib-check -> ../../.agents/skills/contrib-check`, `issue-writer -> ../../.agents/skills/issue-writer` and `release-notes -> ../../.agents/skills/release-notes`. `git ls-files -s` shows mode `120000` for all three, so they are committed as symlinks. |
| 2 | They pass the skill rules | PASS | `make spec-check` exits 0. `go test ./tools/speccheck/ -run Skills -v` exits 0 with `--- PASS: TestSkills_FailsAllForABrokenProcessSkill/process_skill_broken`. That test (`tools/speccheck/skills_test.go:113-125`) writes `.agents/skills/broken/SKILL.md` with a 1025-character description into a fixture root, then runs `run([]string{"-root", repo.Root, "all"}, …)`. It asserts `errFindings`, exactly one stderr line, and that the line contains the path. |
| 3 | `issue-writer` targets the template | PASS | `grep` matches `idea.yml` at lines 3, 9 and 24 and `B-NN` at lines 3, 16 and 32. |
| 4 | `release-notes` reuses PR images and links migrations | PASS | `grep` matches `before/after` at lines 3 and 19 and `migration` at lines 29, 30 and 32. |
| 5 | `contrib-check` runs the CONTRIBUTING checklist | PASS | `grep` matches `CONTRIBUTING.md` at lines 3, 7 and 9 and `make check` at lines 3 and 17. |

## Verify step

    $ make spec-check; echo "exit=$?"
    == spec-check
    go run ./tools/speccheck all
    exit=0

    $ go test ./tools/speccheck/ -run Skills -v
    --- PASS: TestSkills_ChecksEachRule (0.00s)
        --- PASS: TestSkills_ChecksEachRule/valid (0.00s)
        ...
    --- PASS: TestSkills_FailsAllForABrokenProcessSkill (0.00s)
        --- PASS: TestSkills_FailsAllForABrokenProcessSkill/process_skill_broken (0.00s)
    --- PASS: TestSkills_KeepsClaudeCodeNamesUnderTheirHeading (0.00s)
    PASS
    ok  	github.com/TheGostsniperfr/Noodle/tools/speccheck

    $ ls .agents/skills/*
    .agents/skills/contrib-check:  SKILL.md
    .agents/skills/issue-writer:   SKILL.md
    .agents/skills/release-notes:  SKILL.md

    $ make features BASE=origin/main
    go run ./tools/speccheck features -base origin/main   (no findings)

## Findings

- **The tests would catch the feature being reverted.** `tools/speccheck/skills.go:15` adds `.agents/skills/*/SKILL.md` to `skillGlobs`. If that glob is removed, `process_skill_broken` gets no finding and `assert.ErrorIs(err, errFindings)` fails. The `valid` subtest at `skills_test.go:52` also puts a skill under `.agents/skills/` and expects 0 findings, so it covers the passing case.
- **Scope.** Commit `9adb43a` (T15) adds only the three `SKILL.md` files, the three symlinks, the PROGRESS line and the T15 checkbox. The speccheck glob and its test came in earlier, in `a154bce` (PRJ-03), which is within the plan. The `docs/features.json` entry for PRJ-08 was added unchanged and is still `passes: false`. The storage location matches clarification 4: `.agents/skills/<name>/` with symlinks in `.claude/skills/`.
- **Minor, breaks no criterion.** `.agents/skills/issue-writer/SKILL.md:20` finds the highest `B-NN` with `gh issue list --limit 300`. Once the repository has more than 300 issues, this can miss the real maximum and suggest an id that is already taken. Fix: page through all issues, or rely on `docs/features.json` and the board.
- **Minor, outside FR-015.** `.claude/skills/wrap-up/SKILL.md` is a regular directory, not under `.agents/skills/`, so the skill rules do not check it. The spec does not require it, but it is the only skill in the repository that `make spec-check` does not validate.
