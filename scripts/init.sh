#!/usr/bin/env bash
# Session smoke test. Run it first in every session and fix anything red before new work.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "== go vet / test"
go vet ./...
go test ./...

echo "== lint reference example"
go run ./cmd/noodle examples/cnp-runtime/runtime-request-path.yaml
go run ./cmd/noodle render examples/cnp-runtime --view runtime
for f in examples/platform-regression/*.yaml; do go run ./cmd/noodle -icons examples/platform-regression/icons "$f"; done
for f in docs/diagrams/specs/*.yaml; do go run ./cmd/noodle -icons docs/diagrams/icons "$f"; done

echo "== features.json history"
./scripts/check-features.sh

echo "== plugin manifest"
if command -v claude >/dev/null; then
  claude plugin validate . | tail -1
else
  echo "claude CLI absent: skipped"
fi

echo "== next features (docs/features.json)"
jq -r '.phases[] | .features[] | select(.passes == false) | "  \(.id)  \(.description)"' docs/features.json | head -5
