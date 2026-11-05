# Evidence · PRJ-05

- Feature: PRJ-05
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: a51d77b
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Files exist, the code of conduct is a real one | PASS | `ls` on the five paths lists all five, rc=0. `grep -c -i 'Contributor Covenant' CODE_OF_CONDUCT.md` returns `2`. |
| 2 | CONTRIBUTING follows the process | PASS | All five patterns are found: `22:` docs/PROCESS.md, `30:` make check, `33:` make evaluate, `51:` Conventional Commits, `60:` before/after. |
| 3 | Template names | PASS | `idea.yml:1:name: Idea` and `bug.yml:1:name: Bug`. |
| 4 | Idea template fields | PASS | The labels at lines 6, 12, 18, 22, 26, 32 and 36 are Context, Proposal, Scope, Out of scope, Acceptance, Related and Backlog id. There are exactly seven, in the required order. |
| 5 | Bug template fields | PASS | The labels at lines 6, 12, 18 and 24 are What happens, Reproduce, Expected and Version. There are exactly four, in the required order. |
| 6 | SECURITY gives a private channel and says why | PASS | `5:Report it privately at <https://github.com/TheGostsniperfr/Noodle/security/advisories/new>` and `9:## Why an architecture map is sensitive`. In the previous round this channel did not work. Now `gh api repos/TheGostsniperfr/Noodle/private-vulnerability-reporting` returns `{"enabled":true}`, so the channel works. |
| 7 | README links them | PASS | `README.md:125` links `[Contributing](CONTRIBUTING.md)`, `:126` links `[Security policy](SECURITY.md)` and `:127` links `[Code of conduct](CODE_OF_CONDUCT.md)`. |
| 8 | No DCO required | PASS | `grep -n -i 'signed-off-by' CONTRIBUTING.md` prints nothing, rc=1. |

## Verify step

    $ ls CONTRIBUTING.md SECURITY.md CODE_OF_CONDUCT.md .github/ISSUE_TEMPLATE/idea.yml .github/ISSUE_TEMPLATE/bug.yml
    CODE_OF_CONDUCT.md  CONTRIBUTING.md  .github/ISSUE_TEMPLATE/bug.yml  .github/ISSUE_TEMPLATE/idea.yml  SECURITY.md
    rc=0
    $ grep -n 'label:' .github/ISSUE_TEMPLATE/idea.yml
    6: Context / 12: Proposal / 18: Scope / 22: Out of scope / 26: Acceptance / 32: Related / 36: Backlog id
    $ grep -n 'label:' .github/ISSUE_TEMPLATE/bug.yml
    6: What happens / 12: Reproduce / 18: Expected / 24: Version
    $ grep -n -F -e '](CONTRIBUTING.md)' -e '](SECURITY.md)' -e '](CODE_OF_CONDUCT.md)' README.md
    125, 126, 127 (one link each)
    $ gh api repos/TheGostsniperfr/Noodle/private-vulnerability-reporting
    {"enabled":true}
    $ make spec-check
    == spec-check ... rc=0

## Findings

None of these breaks a contract criterion.

- **Previous blocker resolved.** The NEEDS_WORK in evidence at db81c22 came from the disabled advisory form. That form is now enabled on the repository. The fix is a repository setting, not a commit, so `tasks.md` has no Convergence task for PRJ-05.
- **Contract criterion 6 is still only a grep** (`contract.md`, PRJ-05 #6). The previous evidence proposed adding the `private-vulnerability-reporting` API check. That was not done. If the setting is switched off again, the contract would still pass. Fix: add that check as a follow-up.
- **`CODE_OF_CONDUCT.md:39`:** conduct reports go through the security advisory form, with a "Conduct" title prefix. This works now, but it is unusual and mixes two kinds of report. A contact address would keep them apart.
- **No `.github/ISSUE_TEMPLATE/config.yml`.** The chooser will also offer "Open a blank issue". The FR-008 scenario still holds. Add `blank_issues_enabled: false` if Idea and Bug should be the only way to open an issue.
- **The FR-008 scenario can only be checked after merge.** `gh api .../contents/.github/ISSUE_TEMPLATE` returns 404 on the default branch, so the live chooser cannot be checked from this branch. I read both templates by hand: they are well-formed GitHub issue forms (top-level `name`, `description` and `body`, `textarea`/`input` types, `attributes.label`). After the merge, open a new issue once to confirm.
- **Scope and C12 are clean.** The PRJ-05 entry in `docs/features.json` matches its original wording and still has `passes: false`. Since the previous evaluation, nothing has changed in the five files, `README.md`, or the PRJ-05 contract section. This is a docs-only feature with no Go test. Nothing guards these files against removal except this contract's greps.
