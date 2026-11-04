# The one entry point for working on noodle (C14, ADR-0024). Every target with a "##"
# comment is documented in docs/COMMANDS.md, and make spec-check fails if they differ.

SHELL := bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := help

NOODLE := go run ./cmd/noodle
SPECCHECK := go run ./tools/speccheck
BASE ?= HEAD
REGRESSION_ICONS := -icons examples/platform-regression/icons

# Read-only tools for the evaluator; write tools are refused outright. The tool flags take
# lists, so the prompt goes on stdin, never as a trailing argument.
EVALUATOR := claude -p --agent evaluator \
	--disallowedTools "Edit Write NotebookEdit" \
	--allowedTools "Read Glob Grep \
	Bash(go test:*) Bash(go vet:*) \
	Bash(go run ./tools/speccheck all) Bash(go run ./tools/speccheck spec:*) Bash(go run ./tools/speccheck adr) \
	Bash(go run ./tools/speccheck commands) Bash(go run ./tools/speccheck features:*) Bash(go run ./tools/speccheck locate:*) \
	Bash(git diff:*) Bash(git log:*) Bash(git show:*) Bash(git status:*) Bash(git grep:*) Bash(git rev-parse:*) \
	Bash(grep:*) Bash(ls:*) Bash(cat:*) Bash(head:*) Bash(tail:*) Bash(wc:*) Bash(cut:*) Bash(jq:*) \
	Bash(make check) Bash(make init) Bash(make spec-check) Bash(make features) Bash(make features BASE=origin/main) \
	Bash(make examples) Bash(make test) Bash(make vet) Bash(make help) Bash(make -s help) \
	Bash(make evaluate FEATURE=) Bash(make review-contract SPEC=)"

.PHONY: help init check ci vet test examples render features spec-check adr-index \
	evaluate review-contract discover-eval plugin vendor-check next

help: ## List the targets
	@grep -E '^[a-z][a-z0-9-]*:[^=]*## ' $(MAKEFILE_LIST) | sed -E 's/:[^#]*## /\t/' | expand -t 18

init: check plugin next ## Session smoke test: run first, fix anything red before new work

check: vet test examples features spec-check ## What every task ends with: vet, tests, examples, features, spec-check

ci: vet test examples features spec-check vendor-check ## What CI runs, reproducible locally (BASE=<sha>)

vet: ## go vet on every package
	go vet ./...

test: ## go test on every package
	go test ./...

examples: ## Lint every example, regression fixture and docs diagram (no output written)
	@echo "== examples"
	$(NOODLE) render examples/cnp-runtime --view runtime
	$(NOODLE) render examples/cnp-runtime --view oidc-login
	$(NOODLE) render examples/sequence-basics
	$(NOODLE) render examples/landscape --view stack
	$(NOODLE) render examples/catalog --view catalog
	$(NOODLE) render examples/cnp-access --view matrix
	$(NOODLE) render examples/messaging --view broker
	$(NOODLE) render examples/regression-78
	for v in proposal reference focus; do $(NOODLE) render examples/topology-diff --view $$v $(REGRESSION_ICONS); done
	for d in platform-overview app-runtime-view repo-map; do $(NOODLE) render examples/platform-regression/$$d $(REGRESSION_ICONS); done
	$(NOODLE) render docs/diagrams/contract-pipeline -icons docs/diagrams/icons

render: ## Lint and render one view: SYS=<dir> [VIEW=<id>] [OUT=<file.drawio>] [THEME=dark|light] [ICONS=<dirs>] [SLIDE=1]
	@test -n "$(SYS)" || { echo "usage: make render SYS=<dir> [VIEW=<id>] [OUT=<file.drawio>]" >&2; exit 2; }
	$(NOODLE) render $(SYS) $(if $(VIEW),--view $(VIEW)) $(if $(OUT),-o $(OUT)) $(if $(THEME),-theme $(THEME)) $(if $(ICONS),-icons $(ICONS)) $(if $(SLIDE),-slide)

features: ## features.json keeps its history; every new pass has its evidence (BASE=<ref>, default HEAD)
	@echo "== features (base $(BASE))"
	$(SPECCHECK) features -base $(BASE)

spec-check: ## Traceability of specs, ADR shape and index, COMMANDS.md against this Makefile
	@echo "== spec-check"
	$(SPECCHECK) all

adr-index: ## Regenerate the ADR index in docs/adr/README.md
	$(SPECCHECK) adr -write

evaluate: ## Grade one feature with the read-only evaluator; writes <spec>/evidence/<id>.md (FEATURE=<id>)
	@test -n "$(FEATURE)" || { echo "usage: make evaluate FEATURE=<id>" >&2; exit 2; }
	@dir="$$($(SPECCHECK) locate $(FEATURE))"; mkdir -p "$$dir/evidence"; out="$$dir/evidence/$(FEATURE).md"; tmp="$$(mktemp)"; \
	echo "Evaluate feature $(FEATURE). Its spec, contract and tasks are in $$dir (spec.md, contract.md, tasks.md). Follow your instructions and print only the evidence file." | $(EVALUATOR) > "$$tmp"; \
	if grep -q '^- Verdict:' "$$tmp"; then mv "$$tmp" "$$out"; grep -m1 '^- Verdict:' "$$out"; \
	else echo "evaluator output has no Verdict line, kept in $$tmp" >&2; exit 1; fi

review-contract: ## Review a spec's contract before implementation; writes <spec>/contract-review.md (SPEC=<NNN>)
	@test -n "$(SPEC)" || { echo "usage: make review-contract SPEC=<NNN>" >&2; exit 2; }
	@dirs=(specs/$(SPEC)-*/); [[ $${#dirs[@]} -eq 1 && -d "$${dirs[0]}" ]] || { echo "SPEC=$(SPEC) must match exactly one specs/$(SPEC)-* directory" >&2; exit 2; }; \
	dir="$${dirs[0]%/}"; out="$$dir/contract-review.md"; tmp="$$(mktemp)"; \
	echo "Review the contract of $$dir before implementation. Follow your contract-review instructions and print only the review." | $(EVALUATOR) > "$$tmp"; \
	if grep -q '^- Review:' "$$tmp"; then mv "$$tmp" "$$out"; grep -m1 '^- Review:' "$$out"; \
	else echo "review has no '- Review:' line, kept in $$tmp" >&2; exit 1; fi

discover-eval: ## Measure static discovery on real repositories, never written (REPOS="<dir>…" [MATCH=<regex>])
	@test -n "$(REPOS)" || { echo "usage: make discover-eval REPOS=\"<dir>…\" [MATCH=<regex>]" >&2; exit 2; }
	scripts/discover-eval.sh $(if $(MATCH),-m '$(MATCH)') $(REPOS)

plugin: ## Validate the Claude Code plugin manifest (skipped without the claude CLI)
	@echo "== plugin manifest"
	@if command -v claude >/dev/null; then claude plugin validate . | tail -1; else echo "claude CLI absent: skipped"; fi

vendor-check: ## vendor/ matches go.mod
	go mod vendor && git diff --exit-code -- vendor

next: ## The next five features to make pass
	@echo "== next features (docs/features.json)"
	@jq -r '.phases[] | .features[] | select(.passes == false) | "  \(.id)  \(.description)"' docs/features.json | head -5
