# Evidence · PRJ-07

- Feature: PRJ-07
- Verdict: **PASS**
- Evaluator: evaluator (claude-opus-5-5)
- Commit: 9adb43a
- Date: 2026-11-04

## Criteria

| # | Criterion | Result | Evidence |
|---|---|---|---|
| 1 | `noodle version` reads the build variable | PASS | `go test ./cmd/noodle/ -run Version -v -count=1` → `--- PASS: TestVersion_PrintsTheBuildVariable/a_release_build_prints_its_tag`, `--- PASS: .../a_plain_build_prints_dev`, `ok`. The test builds real binaries with and without `-X main.version=v0.1.0` and asserts `v0.1.0\n` and `dev\n` (`cmd/noodle/version_test.go:31-43`) |
| 2 | The flake passes its revision | PASS | `grep -n 'main.version' flake.nix` → one match: `20: ldflags = [ "-X main.version=${self.shortRev or self.dirtyShortRev or "dev"}" ];` |
| 3 | Release workflow on tags only, body from the changelog | PASS | `5: tags:` / `6: - "v*"`; `24: uses: orhun/git-cliff-action@v4`; `40: body_path: RELEASE_NOTES.md`; the `branches`/`pull_request` grep has no match (rc=1) |
| 4 | Four platforms and checksums configured | PASS | `.goreleaser.yaml` `14: goos: [linux, darwin]`, `15: goarch: [amd64, arm64]`, `27: checksum:` |
| 5 | A snapshot builds now | PASS | `make release-snapshot` exit=0 (`release succeeded`). `ls dist/*.tar.gz` lists exactly `noodle_0.0.1-snapshot_{darwin,linux}_{amd64,arm64}.tar.gz` (4). `dist/noodle_0.0.1-snapshot_checksums.txt` has 4 lines, one per archive |
| 6 | Changelog from commits | PASS | `68:changelog:` in Makefile. CHANGELOG.md has these `###` headings: Bug fixes, CI, Chores, Documentation, Features, Refactoring, Tests. All seven are groups in `cliff.toml:33-41`. `head` shows `## Unreleased` at the top |
| 7 | Fallback downloads the right archive, verifies, refuses a mismatch | PASS | `go test ./cmd/noodle/ -run Wrapper -v -count=1` → `--- PASS` for `download_verified`, `download_altered`, `go_on_path`, `nix_on_path`, `noodle_on_path`, `download_cached_is_not_downloaded_twice` and `unsupported_platform_does_not_download`. `download_verified` asserts exact URLs built from the `plugin.json` version and `linux_amd64` from a stubbed `uname`. `download_altered` asserts a non-zero exit, `checksum mismatch` on stderr, no `noodle-*` in the cache and no `download.*` leftovers |

## Verify step

    $ go test ./cmd/noodle/ -run 'Version|Wrapper' -v -count=1
    --- PASS: TestWrapper (0.00s)
        --- PASS: TestWrapper/noodle_on_path (0.03s)
        --- PASS: TestWrapper/nix_on_path (0.04s)
        --- PASS: TestWrapper/unsupported_platform_does_not_download (0.04s)
        --- PASS: TestWrapper/go_on_path (0.05s)
        --- PASS: TestWrapper/download_altered (0.07s)
        --- PASS: TestWrapper/download_verified (0.09s)
        --- PASS: TestWrapper/download_cached_is_not_downloaded_twice (0.12s)
    --- PASS: TestVersion_PrintsTheBuildVariable (0.00s)
        --- PASS: TestVersion_PrintsTheBuildVariable/a_release_build_prints_its_tag (0.24s)
        --- PASS: TestVersion_PrintsTheBuildVariable/a_plain_build_prints_dev (0.24s)
    PASS
    ok  	github.com/TheGostsniperfr/Noodle/cmd/noodle	0.246s

    $ make release-snapshot
      • archiving   name=dist/noodle_0.0.1-snapshot_linux_arm64.tar.gz
      • archiving   name=dist/noodle_0.0.1-snapshot_linux_amd64.tar.gz
      • archiving   name=dist/noodle_0.0.1-snapshot_darwin_amd64.tar.gz
      • archiving   name=dist/noodle_0.0.1-snapshot_darwin_arm64.tar.gz
      • calculating checksums
      • release succeeded after 0s
    exit=0

    $ cat dist/*checksums.txt
    4769…b2e0  noodle_0.0.1-snapshot_darwin_amd64.tar.gz
    4f5c…6911  noodle_0.0.1-snapshot_darwin_arm64.tar.gz
    c9c0…7afa  noodle_0.0.1-snapshot_linux_amd64.tar.gz
    664d…f99f  noodle_0.0.1-snapshot_linux_arm64.tar.gz

    $ tar tzf dist/noodle_0.0.1-snapshot_linux_amd64.tar.gz
    LICENSE
    NOTICE
    noodle

Would the tests catch the feature being reverted? I read each test with that question:
- FR-012: `TestVersion_PrintsTheBuildVariable` builds the binary and checks its stdout. It fails if `main.version` is removed or `version` stops printing it.
- FR-014: `download_verified` fails without the download branch, because the stub archive's output and the exact curl URLs are asserted. `download_altered` fails without the checksum comparison, because it asserts a non-zero exit and an empty cache. `go_on_path`, `nix_on_path` and `noodle_on_path` fail if the fallback ran first.
- FR-013: the contract checks it through configuration greps and a real snapshot build, which I reran.

## Findings

- `.github/workflows/release.yml:23-28`: the release body comes from regenerating notes with `git-cliff --latest`, not from reading the tag's section of the committed `CHANGELOG.md` as FR-013 words it. The content is the same when `CHANGELOG.md` is current, and criterion 3 only requires git-cliff and `body_path`. This breaks no criterion. Fix: either extract the tag's section from `CHANGELOG.md`, or reword FR-013 in a later spec.
- `.goreleaser.yaml:16`: `main.version={{ .Tag }}`. A snapshot binary prints `v0.0.0`; I ran it from `dist/`. This is expected without a tag. A real tag `vX.Y.Z` yields `vX.Y.Z`, and the archive name uses `{{ .Version }}` without the `v`, which matches the `noodle_${version}_…` the wrapper builds from `plugin.json`. Nothing to fix. I note it so nobody reads `v0.0.0` as a bug.
- `docs/adr/0024-spec-driven-process-with-evaluation.md:73-75`: commit 5e8f673 appends a paragraph to an ADR whose status is `accepted`. AGENTS.md says "Never rewrite an accepted ADR; supersede it." ADR-0024 is not on `origin/main` yet, and T13 asked for the risk to be noted there, so this is acceptable before merge. After merge, the same change would need a new ADR.
- The action versions in `release.yml` (`actions/checkout@v7`, `actions/setup-go@v7`, `goreleaser/goreleaser-action@v7`, `softprops/action-gh-release@v3`) cannot be checked offline. The first real tag push is the only end-to-end proof that the workflow publishes. No criterion covers it.
- Scope: the three PRJ-07 commits (422c13a, 5e8f673, 06c4f79) touch only files that T06, T13 and T14 name, plus PROGRESS and tasks ticks. `docs/features.json` only gains lines in this branch, so C12 holds. No lint rule was touched, so C04 holds.
- Stop hook: I did not edit `docs/PROGRESS.md` or `docs/features.json`, because the evaluator is read-only. The only files this run wrote are in `dist/`, from `make release-snapshot`, and `/dist/` is git-ignored. The caller records this evidence, flips PRJ-07 and adds the PROGRESS line.
