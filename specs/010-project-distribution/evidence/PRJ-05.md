# Evidence · PRJ-05

- Feature: PRJ-05
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: db81c22
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Files exist, the code of conduct is a real one | PASS | `ls` lists all five files with rc=0. `grep -c -i 'Contributor Covenant' CODE_OF_CONDUCT.md` returns `2`. The file is the Contributor Covenant 2.1 text: line 73 reads "adapted from the [Contributor Covenant][homepage], version 2.1". |
| 2 | CONTRIBUTING follows the process | PASS | All five patterns match: `22:` docs/PROCESS.md, `30:` make check, `33:` make evaluate, `51:` Conventional Commits, `60:` before/after |
| 3 | Template names | PASS | `idea.yml:1:name: Idea` and `bug.yml:1:name: Bug` |
| 4 | Idea template fields | PASS | The labels at lines 6, 12, 18, 22, 26, 32 and 36 are Context, Proposal, Scope, Out of scope, Acceptance, Related and Backlog id. There are exactly seven, in the required order. |
| 5 | Bug template fields | PASS | The labels at lines 6, 12, 18 and 24 are What happens, Reproduce, Expected and Version. There are exactly four, in the required order. |
| 6 | SECURITY gives a private channel and says why | FAIL | Both greps match: `5:Report it privately at <https://github.com/TheGostsniperfr/Noodle/security/advisories/new>` and `9:## Why an architecture map is sensitive`. But `gh api repos/TheGostsniperfr/Noodle/private-vulnerability-reporting` returns `{"enabled":false}`, and the repo is public (`"visibility":"public"`). Private vulnerability reporting is off, so a reporter cannot file anything through that link. The criterion asks for a private channel and the grep only proves a URL string. The channel in the file does not work. |
| 7 | README links them | PASS | `README.md:125` `[Contributing](CONTRIBUTING.md)`, `:126` `[Security policy](SECURITY.md)`, `:127` `[Code of conduct](CODE_OF_CONDUCT.md)` |
| 8 | No DCO required | PASS | `grep -n -i 'signed-off-by' CONTRIBUTING.md` has no output, rc=1 |

## Verify step

    $ ls CONTRIBUTING.md SECURITY.md CODE_OF_CONDUCT.md .github/ISSUE_TEMPLATE/idea.yml .github/ISSUE_TEMPLATE/bug.yml
    CODE_OF_CONDUCT.md  CONTRIBUTING.md  .github/ISSUE_TEMPLATE/bug.yml  .github/ISSUE_TEMPLATE/idea.yml  SECURITY.md
    $ grep -n 'label:' .github/ISSUE_TEMPLATE/idea.yml
    6: Context / 12: Proposal / 18: Scope / 22: Out of scope / 26: Acceptance / 32: Related / 36: Backlog id
    $ grep -n 'label:' .github/ISSUE_TEMPLATE/bug.yml
    6: What happens / 12: Reproduce / 18: Expected / 24: Version
    $ grep -n -F -e '](CONTRIBUTING.md)' -e '](SECURITY.md)' -e '](CODE_OF_CONDUCT.md)' README.md
    125, 126, 127 (one link each)
    $ gh api repos/TheGostsniperfr/Noodle/private-vulnerability-reporting
    {"enabled":false}
    $ make spec-check
    == spec-check ... rc=0

The verify text is "the five files exist with the fields of FR-008 and the README links the first three". The files, the fields and the links are all there. The verify step does not cover FR-008's "a private reporting channel" in substance, because the channel does not work.

## Findings

- `SECURITY.md:5`: the only reporting channel is `security/advisories/new`, and private vulnerability reporting is disabled on this public repository (`{"enabled":false}`). A reporter cannot report privately, which is what FR-008 requires. Two fixes are possible:
  - Brian turns the setting on: Settings → Code security → Private vulnerability reporting, or `gh api -X PUT repos/TheGostsniperfr/Noodle/private-vulnerability-reporting`. Then rerun the check and record `{"enabled":true}` in the evidence.
  - Or `SECURITY.md` names a channel that works, such as a contact address.

  Contract criterion 6 should also gain this API check, so that a grep for a URL is not enough to pass it.
- `CODE_OF_CONDUCT.md:39`: conduct reports go to the same disabled advisory form, so the enforcement contact does not work either. The fix above covers it too. Using a security advisory form for conduct reports is also unusual. A contact address would separate the two kinds of report.
- Minor, breaks no criterion: there is no `.github/ISSUE_TEMPLATE/config.yml`. The chooser offers Idea and Bug plus "Open a blank issue". The FR-008 scenario still holds. Add `blank_issues_enabled: false` if the two forms should be the only way in.
- Scope: commit 5b5be54 touches only the five files, `README.md`, `docs/PROGRESS.md` and `tasks.md`. I found no out-of-scope change. The `docs/features.json` entry for PRJ-05 is unchanged and still has `passes: false`. This is a docs-only feature, so it has no Go tests, and nothing guards these files against later removal apart from this contract's greps.
