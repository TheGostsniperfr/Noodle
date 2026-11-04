package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	frHeadRe     = regexp.MustCompile(`(?m)^- \*\*(FR-\d{3})\*\*`)
	adrRefRe     = regexp.MustCompile(`\bADR-(\d{4})\b`)
	cRefRe       = regexp.MustCompile(`\bC(\d{2})\b`)
	featureRefRe = regexp.MustCompile(`\*\*([A-Z][A-Za-z0-9]*-\d{2,})\*\*`)
	cDefRe       = regexp.MustCompile(`(?m)^\| \*\*(C\d{2})\*\*`)
	codeSpanRe   = regexp.MustCompile("`[^`\n]*`")
	fenceRe      = regexp.MustCompile("(?s)```.*?```")
	taskLineRe   = regexp.MustCompile(`(?m)^- \[[ x]\] \*\*T\d+\*\*.*(?:\n  .*)*`)
	sectionRe    = regexp.MustCompile(`(?m)^## (\S+)`)
	criterionRe  = regexp.MustCompile(`(?m)^\| \d+ \|`)
)

// CheckSpecs checks every spec, or only dirs. Rules for all specs: cited ADRs, Cnn
// invariants and flipped features exist. Rules for "Mode: full" specs (ADR-0024): every
// requirement is EARS with a scenario, named by a task, covered by a feature of the
// spec's phase, and every feature of that phase has a contract entry.
func (r *Repo) CheckSpecs(dirs []string) ([]Finding, error) {
	if len(dirs) == 0 {
		var err error
		if dirs, err = r.specDirs(); err != nil {
			return nil, err
		}
	}
	feats, err := r.features()
	if err != nil {
		return nil, err
	}
	known := map[string]bool{}
	for _, f := range feats.all() {
		known[f.ID] = true
	}
	invariants, err := r.invariants()
	if err != nil {
		return nil, err
	}
	var out []Finding
	for _, dir := range dirs {
		f, err := r.checkSpec(filepath.Clean(dir), feats, known, invariants)
		if err != nil {
			return nil, err
		}
		out = append(out, f...)
	}
	return out, nil
}

func (r *Repo) invariants() (map[string]bool, error) {
	text, err := r.read("docs", "CONSTITUTION.md")
	if err != nil {
		return nil, fmt.Errorf("read constitution: %w", err)
	}
	out := map[string]bool{}
	for _, m := range cDefRe.FindAllStringSubmatch(text, -1) {
		out[m[1]] = true
	}
	return out, nil
}

func (r *Repo) adrExists(num string) bool {
	m, _ := filepath.Glob(r.path("docs", "adr", num+"-*.md"))
	return len(m) > 0
}

func (r *Repo) checkSpec(dir string, feats *Features, known, invariants map[string]bool) ([]Finding, error) {
	var out []Finding
	docs := map[string]string{}
	for _, name := range []string{"spec.md", "plan.md", "tasks.md", "contract.md"} {
		text, err := r.read(dir, name)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", filepath.Join(dir, name), err)
		}
		docs[name] = normalize(text)
	}
	spec, ok := docs["spec.md"]
	if !ok {
		return []Finding{{dir, "no spec.md"}}, nil
	}
	for name, text := range docs {
		file := filepath.ToSlash(filepath.Join(dir, name))
		// Inline code quotes a message or an example; only prose cites.
		prose := codeSpanRe.ReplaceAllString(fenceRe.ReplaceAllString(text, ""), "")
		for _, num := range uniqueSorted(submatches(adrRefRe, prose)) {
			if !r.adrExists(num) {
				out = append(out, Finding{file, "cites ADR-" + num + ", which does not exist"})
			}
		}
		for _, n := range uniqueSorted(submatches(cRefRe, prose)) {
			if !invariants["C"+n] {
				out = append(out, Finding{file, "cites C" + n + ", which is not in docs/CONSTITUTION.md"})
			}
		}
	}
	tasks := docs["tasks.md"]
	taskLines := strings.Join(taskLineRe.FindAllString(tasks, -1), "\n")
	tasksFile := filepath.ToSlash(filepath.Join(dir, "tasks.md"))
	for _, id := range uniqueSorted(submatches(featureRefRe, tasks)) {
		if !strings.HasPrefix(id, "FR-") && !known[id] {
			out = append(out, Finding{tasksFile, "flips " + id + ", which is not in docs/features.json"})
		}
	}
	if specMode(spec) != "full" {
		return out, nil
	}

	specFile := filepath.ToSlash(filepath.Join(dir, "spec.md"))
	var phase *Phase
	for i := range feats.Phases {
		if filepath.Clean(filepath.FromSlash(feats.Phases[i].Spec)) == filepath.Join(dir, "spec.md") {
			phase = &feats.Phases[i]
		}
	}
	if phase == nil {
		return append(out, Finding{specFile, "no phase in docs/features.json points at this spec"}), nil
	}
	covered := map[string]bool{}
	for _, f := range phase.Features {
		for _, fr := range f.Covers {
			covered[fr] = true
		}
	}
	reqs := requirements(spec)
	if len(reqs) == 0 {
		out = append(out, Finding{specFile, "a full-mode spec needs requirements written as - **FR-NNN**"})
	}
	for _, req := range reqs {
		if !strings.Contains(req.text, "SHALL") {
			out = append(out, Finding{specFile, req.id + " is not EARS: no SHALL"})
		}
		scenario := ""
		if i := strings.Index(req.text, "Scenario"); i >= 0 {
			scenario = req.text[i:]
		}
		if !strings.Contains(scenario, "WHEN") || !strings.Contains(scenario, "THEN") {
			out = append(out, Finding{specFile, req.id + " has no Scenario with WHEN and THEN"})
		}
		if !strings.Contains(taskLines, req.id) {
			out = append(out, Finding{tasksFile, req.id + " is named by no task"})
		}
		if !covered[req.id] {
			out = append(out, Finding{"docs/features.json", req.id + " of " + specFile + " is covered by no feature of phase " + phase.ID})
		}
	}
	declared := map[string]bool{}
	for _, req := range reqs {
		declared[req.id] = true
	}
	for _, f := range phase.Features {
		if f.Verify == nil || strings.TrimSpace(*f.Verify) == "" {
			out = append(out, Finding{"docs/features.json", f.ID + " has no verify step"})
		}
		for _, fr := range f.Covers {
			if !declared[fr] {
				out = append(out, Finding{"docs/features.json", f.ID + " covers " + fr + ", which " + specFile + " does not declare"})
			}
		}
	}
	contract, hasContract := docs["contract.md"]
	contractFile := filepath.ToSlash(filepath.Join(dir, "contract.md"))
	if !hasContract {
		return append(out, Finding{contractFile, "a full-mode spec needs a contract"}), nil
	}
	criteria := contractCriteria(contract)
	for _, f := range phase.Features {
		if criteria[f.ID] == 0 {
			out = append(out, Finding{contractFile, "no criteria for " + f.ID})
		}
	}
	return out, nil
}

// contractCriteria counts the criterion rows ("| 1 | …") under each "## <id> …"
// section, so an id mentioned in prose elsewhere does not count as criteria.
func contractCriteria(contract string) map[string]int {
	out := map[string]int{}
	locs := sectionRe.FindAllStringSubmatchIndex(contract, -1)
	for i, loc := range locs {
		end := len(contract)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		id := contract[loc[2]:loc[3]]
		out[id] += len(criterionRe.FindAllString(contract[loc[1]:end], -1))
	}
	return out
}

type requirement struct{ id, text string }

// requirements splits a spec into its FR blocks: from one "- **FR-NNN**" line to the
// next one or the next heading.
func requirements(spec string) []requirement {
	locs := frHeadRe.FindAllStringSubmatchIndex(spec, -1)
	var out []requirement
	for i, loc := range locs {
		end := len(spec)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		block := spec[loc[0]:end]
		if h := strings.Index(block, "\n#"); h >= 0 {
			block = block[:h]
		}
		out = append(out, requirement{id: spec[loc[2]:loc[3]], text: block})
	}
	return out
}

func submatches(re *regexp.Regexp, s string) []string {
	var out []string
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}
