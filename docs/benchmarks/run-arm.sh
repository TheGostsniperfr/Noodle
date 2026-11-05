#!/usr/bin/env bash
# Runs one arm of the discovery checkpoint (ADR-0029) as a fresh, isolated claude -p
# process and copies its transcript with every tool result removed (C09).
#   docs/benchmarks/run-arm.sh A|C SNAPSHOT FRAGMENTS_DIR WORKDIR
# SNAPSHOT holds the repositories; arm C also gets FRAGMENTS_DIR as discovered/.
set -euo pipefail

[[ $# -eq 4 ]] || { echo "usage: $0 A|C SNAPSHOT FRAGMENTS_DIR WORKDIR" >&2; exit 2; }
arm="$1" snapshot="$2" fragments="$3" work="$4"
here="$(cd "$(dirname "$0")" && pwd)"
model="claude-opus-5-5"

rm -rf "$work"
mkdir -p "$work"
cp -a "$snapshot"/. "$work"/
prompt="$(cat "$here/checkpoint-prompt.md")"
if [[ "$arm" == C ]]; then
  cp -a "$fragments" "$work/discovered"
  prompt="$prompt"$'\n\n'"$(cat "$here/checkpoint-prompt-c.md")"
fi

session="$(uuidgen)"
(cd "$work" && claude -p "$prompt" --setting-sources project --strict-mcp-config --disable-slash-commands \
  --model "$model" --session-id "$session" --max-budget-usd 30 \
  --allowedTools "Read Glob Grep Write Bash(ls *) Bash(find *) Bash(grep *) Bash(cat *) Bash(head *) Bash(wc *)" \
  --disallowedTools "Agent WebFetch WebSearch" \
  --output-format json >"$work/result.json")

transcript="$(find "$HOME/.claude/projects" -name "$session.jsonl" | head -1)"
[[ -n "$transcript" ]] || { echo "no transcript for session $session" >&2; exit 1; }
mkdir -p "$here/transcripts"
# Tool results carry file contents, possibly Secret values, and the other record types
# carry the system prompt, request bodies and account data: keep only the prompt, the
# assistant messages with their usage and tool inputs, and the final cost.
jq -c 'select(.type == "assistant" or .type == "cost-state" or
  (.type == "user" and (tostring | contains("\"tool_result\"") | not)))' "$transcript" |
  grep -v -e '"type":"tool_result"' -e '"toolUseResult"' >"$here/transcripts/checkpoint-$arm.jsonl"
cp "$work/facts.yaml" "$here/checkpoint-$arm.facts.yaml"
echo "arm $arm: session $session, transcript transcripts/checkpoint-$arm.jsonl"
