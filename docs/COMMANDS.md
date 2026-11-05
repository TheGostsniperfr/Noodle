# Noodle · Commands

Every command for working on this repository is a `make` target (C14). This page is the
reference for agents and humans; `make help` prints the same list, and `make spec-check`
fails when the two differ. Process: [`PROCESS.md`](PROCESS.md).

## Everyday

| Command | When | Arguments | A failure means |
|---|---|---|---|
| `make init` | first thing in every session | none | something is red on the branch: fix it, or record it in PROGRESS, before new work |
| `make check` | at the end of every task, before a commit | none | a test, an example, the features gate or spec-check failed; the first error names the file |
| `make vet` | after a Go change, fast | none | `go vet` finding |
| `make test` | after a Go change | none | a failing test |
| `make examples` | after a change to model, lint, render or discovery | none | an example no longer lints, or discovery no longer reproduces `examples/discovery-k8s`'s golden fragment: fix the change, never the rule (C04) |
| `make render` | to look at a diagram | `SYS=<dir>` `VIEW=<id>` `OUT=<file.drawio>` `THEME=dark\|light` `ICONS=<dirs>` `SLIDE=1` | the view fails its lint; without `OUT` nothing is written |
| `make cli-doc` | after adding or changing a flag or a subcommand | none | `go test ./cmd/noodle/` fails with `docs/guide/cli.md` differing from `noodle help`: run it and commit the page |
| `make changelog` | before a release, and to preview the next one | none | `git-cliff` failed or a commit message is not conventional; fix the message, not the tool |
| `make release-snapshot` | to check the release build locally | none | goreleaser could not build a platform; nothing is published and `dist/` is git-ignored |
| `make help` | to list targets | none | never fails |
| `make next` | to see what to make pass next | none | never fails |

## Specs, ADRs and features

| Command | When | Arguments | A failure means |
|---|---|---|---|
| `make spec-check` | after editing a spec, a plan, tasks, a contract, an ADR or this page | none | a requirement, task, feature, contract, ADR or command is missing its other half; the message names both |
| `make adr-index` | after adding or changing the status of an ADR | none | an ADR has no `# ADR-NNNN · Title` heading |
| `make features` | after editing `docs/features.json`, also run by a hook | `BASE=<ref>`, default `HEAD`; use `BASE=origin/main` before a PR | an entry was removed or reworded, or a feature passes without `evidence/<id>.md` saying PASS |
| `make review-contract` | before implementing a full-mode spec | `SPEC=<NNN>` | the review has no `- Review:` line; read `contract-review.md`, apply CHANGES, rerun |
| `make evaluate` | once a feature's tasks are done, before flipping it | `FEATURE=<id>` | the evaluator printed no verdict; on NEEDS_WORK, add its findings as Convergence tasks and fix them |

## Hooks

Run by Claude Code (`.claude/settings.json`), never by hand. They are implementations of
hooks, not commands, so they are not `make` targets.

| Script | Hook | Does |
|---|---|---|
| `scripts/session-context.sh` | SessionStart | prints "Start here", the next features and the git state into the session |
| `scripts/stop-guard.sh` | Stop | blocks ending a turn once when files changed but `docs/PROGRESS.md` did not |
| `scripts/fmt-hook.sh` | PostToolUse | gofmt on every Go file written |
| (inline) `make features` | PostToolUse | runs the features gate when `docs/features.json` is edited |

## CI and tools

| Command | When | Arguments | A failure means |
|---|---|---|---|
| `make ci` | to reproduce CI locally | `BASE=<sha>` | same as CI |
| `make vendor-check` | after touching `go.mod` | none | run `go mod vendor` and commit `vendor/` |
| `make plugin` | after editing `.claude-plugin/`, `skills/` or `agents/` | none | the manifest does not validate |
| `make discover-eval` | to measure discovery on real repositories (read-only) | `REPOS="<dir>…"` `MATCH=<regex>` | a repository failed to render or discover |
