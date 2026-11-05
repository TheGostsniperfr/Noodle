# Contract review · Spec 010

- Review: **CHANGES**
- Evaluator: evaluator (claude-opus-5-5)
- Date: 2026-11-04

Round 3 of 3. Round 2's 19 points are applied, and the fallback table now names literal files and commands. Under PROCESS.md, two kinds of point still get applied after this round: one that lets the evaluator write, and one that makes a criterion pass or fail whatever the code does. One point below is of the second kind, PRJ-04 #2b, and it must be applied before implementation starts. The other rows are follow-ups: record them in `contract.md` and do not hold implementation for them.

What I checked against the repository:
- **Icon count.** `assets.Icons` embeds `icons/*.svg`. `ls assets/icons/*.svg | wc -l` gives 45, so `ATTRIBUTION.md` is not counted.
- **SKILL.md headings.** `grep -n '^#' skills/noodle-diagram/SKILL.md` matches only real headings, with no code-fence comments. Today the only Claude Code-only name in it is `noodle:fact-finder`, on line 22.
- **Test names.** `TestRun_ExitsWithFindingsThroughTheCLI` is table-driven, so a case named `broken skill` matches `/broken_skill`.
- **Phase id.** P6a-05 sits in phase `6a`, which is what the jq in PRJ-06 #4 expects.
- **Allowlist.** `make release-snapshot` is not in `EVALUATOR` yet. T13 and plan.md:88 add it as an exact entry, so if it is missing at grading time, the code is at fault and not the contract.
- **License before triage.** T07 waits until T01 is on `main`, so PRJ-02 #4 can pass.

## Changes

| Feature | Criterion | Problem | Proposed criterion |
|---|---|---|---|
| PRJ-04 | 2b (**must apply**: result does not depend on the code) | `go test -run` splits its pattern on every `/` that sits outside parentheses or brackets, and `\|` does not stop the split. So `Setup_OpenCode/instructions\|Setup_Codex/instructions` becomes three levels: `Setup_OpenCode`, `instructions\|Setup_Codex`, `instructions`. Level 1 never matches `TestSetup_Codex`, so `Setup_Codex/instructions` never runs and its `--- PASS` cannot appear. On the OpenCode side, any per-agent subtest under `instructions` is filtered out by the third level, so the parent can pass without checking anything. | `go test ./cmd/noodle/ -run 'Setup_(OpenCode\|Codex)/instructions' -v`. Threshold: `--- PASS` for `TestSetup_OpenCode/instructions` and `TestSetup_Codex/instructions`. If they have per-agent subtests, add `--- PASS` for `instructions/diagram-reviewer`, `instructions/fact-finder` and `instructions/icon-curator` under each. |
| PRJ-02 | 2 (follow-up) | "Each piece found once" covers the whole file. A draft reply that mentions `Mermaid export` or `non-destructive` to the author would push the count to two and fail a correct triage. | Run the piece grep only on disposition rows: `grep -E '^[\|] .* [\|] (#[0-9]+\|declined: .+) [\|]$' fork-triage.md \| grep -n -F -e '.opencode/agents/*.md' … -e 'roadmap change'`. Threshold: each piece found once, with line numbers in the quoted order. |
| PRJ-07 | 7 (follow-up) | FR-014 holds only "WHILE neither a `noodle` on PATH, Go nor Nix" is available. The subtests cover `go_on_path` and `noodle_on_path` but not Nix. A wrapper that downloads even when `nix` is present would pass. The two download subtests also have no names. | Add a subtest `nix_on_path`: with a stub `nix`, no download happens. Name the other two `download_verified` and `download_altered`, each printing `--- PASS`. |
| PRJ-08 | 2 (follow-up) | FR-015's scenario is that `make spec-check` validates the three skills. If `process_skill_broken` calls the validator on a path directly, nothing proves that `all` walks `.agents/skills/`. In that case `make spec-check` exits 0 because nothing was read. | `process_skill_broken` runs `run([]string{"-root", tmp, "all"}, …)` on a fixture holding `<tmp>/.agents/skills/broken/SKILL.md`. It asserts `errFindings` and one stderr line naming `.agents/skills/broken/SKILL.md`. |
| PRJ-06 | 1 (follow-up) | FR-009 asks for an explanation page covering model, view, layout, fragment and connection semantics. Only the page names are checked, so an empty `explanation.md` passes. | `grep -n -i -e '^## .*model' -e '^## .*view' -e '^## .*layout' -e '^## .*fragment' -e '^## .*connection' docs/guide/explanation.md`: each found. |
| PRJ-03 | 1, 4 (follow-up) | The literal list in #4 (`noodle:*`, `the Agent tool`, `Task tool`, `/plugin`) is not tied to the list the speccheck rule `claude-only-name` uses. The rule could check fewer names than the grep does. | In #1, the subtest `claude_name_outside_heading` is table-driven over the same six literals as #4. Each case gives one finding with id `claude-only-name`. |
| all | Clarifications 1 to 5 (follow-up, precondition) | All five clarifications in `spec.md` are still `pending`. This round needs no contract change for it, since the fallback table is now literal. PROCESS.md steps 2 and 3 still apply, though: T01 cannot start before clarification 1 is answered and recorded in an ADR. | Record Brian's answers and the license ADR before T01. Then apply the matching fallback rows to the criteria and delete the table. |
