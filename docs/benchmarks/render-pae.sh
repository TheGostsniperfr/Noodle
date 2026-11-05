#!/usr/bin/env bash
# Renders the PAE charts the way its ApplicationSets do (release name, destination
# namespace, three value files), then runs discovery on the result.
#   docs/benchmarks/render-pae.sh WORKSPACE OUT_DIR
# WORKSPACE holds argocd/ and k8s-majoutes/ as checked out for the run.
set -euo pipefail

[[ $# -eq 2 ]] || { echo "usage: $0 WORKSPACE OUT_DIR" >&2; exit 2; }
ws="$(cd "$1" && pwd)"
out="$2"
root="$(cd "$(dirname "$0")/../.." && pwd)"
rendered="$(mktemp -d)"
trap 'rm -rf "$rendered"' EXIT

# render TENANT_DIR CHART_DIR NAMESPACE: one directory per Application, since objects
# that name no namespace land in the destination namespace, as Argo CD applies them.
render() {
  local tenant_dir="$1" dir="$2" ns="$3" app env args=()
  app="$(basename "$(dirname "$dir")")"
  env="$(basename "$dir")"
  # ignoreMissingValueFiles: true, so absent files are skipped as Argo CD does.
  for f in "$tenant_dir/values-$(basename "$tenant_dir").yaml" "$dir/../values-common.yaml" "$dir/values.yaml"; do
    [[ -f "$f" ]] && args+=(-f "$f")
  done
  if [[ -f "$dir/.argocd-override.yaml" ]]; then
    ns="$(sed -n 's/^destinationNamespace: *//p' "$dir/.argocd-override.yaml")"
  fi
  local target="$rendered/$ns/$env-$app"
  mkdir -p "$target"
  (cd "$dir" && { helm dependency build >/dev/null 2>&1 || true; } &&
   helm template "$app" . -n "$ns" "${args[@]}" >"$target/manifests.yaml") ||
    { echo "render failed: $dir" >&2; rm -rf "$target"; }
}

for chart in "$ws"/argocd/tenants/sre/*/*/Chart.yaml; do
  dir="$(dirname "$chart")"
  render "$ws/argocd/tenants/sre" "$dir" "$(basename "$(dirname "$dir")")"
done
for chart in "$ws"/k8s-majoutes/majoutes/*/*/Chart.yaml; do
  dir="$(dirname "$chart")"
  render "$ws/k8s-majoutes/majoutes" "$dir" "$(basename "$dir")-majoutes"
done

mkdir -p "$out"
out="$(cd "$out" && pwd)"
go build -C "$root" -o "$rendered/noodle" ./cmd/noodle
for target in "$rendered"/*/*/; do
  ns="$(basename "$(dirname "$target")")"
  name="$(basename "$target")"
  "$rendered/noodle" discover k8s "$target" -namespace "$ns" -o "$out/k8s-$ns-$name.yaml" -observed-at 2026-09-29T00:00:00Z
done
