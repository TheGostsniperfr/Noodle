# Contract 010 · Project and distribution

What "done" means for each feature, agreed before implementation (ADR-0024).
One failed criterion fails the feature. Every "Verified by" runs inside the evaluator's
allowlist (`EVALUATOR` in the Makefile); a command outside it is refused, and fails.
In a table cell, `\|` is the Markdown escape of a pipe: run every command with a plain
`|`. A literal pipe inside a regex is written `[|]`. "exit 0" with subtests means every
subtest passes, each one asserting the failure it provokes.

- Agreed: pending `make review-contract SPEC=010` (at most three rounds, PROCESS.md).
- Round 1 (2026-11-04): CHANGES, 31 points, all applied. Clarifications are still open:
  the criteria below assume the recommended answer of each, and the last section lists
  what changes if Brian answers otherwise.
- The evaluator cannot run `gh`; the triage file copies #94's rows itself.

## PRJ-01 · License and attribution

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | One SPDX id in plugin and skill | `jq -r .license .claude-plugin/plugin.json`; `grep -n '^license:' skills/noodle-diagram/SKILL.md` | both print `Apache-2.0` |
| 2 | The license text is Apache-2.0 | `grep -c 'Apache License' LICENSE`; `grep -c 'Version 2.0, January 2004' LICENSE` | at least `1`; exactly `1` |
| 3 | The flake declares it | `grep -n 'license' flake.nix` | one match with `lib.licenses.asl20` |
| 4 | NOTICE names product, author and URL | `cat NOTICE` | contains `Noodle`, `Brian Perret` and `https://github.com/TheGostsniperfr/Noodle` |
| 5 | The attribution test reads the embedded icons and fails per missing piece | `go test ./assets/ -run Attribution -v` | exit 0; subtests `missing_row` (the `cilium.svg` row removed from an in-memory copy: exactly one error containing `cilium.svg`), `row_without_url` and `row_without_license` (one error each, naming the file); the file list comes from `assets.Icons` |
| 6 | One row per icon file | `grep -c '^[|] `' assets/icons/ATTRIBUTION.md`; `ls assets/icons/*.svg \| wc -l` | equal counts (45 today, `.dark` and `.light` variants counted apart) |
| 7 | Trademark notice | `grep -ci 'trademark' assets/icons/ATTRIBUTION.md` | at least `1` |

## PRJ-02 · Fork triage

The disposition count is checked with this command, outside the table:

```bash
grep -cE '^[|] .* [|] (#[0-9]+|declined: .+) [|]$' specs/010-project-distribution/fork-triage.md
```

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Commits and build result recorded | `grep -n -e 6d61ce0 -e 2541717 specs/010-project-distribution/fork-triage.md`; `grep -n -e '^## Build and test' -e 'network none' specs/010-project-distribution/fork-triage.md` | both SHAs found; the section exists, names `--network none`, and gives the exit code of `go test` |
| 2 | One disposition per row of #94, rows copied verbatim | the command above; `grep -c '^[|] `' specs/010-project-distribution/fork-triage.md` | `7`; the table's first column holds #94's seven pieces in #94's order |
| 3 | Draft reply present | `grep -n -e '^## Reply to the author' -e 'Apache-2.0' -e '#94' specs/010-project-distribution/fork-triage.md` | each found, the last two inside the reply section |
| 4 | The license is on main before triage lands | `git log origin/main --diff-filter=A --format=%h -- LICENSE` | non-empty |
| 5 | No fork commit predates the license | `git log origin/main -i --author=girold --format=%h`; `git log <sha from #4>..origin/main -i --author=girold --format=%h` | the two lists are equal as sets |

## PRJ-03 · Host-neutral skills, checked

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Each rule has a case with its exact message | `go test ./tools/speccheck/ -run Skills -v` | exit 0; subtests for name not equal to folder, name with uppercase, name over 64 characters, empty description, description over 1024 characters, unknown frontmatter key, a Claude Code-only name outside "In Claude Code"; each asserts exactly one finding naming the file and the rule; a valid skill and a name under "In Claude Code" yield 0 findings |
| 2 | `all` fails on a broken skill | `go test ./tools/speccheck/ -run 'TestRun_Exits.*/broken_skill' -v` | `--- PASS` for subtest `broken_skill`, which asserts `errFindings` from `all` |
| 3 | The repository's skills pass | `make spec-check` | exit 0 |
| 4 | Claude Code-only names sit under their heading | `grep -n -e 'noodle:fact-finder' -e 'noodle:diagram-reviewer' -e 'noodle:icon-curator' -e 'the Agent tool' -e 'Task tool' -e '/plugin' skills/noodle-diagram/SKILL.md`; `grep -n '^#' skills/noodle-diagram/SKILL.md` | every match's line number lies after a heading starting `In Claude Code` and before the next heading of the same or higher level |

## PRJ-04 · `noodle setup` for OpenCode and Codex

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | OpenCode project setup | `go test ./cmd/noodle/ -run Setup_OpenCode -v` | exit 0; files equal `cmd/noodle/testdata/setup/opencode/`; stdout is exactly five `wrote` lines: `.agents/skills/noodle-diagram/SKILL.md`, `.agents/skills/noodle-diagram/reference.md`, `.opencode/agents/diagram-reviewer.md`, `.opencode/agents/fact-finder.md`, `.opencode/agents/icon-curator.md` |
| 2 | Codex project setup | `go test ./cmd/noodle/ -run Setup_Codex -v`; `grep -n '^<instructions key from plan.md>' cmd/noodle/testdata/setup/codex/.codex/agents/fact-finder.toml` | exit 0; files equal `cmd/noodle/testdata/setup/codex/`; stdout is exactly the five files written; the key that T08 wrote under "Host facts" in `plan.md` is found |
| 3 | `-user` scope | `go test ./cmd/noodle/ -run Setup_User -v` | exit 0; with `HOME` set to a temp dir, files appear under `~/.agents/skills/`, `~/.config/opencode/agents/` and `~/.codex/agents/` |
| 4 | Idempotent per host | `go test ./cmd/noodle/ -run Setup_Idempotent -v` | exit 0; subtests `opencode` and `codex`: the second run prints no `wrote` line and every file's bytes and modification time are unchanged |
| 5 | Foreign files are kept | `go test ./cmd/noodle/ -run Setup_KeepsForeign -v` | exit 0; a file without the marker is unchanged, named on stderr, and the exit code is 1 |
| 6 | Read-only agents deny edits | `grep -n -e '^mode: subagent' -e 'edit: deny' cmd/noodle/testdata/setup/opencode/.opencode/agents/fact-finder.md`; same grep on `icon-curator.md` | both found for `fact-finder.md`; `mode: subagent` found and `edit: deny` absent for `icon-curator.md` |
| 7 | The written skill still passes the skill rules | `go test ./cmd/noodle/ -run Setup_OpenCode/skill_rules -v` | `--- PASS`: the speccheck skill validator returns 0 findings on the written `SKILL.md` |
| 8 | Formats are sourced | `grep -n -e 'opencode.ai/docs' -e 'developers.openai.com/codex' specs/010-project-distribution/plan.md` | both found, and the Codex instructions key is written under "Host facts" |
| 9 | README installs per host | `grep -n -e '^## ' -e '^### ' README.md` | `### Claude Code`, `### OpenCode`, `### Codex`, `### Nix` have line numbers between `## Install` and the next `## ` |

## PRJ-05 · Contributor files and issue templates

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Files exist | `ls CONTRIBUTING.md SECURITY.md CODE_OF_CONDUCT.md .github/ISSUE_TEMPLATE/idea.yml .github/ISSUE_TEMPLATE/bug.yml` | all five listed |
| 2 | CONTRIBUTING follows the process | `grep -n -e 'docs/PROCESS.md' -e 'make check' -e 'make evaluate' -e 'Conventional' -e 'before/after' CONTRIBUTING.md` | each pattern found |
| 3 | Template names | `grep -n '^name:' .github/ISSUE_TEMPLATE/idea.yml .github/ISSUE_TEMPLATE/bug.yml` | `name: Idea` and `name: Bug` |
| 4 | Idea template fields | `grep -n 'label:' .github/ISSUE_TEMPLATE/idea.yml` | exactly Context, Proposal, Scope, Out of scope, Acceptance, Related, Backlog id, in that order |
| 5 | Bug template fields | `grep -n 'label:' .github/ISSUE_TEMPLATE/bug.yml` | exactly What happens, Reproduce, Expected, Version, in that order |
| 6 | SECURITY gives a private channel and says why | `grep -n -e 'security/advisories/new' -e '@' SECURITY.md`; `grep -n -i 'architecture' SECURITY.md` | at least one match each |
| 7 | README links them | `grep -n -e CONTRIBUTING.md README.md`; same for `SECURITY.md` and `CODE_OF_CONDUCT.md` | each found at least once |
| 8 | No DCO required | `grep -n -i 'signed-off-by' CONTRIBUTING.md` | no match (clarification 5, recommended answer) |

## PRJ-06 · User guide, positioning, discovery reference

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Guide pages exist | `ls docs/guide/` | exactly `cli.md`, `explanation.md`, `howto-diff.md`, `howto-discover.md`, `howto-hosts.md`, `howto-slides.md`, `tutorial.md` |
| 2 | The tutorial's system lints | `ls examples/tutorial`; `grep -n 'examples/tutorial' Makefile`; `make examples` | the directory exists, the `examples` target renders it, exit 0 |
| 3 | The CLI reference cannot go stale | `go test ./cmd/noodle/ -run CLIReference -v`; `grep -n '^cli-doc:' Makefile`; `grep -n 'make cli-doc' docs/COMMANDS.md` | exit 0 with a subtest `stale` asserting an error containing `docs/guide/cli.md` for an altered usage copy; the target and its row exist |
| 4 | Positioning is honest | `grep -n -i -e 'stays in sync' -e 'stay in sync' README.md`; `jq -r '.phases[] \| select(.id=="6a") \| .features[] \| select(.id=="P6a-05") \| .passes' docs/features.json` | while P6a-05 is `false`, at least one match and every match contains `designed to stay in sync` |
| 5 | VISION names the established terms | `grep -n -e 'architecture as code' -e 'living documentation' docs/VISION.md` | both found |
| 6 | Discovery section covers #88 | `grep -n '^## ' skills/noodle-diagram/reference.md`, then `grep -n -F -e <literal>` for each literal below | each of `#71`, `-namespace`, `kustomize`, `helm`, `matches`, `not discovered`, `#78`, `zone/**`, `#74`, `tech`, `status: planned`, `target`, `fact-finder` has a match between `## Discovery` and the next `## ` (case-insensitive for `kustomize` and `helm`) |

## PRJ-07 · Versions and releases

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | `noodle version` reads the build variable | `go test ./cmd/noodle/ -run Version -v` | exit 0; a build with `-X main.version=v0.1.0` prints `v0.1.0`, a plain build prints `dev` |
| 2 | The flake passes its revision | `grep -n 'main.version' flake.nix` | one match setting it from `self.shortRev or self.dirtyShortRev` |
| 3 | Release workflow on tags only, body from the changelog | `grep -n -e 'tags:' -e "'v\*'" .github/workflows/release.yml`; `grep -n -e 'git-cliff' -e 'body_path' .github/workflows/release.yml`; `grep -n -e 'branches' -e 'pull_request' .github/workflows/release.yml` | the trigger is `v*` tags; git-cliff and `body_path` are used; no branch or pull-request trigger |
| 4 | Four platforms and checksums configured | `grep -n -e linux -e darwin -e amd64 -e arm64 -e checksum .goreleaser.yaml` | each found |
| 5 | A snapshot builds now | `make release-snapshot`; `ls dist/`; `cat dist/*checksums.txt` | exit 0; exactly `noodle_*_linux_amd64.tar.gz`, `noodle_*_linux_arm64.tar.gz`, `noodle_*_darwin_amd64.tar.gz`, `noodle_*_darwin_arm64.tar.gz`; four checksum lines, one per archive |
| 6 | Changelog from commits | `grep -n '^changelog:' Makefile`; `grep -n '^### ' CHANGELOG.md`; `grep -n 'group' cliff.toml`; `head -20 CHANGELOG.md` | the `changelog` target exists; every `###` heading of CHANGELOG.md is a group defined in `cliff.toml`; the top has a version or `Unreleased` heading |
| 7 | Fallback downloads the right archive, verifies, refuses a mismatch | `go test ./cmd/noodle/ -run Wrapper -v` | exit 0; subtests: with no `noodle`, `go` or `nix` on PATH and a stub `curl` recording its URLs, a matching archive runs and the URLs contain the `version` of `.claude-plugin/plugin.json` and `<os>_<arch>` from a stubbed `uname`; an altered archive exits non-zero, prints a checksum error and leaves no binary in the cache; with `go` on PATH no download happens |

## PRJ-08 · Process skills

| # | Criterion | Verified by | Threshold |
|---|---|---|---|
| 1 | Three skills, one source | `ls .agents/skills/`; `ls -l .claude/skills/` | `issue-writer`, `release-notes`, `contrib-check` each with a `SKILL.md` under `.agents/skills/`; each also present in `.claude/skills/` as a symlink to it |
| 2 | They pass the skill rules | `make spec-check`; `go test ./tools/speccheck/ -run Skills -v` | exit 0; subtest `process_skill_broken` puts a fixture under `.agents/skills/` with an over-long description and asserts one finding naming that path |
| 3 | `issue-writer` targets the template | `grep -n -e 'idea.yml' -e 'B-NN' .agents/skills/issue-writer/SKILL.md` | both found |
| 4 | `release-notes` reuses PR images and links migrations | `grep -n -e 'before/after' -e 'migration' .agents/skills/release-notes/SKILL.md` | both found |
| 5 | `contrib-check` runs the CONTRIBUTING checklist | `grep -n -e 'CONTRIBUTING.md' -e 'make check' .agents/skills/contrib-check/SKILL.md` | both found |

## If Brian answers otherwise

| Clarification | Other answer | Criteria that change |
|---|---|---|
| 1 | MIT | PRJ-01 #1 and #3 read `MIT` and `lib.licenses.mit`; #2 becomes `grep -c '^MIT License' LICENSE` = `1`; #4 is dropped (no NOTICE); PRJ-02 #3 reads `MIT` |
| 2 | Code review only, nothing run | PRJ-02 #1: the section states "not run" and why, and lists the files read; no `--network none` |
| 3 | release-please with goreleaser | PRJ-07 #3 and #6 read `release-please` and its config file instead of git-cliff and `cliff.toml` |
| 4 | `.claude/skills/` only | PRJ-08 #1 to #5 read `.claude/skills/` and the symlink check is dropped |
| 5 | DCO required | PRJ-05 #8: `grep -n -i 'signed-off-by' CONTRIBUTING.md` finds a match, and a CI check of sign-offs joins PRJ-05 |
