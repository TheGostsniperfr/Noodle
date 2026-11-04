package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	indexStart = "<!-- adr-index:start -->"
	indexEnd   = "<!-- adr-index:end -->"
)

var (
	adrFileRe    = regexp.MustCompile(`^(\d{4})-.+\.md$`)
	adrTitleRe   = regexp.MustCompile(`(?m)^# ADR-\d{4} · (.+)$`)
	adrStatusRe  = regexp.MustCompile(`(?m)^- Status: (.+)$`)
	adrStatuses  = map[string]bool{"proposed": true, "accepted": true, "rejected": true, "deprecated": true, "superseded": true}
	optionItemRe = regexp.MustCompile(`(?m)^(?:- |### |\d+\. )`)
)

// frontmatter of ADRs written with docs/adr/_template.md (ADR-0024). Older ADRs keep
// their "- Status:" line and are never rewritten.
type adrMeta struct {
	Status   string   `yaml:"status"`
	Date     string   `yaml:"date"`
	Deciders []string `yaml:"deciders"`
}

type adr struct {
	num, file, title, status string
}

func splitFrontmatter(text string) (string, string, bool) {
	if !strings.HasPrefix(text, "---\n") {
		return "", text, false
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return "", text, false
	}
	return text[4 : 4+end], text[4+end+5:], true
}

// section returns the body of the first heading whose text starts with name, up to the
// next heading of the same or a higher level.
func section(body, name string) (string, bool) {
	lines := strings.Split(body, "\n")
	for i, l := range lines {
		level := len(l) - len(strings.TrimLeft(l, "#"))
		if level == 0 || !strings.HasPrefix(strings.TrimSpace(l[level:]), name) {
			continue
		}
		var out []string
		for _, n := range lines[i+1:] {
			nl := len(n) - len(strings.TrimLeft(n, "#"))
			if nl > 0 && nl <= level && strings.HasPrefix(n[nl:], " ") {
				break
			}
			out = append(out, n)
		}
		return strings.TrimSpace(strings.Join(out, "\n")), true
	}
	return "", false
}

func (r *Repo) loadADRs() ([]adr, []Finding, error) {
	entries, err := os.ReadDir(r.path("docs", "adr"))
	if err != nil {
		return nil, nil, fmt.Errorf("read docs/adr: %w", err)
	}
	var out []adr
	var findings []Finding
	for _, e := range entries {
		m := adrFileRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		file := "docs/adr/" + e.Name()
		raw, err := r.read("docs", "adr", e.Name())
		if err != nil {
			return nil, nil, err
		}
		text := normalize(raw)
		a := adr{num: m[1], file: e.Name()}
		fm, body, hasFM := splitFrontmatter(text)
		if t := adrTitleRe.FindStringSubmatch(body); t != nil {
			a.title = strings.TrimSpace(t[1])
		} else {
			findings = append(findings, Finding{file, "no '# ADR-" + a.num + " · Title' heading"})
		}
		if !hasFM {
			if s := adrStatusRe.FindStringSubmatch(body); s != nil {
				a.status = strings.TrimSpace(s[1])
			} else {
				findings = append(findings, Finding{file, "no status"})
			}
			out = append(out, a)
			continue
		}
		var meta adrMeta
		if err := yaml.Unmarshal([]byte(fm), &meta); err != nil {
			findings = append(findings, Finding{file, "frontmatter: " + err.Error()})
		}
		if !adrStatuses[meta.Status] {
			findings = append(findings, Finding{file, fmt.Sprintf("status %q is not one of proposed, accepted, rejected, deprecated, superseded", meta.Status)})
		}
		if meta.Date == "" || len(meta.Deciders) == 0 {
			findings = append(findings, Finding{file, "frontmatter needs date and deciders"})
		}
		a.status = strings.ToUpper(meta.Status[:min(1, len(meta.Status))]) + meta.Status[min(1, len(meta.Status)):]
		if opts, ok := section(body, "Considered options"); !ok || len(optionItemRe.FindAllString(opts, -1)) < 2 {
			findings = append(findings, Finding{file, "needs at least two considered options"})
		}
		if conf, ok := section(body, "Confirmation"); !ok || conf == "" {
			findings = append(findings, Finding{file, "needs a Confirmation section: what proves the decision is kept"})
		}
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].num < out[j].num })
	return out, findings, nil
}

func renderIndex(adrs []adr) string {
	var b strings.Builder
	b.WriteString(indexStart + "\n| ADR | Title | Status |\n|---|---|---|\n")
	for _, a := range adrs {
		fmt.Fprintf(&b, "| [%s](%s) | %s | %s |\n", a.num, a.file, a.title, a.status)
	}
	b.WriteString(indexEnd)
	return b.String()
}

// CheckADRs validates every ADR and compares the generated index with the one in
// docs/adr/README.md; with write it replaces the README's index instead.
func (r *Repo) CheckADRs(write bool) ([]Finding, error) {
	adrs, findings, err := r.loadADRs()
	if err != nil {
		return nil, err
	}
	readmePath := r.path("docs", "adr", "README.md")
	raw, err := os.ReadFile(readmePath)
	if err != nil {
		return nil, fmt.Errorf("read ADR index: %w", err)
	}
	readme := normalize(string(raw))
	start, end := strings.Index(readme, indexStart), strings.Index(readme, indexEnd)
	if start < 0 || end < start {
		return append(findings, Finding{"docs/adr/README.md", "no " + indexStart + " … " + indexEnd + " block"}), nil
	}
	want := renderIndex(adrs)
	current := readme[start : end+len(indexEnd)]
	if current == want {
		return findings, nil
	}
	if write {
		updated := readme[:start] + want + readme[end+len(indexEnd):]
		if err := os.WriteFile(readmePath, []byte(updated), 0o644); err != nil {
			return nil, fmt.Errorf("write ADR index: %w", err)
		}
		return findings, nil
	}
	return append(findings, Finding{filepath.ToSlash("docs/adr/README.md"), "index is stale: run make adr-index"}), nil
}
