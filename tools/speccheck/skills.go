package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheGostsniperfr/Noodle/internal/skillrules"
)

// skillGlobs are the skills that ship with or work on noodle: the plugin skills and the
// process skills stored once under .agents/skills (spec 010, FR-004 and FR-015).
var skillGlobs = []string{
	filepath.Join("skills", "*", "SKILL.md"),
	filepath.Join(".agents", "skills", "*", "SKILL.md"),
}

// CheckSkills validates every skill against the Agent Skills rules so the same file loads
// in Claude Code, OpenCode and Codex.
func (r *Repo) CheckSkills() ([]Finding, error) {
	var out []Finding
	for _, glob := range skillGlobs {
		matches, err := filepath.Glob(r.path(glob))
		if err != nil {
			return nil, fmt.Errorf("list skills: %w", err)
		}
		for _, m := range matches {
			content, err := os.ReadFile(m)
			if err != nil {
				return nil, fmt.Errorf("read skill: %w", err)
			}
			rel, _ := filepath.Rel(r.Root, m)
			for _, v := range skillrules.Check(filepath.Base(filepath.Dir(m)), string(content)) {
				out = append(out, Finding{filepath.ToSlash(rel), v.String()})
			}
		}
	}
	return out, nil
}
