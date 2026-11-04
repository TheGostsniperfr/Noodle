package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Repo struct{ Root string }

type Feature struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Verify      *string  `json:"verify,omitempty"`
	Passes      bool     `json:"passes"`
	Covers      []string `json:"covers,omitempty"`
}

type Phase struct {
	ID       string    `json:"id"`
	Spec     string    `json:"spec"`
	Features []Feature `json:"features"`
}

type Features struct {
	Phases  []Phase   `json:"phases"`
	Backlog []Feature `json:"backlog"`
}

func (r *Repo) path(p ...string) string { return filepath.Join(append([]string{r.Root}, p...)...) }

func (r *Repo) read(p ...string) (string, error) {
	b, err := os.ReadFile(r.path(p...))
	return string(b), err
}

func parseFeatures(b []byte) (*Features, error) {
	var f Features
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("parse features.json: %w", err)
	}
	return &f, nil
}

func (r *Repo) features() (*Features, error) {
	b, err := os.ReadFile(r.path("docs", "features.json"))
	if err != nil {
		return nil, fmt.Errorf("read features.json: %w", err)
	}
	return parseFeatures(b)
}

// all lists phase features then backlog entries, the order check-features used.
func (f *Features) all() []Feature {
	var out []Feature
	for _, p := range f.Phases {
		out = append(out, p.Features...)
	}
	return append(out, f.Backlog...)
}

// phaseOf returns the phase that lists a feature, or nil for backlog entries.
func (f *Features) phaseOf(id string) *Phase {
	for i := range f.Phases {
		for _, ft := range f.Phases[i].Features {
			if ft.ID == id {
				return &f.Phases[i]
			}
		}
	}
	return nil
}

// specDirOf maps a phase to the directory of its spec, docs/evidence when it has none.
func specDirOf(p *Phase) string {
	if p == nil || p.Spec == "" {
		return filepath.Join("docs", "evidence")
	}
	return filepath.Dir(filepath.FromSlash(p.Spec))
}

func (r *Repo) Locate(id string) (string, error) {
	f, err := r.features()
	if err != nil {
		return "", err
	}
	for _, ft := range f.all() {
		if ft.ID == id {
			return specDirOf(f.phaseOf(id)), nil
		}
	}
	return "", fmt.Errorf("feature %s is not in docs/features.json", id)
}

var modeRe = regexp.MustCompile(`(?m)^- Mode: \**(full|quick)\**`)

// specMode is "full", "quick", or "" for specs written before ADR-0024.
func specMode(spec string) string {
	if m := modeRe.FindStringSubmatch(spec); m != nil {
		return m[1]
	}
	return ""
}

func (r *Repo) specDirs() ([]string, error) {
	matches, err := filepath.Glob(r.path("specs", "[0-9][0-9][0-9]-*"))
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, m := range matches {
		if st, err := os.Stat(m); err == nil && st.IsDir() {
			rel, _ := filepath.Rel(r.Root, m)
			dirs = append(dirs, rel)
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

func uniqueSorted(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out
}

func normalize(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
