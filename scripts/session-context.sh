#!/usr/bin/env bash
# SessionStart hook: its stdout is added to Claude's context, so a new, resumed, cleared
# or compacted session starts knowing where the work stands. Keep it fast.
set -euo pipefail
cd "${CLAUDE_PROJECT_DIR:-$(git rev-parse --show-toplevel)}"

echo "# noodle · session context (injected by the SessionStart hook)"
echo
awk '/^## Start here/{on=1} /^## Open threads/{on=0} on' docs/PROGRESS.md
echo "## Next features (docs/features.json)"
jq -r '.phases[] | .features[] | select(.passes == false) | "- \(.id) \(.description)"' docs/features.json | head -5
echo
echo "## Git"
echo "- branch $(git branch --show-current), last commit: $(git log -1 --format='%h %s (%cr)')"
dirty="$(git status --porcelain | wc -l)"
if [[ "$dirty" -gt 0 ]]; then
  echo "- $dirty uncommitted change(s): a previous session stopped mid-work, check them first"
fi
echo
echo "Run ./scripts/init.sh before changing code. Protocol: AGENTS.md."
