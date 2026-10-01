#!/usr/bin/env bash
# Measures static discovery on real repositories: what an agent would read against the
# fragment it gets instead. Works on a copy; the repositories are never written.
#   scripts/discover-eval.sh [-m REGEX] REPO...
# -m keeps only the chart and kustomization directories whose path matches, e.g. /prod
set -euo pipefail

match='.'
if [[ "${1:-}" == "-m" ]]; then match="$2"; shift 2; fi
[[ $# -gt 0 ]] || { echo "usage: $0 [-m REGEX] REPO..." >&2; exit 2; }

root="$(cd "$(dirname "$0")/.." && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
go build -C "$root" -o "$work/noodle" ./cmd/noodle

for repo in "$@"; do
  name="$(basename "$repo")"
  echo "== $name"
  rsync -a --exclude .git "$repo/" "$work/src-$name/"
  out="$work/rendered-$name"
  mkdir -p "$out"

  echo "-- raw, as an agent would find it"
  # Relative paths, as a user would pass them: every src repeats its file path.
  (cd "$work" && ./noodle discover k8s "src-$name" -o /dev/null -observed-at 2026-01-01T00:00:00Z)

  ok=0 failed=0
  # Library charts under a charts/ directory are dependencies, not releases; Kustomize
  # bases are rendered through their overlays.
  while IFS= read -r marker; do
    dir="$(dirname "$marker")"
    rel="${dir#"$work/src-$name/"}"
    [[ "$rel" =~ (^|/)(charts|base|bases|components)(/|$) ]] && continue
    [[ "$rel" =~ $match ]] || continue
    file="$out/$(echo "$rel" | tr / -).yaml"
    if [[ "$(basename "$marker")" == Chart.yaml ]]; then
      (cd "$dir" && helm dependency build >/dev/null 2>&1; helm template "$(basename "$dir")" . >"$file" 2>/dev/null) && ok=$((ok + 1)) || { failed=$((failed + 1)); echo "   render failed: $rel"; rm -f "$file"; }
    else
      kustomize build --enable-helm --load-restrictor LoadRestrictionsNone "$dir" >"$file" 2>/dev/null && ok=$((ok + 1)) || { failed=$((failed + 1)); echo "   render failed: $rel"; rm -f "$file"; }
    fi
  done < <(find "$work/src-$name" \( -name Chart.yaml -o -name kustomization.yaml -o -name kustomization.yml \) | sort)

  echo "-- rendered: $ok ok, $failed failed"
  (cd "$work" && ./noodle discover k8s "rendered-$name" -o "$name.fragment.yaml" -observed-at 2026-01-01T00:00:00Z)
  echo "-- unresolved by kind"
  grep -o 'kind: [a-z-]*, value: [^,}]*' "$work/$name.fragment.yaml" | grep -v 'kind: \(backend\|group\)' | sed 's/^/   /' || true
done
