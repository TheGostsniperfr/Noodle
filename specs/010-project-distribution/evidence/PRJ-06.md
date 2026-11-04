# Evidence · PRJ-06

- Feature: PRJ-06
- Verdict: **NEEDS_WORK**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: db81c22
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Guide pages exist | PASS | `ls docs/guide/` lists exactly `cli.md explanation.md howto-diff.md howto-discover.md howto-hosts.md howto-slides.md tutorial.md` |
| 2 | The tutorial's system lints and the page uses it | PASS | `ls examples/tutorial` shows `layouts model.yaml request.dark.png views`. The grep matched `Makefile:56: $(NOODLE) render examples/tutorial`, and also `docs/guide/tutorial.md:4` and `:105`. `make examples` gave EXIT=0 and ran `go run ./cmd/noodle render examples/tutorial` |
| 3 | The CLI reference is current and cannot go stale | PASS | `go test ./cmd/noodle/ -run CLIReference -v` showed `--- PASS: TestCLIReference/current` and `--- PASS: TestCLIReference/stale` with EXIT=0. The `stale` subtest asserts that the error contains `docs/guide/cli.md`. Other matches: `cli.md:30:  -namespace string`, `Makefile:65:cli-doc:`, and `COMMANDS.md:17` has the `make cli-doc` row |
| 4 | Positioning is honest | PASS | P6a-05 `passes` is `false`. The README grep has one match, `README.md:6:diagrams that are designed to stay in sync with the systems they describe` |
| 5 | VISION names the established terms | **FAIL** | `grep -n -e 'architecture as code' -e 'living documentation' docs/VISION.md` printed nothing and exited 1. Each count is `0`. The terms appear only capitalised, at `VISION.md:35` (`**Architecture as code.**`) and `VISION.md:38` (`**Living documentation.**`). The contract's command is case-sensitive, so the threshold "both found" is not met |
| 6 | Discovery section covers #88 | PASS | `## Discovery` runs from line 187 to `## Layout` at line 236. All 13 literals match inside that range: `#71`:212, `-namespace`:196, kustomize/helm (-i):197, `matches:`:201, `not discovered`:207, `#78`:217, `zone/**`:219, `#74`:223, `tech:`:225, `status: planned`:227, `target:`:228, `fact-finder`:230 |

## Verify step

    $ go test ./cmd/noodle/ -run CLIReference -v
    --- PASS: TestCLIReference (0.00s)
        --- PASS: TestCLIReference/current (0.00s)
        --- PASS: TestCLIReference/stale (0.00s)
        --- PASS: TestCLIReference/lists_every_flag_of_every_subcommand (0.00s)
    PASS
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	(cached)

    $ ls docs/guide/
    cli.md explanation.md howto-diff.md howto-discover.md howto-hosts.md howto-slides.md tutorial.md

The verify step passes. Contract criterion 5 does not, so the feature does not pass.

## Findings

- `docs/VISION.md:35,38`: contract PRJ-06 #5 runs a case-sensitive grep for `architecture as code` and `living documentation`, and it finds neither. Both terms appear only as capitalised bold lead-ins, and the lowercase sentence that names them is in `README.md`, not VISION. Fix: add a sentence to the "Where it sits" section of VISION that uses both terms in lowercase, for example "noodle sits between architecture as code and living documentation." Do not loosen the criterion.
- Tests that would catch a revert:
  - FR-009: `cmd/noodle/cli_doc_test.go` `stale` changes `-namespace` to `-ns` and asserts the error names `docs/guide/cli.md`. `current` compares the committed page byte for byte with `cliReference()`. The render, discover and setup commands build their flags with the same constructors that `help.go` uses (`newRenderFlags`, `newDiscoverFlags`, `newSetupFlags`, `newListIconsFlags`). A new flag that is not regenerated into `cli.md` therefore fails `current`. These tests would catch a revert.
  - FR-010 and FR-011 are docs-only. No Go test covers them. Only the contract greps check them, and those greps would catch a revert.
- `docs/guide/tutorial.md:25-79`: no check ties the YAML blocks in the tutorial to `examples/tutorial/*`. Today they are identical, apart from the `meta` source path, which differs on purpose. The page can drift from the system that `make examples` lints. Suggestion, not graded: add a test that extracts the fenced blocks and compares them to the files.
- `examples/tutorial/request.dark.png`: I read it. It shows two zones, INTERNET and CLUSTER. Inside them, User → `1 GET · HTTPS` → `TCP 443` on API, and API → `2 query · SQL` → `TCP 5432` on Postgres. Each listening port sits on the server side (C02). There are no crossings, and the labels are legible. Two minor points that break no criterion:
  - The legend lists "⚠ Gx: docs and code disagree, see Gaps", but the view has no gaps.
  - The CloudNativePG icon touches the top rim of the cylinder.
- Scope: the PRJ-06 commits (7f7ce9c, e503a61) stay within `cmd/noodle`, `docs/`, `examples/tutorial`, `skills/noodle-diagram`, the setup goldens, `Makefile` and `README.md`, which matches T11 and T12. Nothing outside the plan. No `features.json` entry is edited, and no lint rule is relaxed.
