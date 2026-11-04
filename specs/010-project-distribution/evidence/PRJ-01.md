# Evidence · PRJ-01

- Feature: PRJ-01
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 9adb43a
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | One SPDX id in plugin and skill | PASS | `jq -r .license .claude-plugin/plugin.json` printed `Apache-2.0`. `grep -n '^license:' skills/noodle-diagram/SKILL.md` printed `4:license: Apache-2.0` |
| 2 | The license text is Apache-2.0 | PASS | `grep -c 'Apache License' LICENSE` printed `4`, which meets ≥1. `grep -c 'Version 2.0, January 2004' LICENSE` printed `1`, which meets exactly 1 |
| 3 | The flake declares it | PASS | `grep -n 'license' flake.nix` gave one match: `23:            license = pkgs.lib.licenses.asl20;` |
| 4 | NOTICE names product, author and URL | PASS | `cat NOTICE` contains `Noodle`, `Copyright 2026 Brian Perret` and `https://github.com/TheGostsniperfr/Noodle` |
| 5 | The attribution test reads the embedded icons and fails per missing piece | PASS | `go test ./assets/ -run Attribution -v -count=1` exited 0. These subtests pass: `repository`, `missing_row`, `row_without_url`, `row_without_license`, `row_for_unknown_file` and `no_trademark_notice`. In `attribution_test.go:96-97,105-106,114-115`, each broken-row subtest requires `Len(errs, 1)` and that the error contains `cilium.svg`. The file list comes from `fs.ReadDir(assets.Icons, "icons")` at `attribution_test.go:18` |
| 6 | One row per icon file | PASS | `grep -c '^[|] `' assets/icons/ATTRIBUTION.md` printed `45`. `ls assets/icons/*.svg \| wc -l` printed `45` |
| 7 | Trademark notice | PASS | `grep -ci 'trademark' assets/icons/ATTRIBUTION.md` printed `12`, which meets ≥1 |

## Verify step

    $ go test ./assets/ -run Attribution -v -count=1
    --- PASS: TestAttribution (0.00s)
        --- PASS: TestAttribution/repository (0.00s)
        --- PASS: TestAttribution/missing_row (0.00s)
        --- PASS: TestAttribution/row_without_url (0.00s)
        --- PASS: TestAttribution/row_without_license (0.00s)
        --- PASS: TestAttribution/row_for_unknown_file (0.00s)
        --- PASS: TestAttribution/no_trademark_notice (0.00s)
    ok  	github.com/TheGostsniperfr/Noodle/assets	0.003s
    exit=0

    $ jq -r .license .claude-plugin/plugin.json ; grep -n '^license:' skills/noodle-diagram/SKILL.md ; grep -n 'license' flake.nix
    Apache-2.0
    4:license: Apache-2.0
    23:            license = pkgs.lib.licenses.asl20;
    (LICENSE: "Version 2.0, January 2004" x1, "Apache License" x4)

    $ make spec-check
    go run ./tools/speccheck all
    exit=0

## Findings

**Test review.** I checked whether the tests would fail if the feature were reverted. For FR-002 they would: deleting `ATTRIBUTION.md` makes `os.ReadFile` fail at `attribution_test.go:84`, and deleting any row fails `repository`. Adding a new icon with no row also fails `repository`, because the file list comes from the embedded FS and not from a hard-coded list. FR-001 has no Go test. Its fields are checked only by the contract commands above, which is what the contract asks for.

**Scope.** I checked commits 6e54b29 and d6397c7. They touch only the license files, the license fields, `assets/`, PROGRESS and tasks. `docs/features.json` only gains entries and changes no existing one, so C12 holds.

**Minor findings.** None of these break a contract criterion:

- `assets/attribution_test.go:52`: the source check only looks for an `https://` prefix, so any https URL passes, even one that points to a page rather than the file. Two rows do this: `loki.svg` points to `https://grafana.com/oss/loki/` and `dagster.svg` points to `https://dagster.io/brand`. Pointing those rows at the files themselves would make them traceable.
- `assets/icons/ATTRIBUTION.md`: the source URL in the `sonarqube.svg` row is the redrawn upstream file `sonarqubeserver.svg`, not the revision that was committed. The notes section says so. Pinning the simple-icons URLs to a commit SHA rather than `develop` would make every row reproducible.
- The `features.json` verify text says "the four license fields", but the contract checks three fields plus the LICENSE text. The wording is inconsistent, but nothing is unverified.
