# Spec 010 · Project and distribution: license, contribution, hosts, docs, releases

- Status: **Draft**
- Mode: **full**
- Phase: proj, batch E of wave 1, pulled forward by ADR-0025 (C11)
- Issues: #93 (B-42), #94 (B-58), #100 (B-50), #106 (B-56), #107 (B-57), #88 (B-39)
- Decisions it builds on: ADR-0024, ADR-0025
- Invariants it touches: C05, C08, C09, C12, C14, C15

## Why

Noodle can be installed, but not legally reused, contributed to, run outside Claude
Code, or pinned to a version:

- **No license.** The default is "all rights reserved". The first external contribution
  (NoodleOpenCode, #94) arrived with no stated terms, and the 45 logos in `assets/icons/`
  carry their own terms that nobody lists.
- **One host.** The skill and the three subagents ship as a Claude Code plugin only. The
  fork author had to fork to run them in OpenCode. The skill body names Claude Code
  tools (`noodle:fact-finder`, "the Agent tool").
- **Docs for builders, not users.** VISION, ARCHITECTURE, 25 ADRs and seven specs explain
  how noodle is built; nothing explains how to use it, contribute, or report a security
  issue. `skills/noodle-diagram/reference.md` has no discovery section (#88).
- **No versions.** The plugin follows `main`, there are no tags, no changelog, no
  `noodle version`, and a machine without Go or Nix cannot run `bin/noodle` at all.

## User scenarios

**A company engineer checks whether noodle may be used at work.** They open the
repository, see the license in the sidebar, read `NOTICE` and `assets/icons/ATTRIBUTION.md`,
and know what they may do with the code and that logos remain their owners' trademarks.

**The fork author offers a piece of their work.** They read the triage of their branch,
see which issue each idea went to, confirm their contribution is under the project
license, and open a small PR against `main` following `CONTRIBUTING.md`.

**An OpenCode user installs noodle.** They install the `noodle` binary, run
`noodle setup opencode -project` in their repository, and OpenCode lists the
`noodle-diagram` skill and the three subagents. A Codex user does the same with
`noodle setup codex -project`.

**A newcomer draws a first system.** They follow `docs/guide/tutorial.md` from a fresh
machine to a rendered PNG, then look up a flag in the generated CLI reference.

**Brian cuts a release.** He tags `v0.1.0`; CI publishes binaries for four platforms,
checksums and the changelog section. A machine without Go or Nix runs `bin/noodle`,
which downloads and verifies the matching binary.

## Requirements

### License and attribution (#93)

- **FR-001** THE repository SHALL carry a `LICENSE` file with the license chosen in
  clarification 1, and WHERE that license is Apache-2.0 a `NOTICE` file naming
  "Noodle", "Brian Perret" and the repository URL; `.claude-plugin/plugin.json`
  (`license`), `flake.nix` (`meta.license`) and the frontmatter of
  `skills/noodle-diagram/SKILL.md` (`license`) SHALL name the same SPDX identifier.
  *Scenario:* WHEN `jq -r .license .claude-plugin/plugin.json` runs THEN it prints the
  SPDX identifier that the first line of `LICENSE` names.
- **FR-002** IF a file in `assets/icons/` has no row in `assets/icons/ATTRIBUTION.md`
  giving its source URL and license, THEN THE SYSTEM SHALL fail `go test ./assets/`,
  naming the file; and the attribution page SHALL state that logos remain trademarks of
  their owners whatever the code license.
  *Scenario:* WHEN `cilium.svg` loses its row THEN `go test ./assets/` fails with a
  message naming `cilium.svg`.

### The NoodleOpenCode contribution (#94)

- **FR-003** THE SYSTEM SHALL record the triage of
  `AlexandreGirold/NoodleOpenCode@feat/project-architecture-mermaid` in
  `specs/010-project-distribution/fork-triage.md`: the commits examined (`6d61ce0`,
  `2541717`), the build and test result obtained as decided in clarification 2, one
  disposition per row of #94's table (a target issue, or "declined" with its reason),
  and a draft reply to the author; and IF a commit authored by the fork author reaches
  `main` before the commit that adds `LICENSE`, THEN that order SHALL be reported as a
  failure of this requirement.
  *Scenario:* WHEN `fork-triage.md` is read THEN each of the seven rows of #94's table
  has a disposition, and `git log main --author=AlexandreGirold` lists no commit older
  than the `LICENSE` commit.

### Hosts: OpenCode and Codex (#100)

- **FR-004** IF `skills/noodle-diagram/SKILL.md` or a process skill breaks the Agent
  Skills rules (a `name` equal to its folder name, lowercase with hyphens, at most 64
  characters; a `description` of 1 to 1024 characters; no frontmatter key outside
  `name`, `description`, `license`, `metadata`), or names a Claude Code-only tool outside
  a section titled "In Claude Code", THEN THE SYSTEM SHALL fail `make spec-check` with
  the file and the rule.
  *Scenario:* WHEN the skill's `description` is longer than 1024 characters THEN
  `make spec-check` fails naming `skills/noodle-diagram/SKILL.md` and the length rule.
- **FR-005** WHEN `noodle setup opencode -project DIR` runs, THE SYSTEM SHALL write the
  `noodle-diagram` skill to `DIR/.agents/skills/noodle-diagram/` and one
  `DIR/.opencode/agents/<name>.md` per agent of `agents/`, with OpenCode frontmatter
  (`mode: subagent`, edits denied for read-only agents) and the agent's instructions,
  print every path written, and leave every file unchanged on a second run; and WHERE
  `-user` is given instead of `-project`, it SHALL write under `~/.agents/skills/` and
  `~/.config/opencode/agents/`.
  *Scenario:* WHEN `noodle setup opencode -project` runs twice in an empty directory
  THEN the second run writes nothing and both runs list the same four paths.
- **FR-006** WHEN `noodle setup codex -project DIR` runs, THE SYSTEM SHALL write the
  `noodle-diagram` skill to `DIR/.agents/skills/noodle-diagram/` and one
  `DIR/.codex/agents/<name>.toml` per agent of `agents/`, in the format of the official
  Codex subagents page cited in the plan, print every path written, and be idempotent;
  WHERE `-user` is given, it SHALL write under `~/.agents/skills/` and `~/.codex/agents/`.
  *Scenario:* WHEN `noodle setup codex -project` runs in an empty directory THEN
  `.codex/agents/fact-finder.toml` equals its golden file, which carries the agent's
  instructions in the field the Codex page names.
- **FR-007** IF a file that `noodle setup` would write already exists and was not
  written by `noodle setup` (it lacks the `noodle setup` marker line), THEN THE SYSTEM
  SHALL leave it unchanged, report it, and exit non-zero; and the README SHALL have one
  install section per host: Claude Code, OpenCode, Codex, Nix.
  *Scenario:* WHEN `.opencode/agents/fact-finder.md` holds a user's own text THEN
  `noodle setup opencode -project` keeps it byte for byte and exits 1 naming it.

### Docs for users and contributors (#106, #88)

- **FR-008** THE repository SHALL have `CONTRIBUTING.md` (spec, contract and evaluation
  workflow of `docs/PROCESS.md`, `make check`, conventional commits, PR visuals),
  `SECURITY.md` (a private reporting channel; an architecture map is sensitive),
  `CODE_OF_CONDUCT.md`, and `.github/ISSUE_TEMPLATE/idea.yml` and `bug.yml` whose fields
  are Context, Proposal, Scope, Out of scope, Acceptance, Related and Backlog id for an
  idea, and What happens, Reproduce, Expected, Version for a bug; the README SHALL link
  the first three.
  *Scenario:* WHEN a new issue is opened on GitHub THEN the chooser offers "Idea" and
  "Bug" with those fields.
- **FR-009** THE repository SHALL have a user guide in `docs/guide/` with a tutorial
  (first system to a rendered PNG), how-to pages (discover from manifests, draw a
  proposal as a diff, export for slides, use in OpenCode or Codex), an explanation page
  (model, view, layout, fragment; connection semantics in plain words), and a CLI
  reference `docs/guide/cli.md` generated from the CLI usage text; IF `cli.md` differs
  from the usage text the binary prints, THEN `go test ./cmd/noodle/` SHALL fail.
  *Scenario:* WHEN a flag is added to `render` and `cli.md` is not regenerated THEN
  `go test ./cmd/noodle/` fails naming `docs/guide/cli.md`.
- **FR-010** THE README SHALL position noodle as architecture as code, checked against
  the real infrastructure, and SHALL NOT claim that diagrams stay in sync until
  `noodle drift` ships (P6a-05); `docs/VISION.md` SHALL relate the idea to the
  established terms "architecture as code" and "living documentation".
  *Scenario:* WHEN the README is read before P6a-05 passes THEN its tagline says
  "designed to stay in sync", not that it stays in sync.
- **FR-011** THE skill reference SHALL have a "Discovery" section covering each item of
  #88: the PNG width caveat of the export command (#71), `-namespace`, rendering
  Kustomize or Helm before discovery, `matches` syntax, what is and is not discovered,
  empty labels that still carry annotation badges (#78), `include` with `zone/**` pulling
  in edges to route, icons resolved only at render (#74), landscape items showing `tech`
  as their second line, planned namespaces (`status: planned` on the zone and `target`
  on children), and what to do when the `fact-finder` agent is not available.
  *Scenario:* WHEN an agent reads `reference.md` for a discovery task THEN it finds each
  of those items under the "Discovery" heading.

### Versions and releases (#107)

- **FR-012** WHEN `noodle version` runs, THE SYSTEM SHALL print the version set at build
  time: the release tag for release builds, the flake's revision for Nix builds, and
  `dev` otherwise.
  *Scenario:* WHEN a binary built with `-ldflags "-X main.version=v0.1.0"` runs
  `noodle version` THEN it prints `v0.1.0`.
- **FR-013** WHEN a tag `vX.Y.Z` is pushed, THE SYSTEM SHALL run a release workflow
  that builds `noodle` for linux and darwin on amd64 and arm64, publishes the archives
  and a SHA-256 checksums file on a GitHub Release, and uses the tag's section of
  `CHANGELOG.md`, generated from conventional commits with the tool chosen in
  clarification 3, as the release body.
  *Scenario:* WHEN the release configuration builds a snapshot locally THEN `dist/`
  holds four archives and one checksums file.
- **FR-014** WHILE neither a `noodle` on PATH, Go nor Nix is available, WHEN `bin/noodle`
  runs, THE SYSTEM SHALL download the release archive matching the plugin's version
  and platform from GitHub Releases, verify it against the release checksums file, and
  run it; IF the checksum does not match, THEN it SHALL refuse to run it and exit non-zero.
  *Scenario:* WHEN the archive served for `v0.1.0` is altered THEN `bin/noodle` prints a
  checksum error, leaves no binary in its cache, and exits non-zero.
- **FR-015** THE repository SHALL ship three process skills for working on noodle,
  each a `SKILL.md` that passes FR-004: `issue-writer` (a raw idea to an issue with the
  `idea.yml` fields and the next free `B-NN`), `release-notes` (a release's changelog to
  highlights, reused PR before/after images, breaking changes with migration links) and
  `contrib-check` (the CONTRIBUTING checklist run on a branch), stored where
  clarification 4 says.
  *Scenario:* WHEN `make spec-check` runs THEN it validates all three skills with FR-004's
  rules.

## Out of scope

- Merging any part of the fork: each piece lands through its own issue (#16, #95, #100,
  #101, #104, a new Compose adapter issue), after FR-001.
- Marketplace channels (`stable`, `edge`) and a hosted docs site.
- Generated references for contracts and lint rules: they follow `noodle check` (#95).
- An MCP server as a host integration (roadmap phase 2b).
- Rendering without draw.io in a cloud session (#91, batch D).

## Clarifications

| # | Question | Answer | ADR |
|---|---|---|---|
| 1 | License: Apache-2.0 with NOTICE, or MIT? | pending | |
| 2 | How is the fork's code built and tested? | pending | |
| 3 | Changelog and release tooling? | pending | |
| 4 | Where do the process skills live? | pending | |
| 5 | Do contributions need a DCO sign-off? | pending | |

## Success criteria

Features `PRJ-*` in `docs/features.json` pass their verify step, graded in `contract.md`.
