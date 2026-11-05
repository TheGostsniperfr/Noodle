# Evidence · PRJ-06

- Feature: PRJ-06
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: a51d77b
- Date: 2026-11-05

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | Guide pages exist | PASS | `ls docs/guide/` printed exactly `cli.md explanation.md howto-diff.md howto-discover.md howto-hosts.md howto-slides.md tutorial.md` |
| 2 | The tutorial's system lints and the page uses it | PASS | `ls examples/tutorial` printed `layouts model.yaml request.dark.png views`. `Makefile:56: $(NOODLE) render examples/tutorial` and `docs/guide/tutorial.md:4` and `:101` both reference it. `make examples >/dev/null 2>&1; echo $?` printed `0`. `git status` stayed clean afterwards, so the committed PNG matches a fresh render. |
| 3 | The CLI reference is current and cannot go stale | PASS | `go test ./cmd/noodle/ -run CLIReference -v` printed `--- PASS: TestCLIReference/current`, `--- PASS: TestCLIReference/stale` and `ok`. `current` compares the whole of `cli.md` with `cliReference()`, which is built from `fullUsage()` and includes every subcommand's `FlagSet.PrintDefaults` (help.go:41-69). `stale` changes `-namespace` to `-ns` and asserts the error contains `docs/guide/cli.md`. `docs/guide/cli.md:30: -namespace string` was found. `Makefile:65: cli-doc:` was found. `docs/COMMANDS.md:17` has a `make cli-doc` row. |
| 4 | Positioning is honest | PASS | jq printed `false` for P6a-05. There is one match, and it contains the required phrase: `README.md:6: diagrams that are designed to stay in sync with the systems they describe` |
| 5 | VISION names the established terms | PASS | `docs/VISION.md:33: Two established terms say what noodle is for: architecture as code and living documentation.` |
| 6 | Discovery section covers #88 | PASS | The `## Discovery` heading is at line 187 and the next `## ` is at line 236. The first match of each literal falls inside that range: `#71`:212, `-namespace`:196, `kustomize`/`helm` (-i):197, `matches:`:201, `not discovered`:207, `#78`:217, `zone/**`:219, `#74`:223, `tech:`:225, `status: planned`:227, `target:`:228, `fact-finder`:230 |

## Verify step

    $ go test ./cmd/noodle/ -run CLIReference -v
    --- PASS: TestCLIReference (0.00s)
        --- PASS: TestCLIReference/current (0.00s)
        --- PASS: TestCLIReference/lists_every_flag_of_every_subcommand (0.00s)
        --- PASS: TestCLIReference/stale (0.00s)
    PASS
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.008s
    $ go test ./cmd/noodle/
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.697s
    $ ls docs/guide/
    cli.md explanation.md howto-diff.md howto-discover.md howto-hosts.md howto-slides.md tutorial.md

## Findings

Tests that test nothing:
- FR-009: `TestCLIReference/current` does catch a regression. A flag added to `render`, `discover`, `setup` or `-list-icons` without running `make cli-doc` changes `fullUsage()` and fails the equality check, which is the scenario the spec gives. `TestTutorial_PageShowsTheSystemOfExamplesTutorial` (cmd/noodle/tutorial_test.go) fails if the YAML on the page drifts from `examples/tutorial/*`.
- FR-010 and FR-011 have no Go test, only the contract's greps. That matches the contract and is not a defect.

Visual (`examples/tutorial/request.dark.png`, read directly):
- The diagram is readable: two zones (INTERNET, CLUSTER), User → API → Postgres, numbered steps 1 and 2 with protocol labels, port badges `TCP 443` and `TCP 5432`, and no legend card (T17).
- Minor: on edge 2, the label `query · SQL` runs into the arrowhead, and the arrowhead runs into the `TCP 5432` badge, with no visible gap. The elephant icon also pokes above the top ellipse of the Postgres cylinder. Neither breaks a criterion, and the lint passes. A small edge-label inset would fix the first.

Scope and constitution:
- The PRJ-06 files of the diff are under `docs/guide/`, `README.md`, `docs/VISION.md`, `skills/noodle-diagram/reference.md`, `cmd/noodle/help.go`, the two tests, `examples/tutorial/`, `Makefile` and `docs/COMMANDS.md`. All of them are in plan scope.
- In `docs/features.json` the `proj` phase is added as a whole in this branch. No existing entry is edited (C12), and no lint rule is relaxed (C04).

Not graded:
- Round-3 follow-up: `explanation.md` has the headings Model, View, Layout, Fragment and Connection semantics, so the follow-up is already met.
