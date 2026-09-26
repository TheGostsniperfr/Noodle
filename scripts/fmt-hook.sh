#!/usr/bin/env bash
# PostToolUse hook: gofmt every Go file Claude writes, so formatting never shows up in review.
set -euo pipefail
file="$(jq -r '.tool_input.file_path // empty')"
if [[ "$file" == *.go && -f "$file" ]] && command -v gofmt >/dev/null; then
  gofmt -w "$file"
fi
exit 0
