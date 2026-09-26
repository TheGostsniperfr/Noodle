#!/usr/bin/env bash
# Guards docs/features.json against rewriting history: a feature may flip "passes", but
# never disappear or change its description or verify step. Supersede it with a new entry
# instead. Usage: check-features.sh [base-ref], default HEAD (i.e. uncommitted edits).
# Brian can override a deliberate change with ALLOW_FEATURE_EDIT=1.
set -euo pipefail
cd "$(dirname "$0")/.."
base="${1:-HEAD}"
[[ "${ALLOW_FEATURE_EDIT:-0}" == "1" ]] && exit 0
base_json="$(git show "$base:docs/features.json" 2>/dev/null)" || exit 0

violations="$(jq -r -n --argjson base "$base_json" --slurpfile cur docs/features.json '
  def items(x): [x.phases[].features[], (x.backlog // [])[]];
  (items($cur[0]) | map({key: .id, value: .}) | from_entries) as $c
  | items($base)[] as $b
  | $c[$b.id] as $n
  | if $n == null then "\($b.id): removed"
    elif $n.description != $b.description then "\($b.id): description changed"
    elif ($n.verify // null) != ($b.verify // null) then "\($b.id): verify step changed"
    else empty end')"

if [[ -n "$violations" ]]; then
  echo "docs/features.json rewrites existing features (base: $base):" >&2
  echo "$violations" | sed 's/^/  /' >&2
  exit 1
fi
