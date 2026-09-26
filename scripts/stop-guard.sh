#!/usr/bin/env bash
# Stop hook: keeps docs/PROGRESS.md in step with the work. When files changed and the
# progress log did not, it blocks the stop once (exit 2) and asks Claude to record where
# things stand. stop_hook_active prevents a loop.
set -euo pipefail
input="$(cat)"
[[ "$(jq -r '.stop_hook_active // false' <<<"$input")" == "true" ]] && exit 0
cd "${CLAUDE_PROJECT_DIR:-$(git rev-parse --show-toplevel)}"

changes="$(git status --porcelain --untracked-files=all | grep -v 'docs/PROGRESS.md$' || true)"
[[ -z "$changes" ]] && exit 0
git status --porcelain -- docs/PROGRESS.md | grep -q . && exit 0

cat >&2 <<'MSG'
Files changed but docs/PROGRESS.md did not. Before stopping, record where things stand:
- add a line under today's log entry (what changed, what is next);
- if a feature's verify step passed, flip it in docs/features.json and commit (AGENTS.md).
If you are only pausing to ask Brian something, a one-line "in progress" note is enough.
MSG
exit 2
