package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	verdictRe   = regexp.MustCompile(`(?m)^- Verdict: \**([A-Z_]+)\**`)
	evaluatorRe = regexp.MustCompile(`(?m)^- Evaluator: (.+)$`)
)

// CheckFeatures guards docs/features.json against rewriting history (an entry never
// disappears, nor changes its description or verify step) and gates every feature that
// passes now but did not at base: it needs specs/<spec>/evidence/<id>.md with
// "Verdict: PASS", from the evaluator agent when its spec is in full mode (ADR-0024).
// allowEdit skips the history rule only, for a deliberate change Brian made.
func (r *Repo) CheckFeatures(base string, allowEdit bool) ([]Finding, error) {
	cur, err := r.features()
	if err != nil {
		return nil, err
	}
	baseFeatures, err := r.baseFeatures(base)
	if err != nil {
		return nil, err
	}
	var out []Finding
	before := map[string]Feature{}
	if baseFeatures != nil {
		for _, f := range baseFeatures.all() {
			before[f.ID] = f
		}
	}
	now := map[string]Feature{}
	for _, f := range cur.all() {
		now[f.ID] = f
	}
	if baseFeatures != nil && !allowEdit {
		for _, b := range baseFeatures.all() {
			n, ok := now[b.ID]
			switch {
			case !ok:
				out = append(out, Finding{"docs/features.json", b.ID + ": removed"})
			case n.Description != b.Description:
				out = append(out, Finding{"docs/features.json", b.ID + ": description changed"})
			case deref(n.Verify) != deref(b.Verify):
				out = append(out, Finding{"docs/features.json", b.ID + ": verify step changed"})
			}
		}
	}
	for _, f := range cur.all() {
		if !f.Passes || before[f.ID].Passes {
			continue
		}
		if msg := r.evidenceProblem(cur, f.ID); msg != "" {
			out = append(out, Finding{"docs/features.json", f.ID + " passes without evidence: " + msg})
		}
	}
	return out, nil
}

func (r *Repo) evidenceProblem(feats *Features, id string) string {
	dir := specDirOf(feats.phaseOf(id))
	file := filepath.ToSlash(filepath.Join(dir, "evidence", id+".md"))
	raw, err := r.read(dir, "evidence", id+".md")
	if os.IsNotExist(err) {
		return "no " + file + " (make evaluate FEATURE=" + id + ")"
	}
	if err != nil {
		return err.Error()
	}
	text := normalize(raw)
	v := verdictRe.FindStringSubmatch(text)
	if v == nil || v[1] != "PASS" {
		return file + " has no '- Verdict: PASS' line"
	}
	spec, _ := r.read(dir, "spec.md")
	if specMode(spec) == "full" {
		e := evaluatorRe.FindStringSubmatch(text)
		if e == nil || !strings.HasPrefix(strings.TrimSpace(e[1]), "evaluator") {
			return file + " must come from the evaluator agent: the spec is in full mode"
		}
	}
	return ""
}

func (r *Repo) baseFeatures(base string) (*Features, error) {
	cmd := exec.Command("git", "show", base+":docs/features.json")
	cmd.Dir = r.Root
	b, err := cmd.Output()
	if err != nil {
		// No such ref or no file at that ref: nothing to compare against, as before.
		return nil, nil
	}
	f, err := parseFeatures(b)
	if err != nil {
		return nil, fmt.Errorf("features.json at %s: %w", base, err)
	}
	return f, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
