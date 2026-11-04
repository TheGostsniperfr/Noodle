# Contract 010 · Project and distribution

What "done" means for each feature, agreed before implementation (ADR-0024).
One failed criterion fails the feature. Every "Verified by" runs inside the evaluator's
allowlist (`EVALUATOR` in the Makefile); a command outside it is refused, and fails.
In a table cell, `\|` is the Markdown escape of a pipe: run every command with a plain `|`.

- Agreed: pending `make review-contract SPEC=010` (at most three rounds, PROCESS.md).
- Commands the implementer runs and the evaluator only inspects: `make release-snapshot`
  style builds that write `dist/`, and posting the fork reply. Their results are read
  with `ls`, `cat` and `git` only.

## PRJ-01 · License and attribution

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | One SPDX id everywhere | `head -1 LICENSE`; `jq -r .license .claude-plugin/plugin.json`; `grep -n 'license' flake.nix`; `grep -n '^license:' skills/noodle-diagram/SKILL.md` | the four name the same SPDX id (`Apache-2.0` or `MIT`, per clarification 1) |
| 2 | NOTICE names product, author and URL when Apache-2.0 | `cat NOTICE` | contains `Noodle`, `Brian Perret` and `https://github.com/TheGostsniperfr/Noodle`; criterion not applicable when the license is MIT, stated as such |
| 3 | Every icon has a row | `go test ./assets/ -run Attribution -v` | exit 0; a subtest that removes one row from a copy of the page fails naming that file |
| 4 | Rows carry a source and a license | `grep -c '^| `' assets/icons/ATTRIBUTION.md`; `ls assets/icons/*.svg \| wc -l` | the two counts are equal; `grep -c 'https://' assets/icons/ATTRIBUTION.md` is at least that count |
| 5 | Trademark notice | `grep -ci 'trademark' assets/icons/ATTRIBUTION.md` | at least `1` |

## PRJ-02 · Fork triage

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Commits and result recorded | `grep -n -e 6d61ce0 -e 2541717 specs/010-project-distribution/fork-triage.md`; `grep -n '^## Build and test' specs/010-project-distribution/fork-triage.md` | both SHAs found; the section exists and states how the code was run (clarification 2) and its exit codes |
| 2 | One disposition per row of #94 | `grep -cE '^\| .* \| (#[0-9]+\|declined: .+) \|$' specs/010-project-distribution/fork-triage.md` | `7` |
| 3 | Draft reply present | `grep -n '^## Reply to the author' specs/010-project-distribution/fork-triage.md` | found, and the section names the license and links #94 |
| 4 | Nothing of the fork reached main before the license | `git log main --format='%h %an' --author=AlexandreGirold`; `git log main --format=%h -1 -- LICENSE` | the first list is empty, or every commit in it is a descendant of the `LICENSE` commit (`git log main --format=%h <license sha>..main --author=AlexandreGirold` lists them all) |

## PRJ-03 · Host-neutral skills, checked

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Each rule has a failing case with its exact message | `go test ./tools/speccheck/ -run Skills -v` | exit 0; one subtest each for name not equal to folder, name with uppercase, name over 64 characters, empty description, description over 1024 characters, unknown frontmatter key, a Claude Code-only name outside "In Claude Code"; each asserts exactly one finding naming the file and the rule; a valid skill and a name under "In Claude Code" yield 0 findings |
| 2 | `all` runs it | `go test ./tools/speccheck/ -run TestRun_Exits -v` | a case with a broken skill returns `errFindings` from `all` |
| 3 | The repository's skills pass | `make spec-check` | exit 0 |
| 4 | The skill body is neutral | `grep -n -e 'noodle:fact-finder' -e 'the Agent tool' skills/noodle-diagram/SKILL.md` | every match is below a heading that starts with `In Claude Code` (read the file) |

## PRJ-04 · `noodle setup` for OpenCode and Codex

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | OpenCode project setup writes the skill and one agent file per agent | `go test ./cmd/noodle/ -run Setup_OpenCode -v` | exit 0; files equal their golden copies; printed paths are exactly the files written |
| 2 | Codex project setup | `go test ./cmd/noodle/ -run Setup_Codex -v` | exit 0; `.codex/agents/<name>.toml` per agent equal their golden copies |
| 3 | `-user` writes under the documented home paths | `go test ./cmd/noodle/ -run Setup_User -v` | exit 0; with `HOME` set to a temp dir, files appear under `~/.agents/skills/`, `~/.config/opencode/agents/` and `~/.codex/agents/` |
| 4 | Idempotent | `go test ./cmd/noodle/ -run Setup_Idempotent -v` | exit 0; a second run writes nothing and every file is byte-identical |
| 5 | Foreign files are kept | `go test ./cmd/noodle/ -run Setup_KeepsForeign -v` | exit 0; a file without the marker is unchanged, named on stderr, and the exit code is 1 |
| 6 | Read-only agents deny edits | `go test ./cmd/noodle/ -run Setup_OpenCode -v`, the golden of `fact-finder.md` | its frontmatter has `mode: subagent` and edits denied |
| 7 | Formats are sourced | `grep -n -e 'opencode.ai/docs' -e 'developers.openai.com/codex' specs/010-project-distribution/plan.md` | both found, and the Codex TOML keys are written in the plan (T08) |
| 8 | README has one install section per host | `grep -n '^### ' README.md` | headings for Claude Code, OpenCode, Codex and Nix under "Install" |

## PRJ-05 · Contributor files and issue templates

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Files exist | `ls CONTRIBUTING.md SECURITY.md CODE_OF_CONDUCT.md .github/ISSUE_TEMPLATE/idea.yml .github/ISSUE_TEMPLATE/bug.yml` | all listed |
| 2 | CONTRIBUTING follows the process | `grep -n -e 'docs/PROCESS.md' -e 'make check' -e 'make evaluate' -e 'Conventional' CONTRIBUTING.md` | each found |
| 3 | Idea template fields | `grep -n 'label:' .github/ISSUE_TEMPLATE/idea.yml` | exactly Context, Proposal, Scope, Out of scope, Acceptance, Related, Backlog id |
| 4 | Bug template fields | `grep -n 'label:' .github/ISSUE_TEMPLATE/bug.yml` | exactly What happens, Reproduce, Expected, Version |
| 5 | SECURITY gives a private channel | `grep -ni -e 'security advisor' -e 'private' SECURITY.md` | at least one match naming GitHub private vulnerability reporting or an address |
| 6 | README links them | `grep -n -e CONTRIBUTING.md -e SECURITY.md -e CODE_OF_CONDUCT.md README.md` | three matches |

## PRJ-06 · User guide, positioning, discovery reference

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Guide pages exist | `ls docs/guide/` | `tutorial.md`, `explanation.md`, `cli.md` and at least four how-to pages (discover, diff, slides, hosts) |
| 2 | CLI reference cannot go stale | `go test ./cmd/noodle/ -run CLIReference -v` | exit 0; a subtest that alters a copy of the usage text fails naming `docs/guide/cli.md` |
| 3 | Positioning is honest | `grep -n 'designed to stay in sync' README.md`; `jq -r '.phases[] \| select(.id=="6a") \| .features[] \| select(.id=="P6a-05") \| .passes' docs/features.json` | the phrase is present while P6a-05 is `false` |
| 4 | VISION names the established terms | `grep -n -e 'architecture as code' -e 'living documentation' docs/VISION.md` | both found |
| 5 | Discovery section covers #88 | `grep -n '^## Discovery' skills/noodle-diagram/reference.md`, then read that section | it covers each of the eleven items listed in FR-011 |

## PRJ-07 · Versions and releases

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | `noodle version` reads the build variable | `go test ./cmd/noodle/ -run Version -v` | exit 0; a build with `-X main.version=v0.1.0` prints `v0.1.0`, a plain build prints `dev` |
| 2 | Release workflow on tags | `grep -n -e "tags:" -e "v\*" .github/workflows/release.yml` | the workflow triggers on `v*` tags only |
| 3 | Four platforms and checksums | `grep -n -e linux -e darwin -e amd64 -e arm64 -e checksum .goreleaser.yaml` (or the file clarification 3 names) | each found |
| 4 | A snapshot was built | `ls dist/` after the implementer's snapshot build | four archives and one checksums file |
| 5 | Changelog from commits | `head -20 CHANGELOG.md`; `cat cliff.toml` (or the config clarification 3 names) | sections grouped by conventional type, generated by the chosen tool |
| 6 | Fallback downloads, verifies, refuses a mismatch | `go test ./cmd/noodle/ -run Wrapper -v` | exit 0; with no `noodle`, `go` or `nix` on PATH and a stub `curl`, a matching archive runs; an altered one exits non-zero, prints a checksum error, and leaves no binary in the cache |

## PRJ-08 · Process skills

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Three skills where clarification 4 says | `ls` of that directory | `issue-writer`, `release-notes`, `contrib-check`, each with a `SKILL.md` |
| 2 | They pass the skill rules | `make spec-check` | exit 0, and `go test ./tools/speccheck/ -run Skills` covers process skills (a broken one fails) |
| 3 | `issue-writer` targets the template | `grep -n -e 'idea.yml' -e 'B-NN' <issue-writer SKILL.md>` | both found |
| 4 | `release-notes` reuses PR images and links migrations | `grep -n -e 'before' -e 'migration' <release-notes SKILL.md>` | both found |
| 5 | `contrib-check` runs the CONTRIBUTING checklist | `grep -n -e 'CONTRIBUTING.md' -e 'make check' <contrib-check SKILL.md>` | both found |
