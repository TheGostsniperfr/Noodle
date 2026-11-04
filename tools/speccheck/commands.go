package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	targetRe     = regexp.MustCompile(`(?m)^([a-z][a-z0-9-]*):[^=\n]*## `)
	anyTargetRe  = regexp.MustCompile(`(?m)^([a-z][a-z0-9-]*):([^=\n]*)$`)
	commandRowRe = regexp.MustCompile("(?m)^\\| `make ([a-z][a-z0-9-]*)")
)

// CheckCommands keeps docs/COMMANDS.md and the documented Makefile targets in step, so
// an agent reading the reference never meets a target that does not exist, or misses one.
func (r *Repo) CheckCommands() ([]Finding, error) {
	mk, err := r.read("Makefile")
	if err != nil {
		return nil, fmt.Errorf("read Makefile: %w", err)
	}
	doc, err := r.read("docs", "COMMANDS.md")
	if err != nil {
		return nil, fmt.Errorf("read COMMANDS.md: %w", err)
	}
	targets := map[string]bool{}
	for _, t := range submatches(targetRe, normalize(mk)) {
		targets[t] = true
	}
	var out []Finding
	for _, m := range anyTargetRe.FindAllStringSubmatch(normalize(mk), -1) {
		if !strings.Contains(m[2], "## ") {
			out = append(out, Finding{"Makefile", "target " + m[1] + " has no ## description"})
		}
	}
	rows := map[string]bool{}
	for _, t := range submatches(commandRowRe, normalize(doc)) {
		rows[t] = true
	}
	for _, t := range uniqueSorted(keys(targets)) {
		if !rows[t] {
			out = append(out, Finding{"docs/COMMANDS.md", "no row for make " + t})
		}
	}
	for _, t := range uniqueSorted(keys(rows)) {
		if !targets[t] {
			out = append(out, Finding{"Makefile", "no documented target " + t + " (COMMANDS.md lists it; add it with a ## comment)"})
		}
	}
	return out, nil
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, strings.TrimSpace(k))
	}
	return out
}
