# Plan 010 · Project and distribution

Spec: [`spec.md`](spec.md). Decisions: ADR-0024, ADR-0025; clarifications 1 to 5 of the spec.

## Approach

Six independent strands, cut along file lines so they can run in parallel:

1. **Legal** (FR-001, FR-002): text files plus one Go test in `assets/` that compares
   the embedded icon set with `ATTRIBUTION.md`. No network: sources come from the
   icon-curator's known origins (CNCF artwork, Kubernetes community icons,
   simple-icons) and are written by hand, once.
2. **Fork triage** (FR-003): a Markdown record. The build follows clarification 2; the
   reply is a draft, posted by Brian (outward-facing). Lands after strand 1 is on `main`.
3. **Hosts** (FR-004 to FR-007): the skill and agents become embeddable content; a
   `setup` subcommand writes them per host; `speccheck skills` validates skills.
4. **Docs** (FR-008 to FR-011): Markdown, plus one test that keeps `cli.md` equal to the
   usage text.
5. **Releases** (FR-012 to FR-014): a `version` variable set by ldflags, a release
   workflow, a changelog config, and a download fallback in `bin/noodle`.
6. **Process skills** (FR-015): three `SKILL.md` files, validated by strand 3's check.

## Host facts (checked 2026-11-04)

| Host | Skills | Subagents | Source |
|---|---|---|---|
| OpenCode | `.opencode/skills/`, and also `.claude/skills/` and `.agents/skills/` (project); `~/.config/opencode/skills/`, `~/.claude/skills/`, `~/.agents/skills/` (user); folder name equals `name`, `description` 1 to 1024 characters | `.opencode/agents/<name>.md` or `~/.config/opencode/agents/<name>.md`, frontmatter `mode: subagent`, `permission` | [opencode.ai/docs/skills](https://opencode.ai/docs/skills), [opencode.ai/docs/agents](https://opencode.ai/docs/agents/) |
| Codex | `.agents/skills/` at the repository root, `~/.agents/skills/` | one TOML file per agent in `.codex/agents/` or `~/.codex/agents/` | [developers.openai.com/codex/skills](https://developers.openai.com/codex/skills), [developers.openai.com/codex/subagents](https://developers.openai.com/codex/subagents) |

Unverified, to confirm at implementation and record in `fork-triage.md` or PROGRESS:
- OpenCode's v2 docs (v2.opencode.ai) use a different permission syntax (`permissions`
  as an ordered array); the generator targets the current `/docs/agents` format and
  states the version it targets in the marker line.
- The exact TOML keys of a Codex agent file (instructions field name) were read from a
  search summary, not the page itself. T08 reads the page and writes them into
  this plan before coding.
- The Codex plugin manifest (`.codex-plugin/plugin.json`) has no official page found:
  not used. `.agents/skills/` is enough for Codex to find the skill.

`.agents/skills/` is read by both hosts, so the skill is written there once for both.

## Packages and files

| Path | Change |
|---|---|
| `LICENSE`, `NOTICE` | new (FR-001) |
| `.claude-plugin/plugin.json`, `flake.nix`, `skills/noodle-diagram/SKILL.md` | `license` field (FR-001) |
| `assets/icons/ATTRIBUTION.md`, `assets/attribution_test.go` | new (FR-002) |
| `specs/010-project-distribution/fork-triage.md` | new (FR-003) |
| `content.go` (root package `noodle`) | new: `go:embed skills agents`, so the binary carries them (FR-005, FR-006) |
| `cmd/noodle/setup.go`, `setup_test.go` | new: `noodle setup opencode\|codex -project\|-user [DIR]` (FR-005 to FR-007) |
| `skills/noodle-diagram/SKILL.md` | Claude Code-only wording moved under "In Claude Code" (FR-004) |
| `tools/speccheck/skills.go` | new: `speccheck skills`, run by `all` (FR-004, FR-015) |
| `README.md` | install per host, positioning, links (FR-007, FR-008, FR-010) |
| `docs/VISION.md` | positioning paragraph (FR-010) |
| `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, `.github/ISSUE_TEMPLATE/` | new (FR-008) |
| `docs/guide/` | new: tutorial, how-to, explanation, `cli.md` (FR-009) |
| `cmd/noodle/cli_doc_test.go` | new: `cli.md` equals the usage text (FR-009) |
| `skills/noodle-diagram/reference.md` | "Discovery" section (FR-011) |
| `cmd/noodle/main.go` | `version` variable and subcommand (FR-012) |
| `.goreleaser.yaml`, `.github/workflows/release.yml`, `cliff.toml`, `CHANGELOG.md` | new, if clarification 3 keeps that pair (FR-013) |
| `bin/noodle`, `cmd/noodle/wrapper_test.go` | download fallback; a Go test runs the script with a stub `curl` and fixture archives (FR-014) |
| process skills, per clarification 4 | three `SKILL.md` (FR-015) |
| `flake.nix` | `goreleaser`, `git-cliff` in the dev shell |

## Fixed choices for implementation

- Read-only agent: an agent whose `tools` list has neither `Write` nor `Edit`. Today
  `fact-finder` and `diagram-reviewer`; `icon-curator` is not.
- OpenCode agent frontmatter written by setup: `mode: subagent`, and for read-only
  agents `permission:` with `edit: deny`.
- Marker: the first line of the Markdown body (after frontmatter) for `.md` files, and the
  first comment line for `.toml` files: `noodle setup <host> <noodle version>, do not edit:
  rerun noodle setup`. Never a frontmatter key, so the skill rules hold (FR-004).
- `noodle setup` prints one `wrote <path>` line per file written, nothing for files left
  as they were.
- Golden files: `cmd/noodle/testdata/setup/opencode/` and `cmd/noodle/testdata/setup/codex/`.
- Plugin version: a `version` field in `.claude-plugin/plugin.json`, set at each release;
  `bin/noodle` downloads that version.
- New Makefile targets, listed in COMMANDS.md: `make cli-doc` (writes `docs/guide/cli.md`),
  `make changelog` (writes `CHANGELOG.md`), `make release-snapshot` (builds `dist/`, no
  publishing). `make release-snapshot` joins the evaluator allowlist as an exact entry;
  it writes only `dist/`, which is git-ignored. This extends the set DEV-05 reviewed and
  is recorded in ADR-0024's accepted risks by T13.
- Fork build (clarification 2, recommended answer): the fork is cloned into a temporary
  directory and tested with `podman run --rm --network none -v <dir>:/src:ro -w /src
  golang:1.26 go test -mod=vendor ./...`; the fork vendors its dependencies, so no
  network is needed.

## Rules of `speccheck skills`

Skills checked: `skills/*/SKILL.md` and the process skills of FR-015. Claude Code-only
names: `noodle:fact-finder`, `noodle:diagram-reviewer`, `noodle:icon-curator`,
`the Agent tool`, `Task tool`, `/plugin`. They are allowed under a heading that starts
with "In Claude Code".

## Alternatives set aside

- Hand-written adapters per host committed in the repository: they drift from
  `agents/`. Generated by `noodle setup` instead, so the binary is the one source.
- A `.codex-plugin/plugin.json`: no official reference found.
- Docs generated for contracts and rules now: they wait for `noodle check` (#95).

## Risks

| Risk | Mitigation |
|---|---|
| Host formats change | sources cited above; the marker line names the format version; setup tests use golden files that a format change updates in one place |
| Executing the fork's code | clarification 2; default is a container without network |
| Download fallback fetches a tampered binary | checksum verified against the release's checksums file; a mismatch refuses to run (FR-014); signing with cosign left for later |
| Attribution of an icon is wrong | each row cites its source URL so a reviewer can check it |
