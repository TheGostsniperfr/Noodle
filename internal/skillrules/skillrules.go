// Package skillrules checks a SKILL.md against the Agent Skills format that Claude Code,
// OpenCode and Codex all read, and keeps Claude Code-only names out of the shared text.
package skillrules

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const (
	RuleFrontmatter      = "frontmatter"
	RuleNameFolder       = "name-folder"
	RuleNameFormat       = "name-format"
	RuleNameLength       = "name-length"
	RuleDescriptionEmpty = "description-empty"
	RuleDescriptionLen   = "description-length"
	RuleFrontmatterKey   = "frontmatter-key"
	RuleClaudeOnlyName   = "claude-only-name"

	maxNameLength        = 64
	maxDescriptionLength = 1024
	claudeCodeHeading    = "In Claude Code"
)

// ClaudeOnlyNames are tools and commands that exist in Claude Code and nowhere else.
var ClaudeOnlyNames = []string{
	"noodle:fact-finder",
	"noodle:diagram-reviewer",
	"noodle:icon-curator",
	"the Agent tool",
	"Task tool",
	"/plugin",
}

var (
	allowedKeys = map[string]bool{"name": true, "description": true, "license": true, "metadata": true}
	nameRe      = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	headingRe   = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
)

// Violation is one broken rule, printed as "<rule>: <detail>".
type Violation struct {
	Rule   string
	Detail string
}

func (v Violation) String() string { return v.Rule + ": " + v.Detail }

// Check validates the content of the SKILL.md found in the folder named folder.
func Check(folder, content string) []Violation {
	front, body, bodyStart, ok := splitFrontmatter(content)
	if !ok {
		return []Violation{{RuleFrontmatter, "no YAML frontmatter between --- lines"}}
	}
	var fields map[string]any
	if err := yaml.Unmarshal([]byte(front), &fields); err != nil {
		return []Violation{{RuleFrontmatter, "frontmatter is not valid YAML: " + err.Error()}}
	}

	var out []Violation
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !allowedKeys[k] {
			out = append(out, Violation{RuleFrontmatterKey, fmt.Sprintf("key %q is not allowed (name, description, license, metadata)", k)})
		}
	}

	name, _ := fields["name"].(string)
	switch {
	case name == "":
		out = append(out, Violation{RuleNameFormat, "name is missing"})
	default:
		if name != folder {
			out = append(out, Violation{RuleNameFolder, fmt.Sprintf("name %q differs from its folder %q", name, folder)})
		}
		if !nameRe.MatchString(name) {
			out = append(out, Violation{RuleNameFormat, fmt.Sprintf("name %q must be lowercase letters, digits and single hyphens", name)})
		}
		if n := utf8.RuneCountInString(name); n > maxNameLength {
			out = append(out, Violation{RuleNameLength, fmt.Sprintf("name has %d characters, the limit is %d", n, maxNameLength)})
		}
	}

	description, _ := fields["description"].(string)
	if n := utf8.RuneCountInString(strings.TrimSpace(description)); n == 0 {
		out = append(out, Violation{RuleDescriptionEmpty, "description is missing or empty"})
	} else if n > maxDescriptionLength {
		out = append(out, Violation{RuleDescriptionLen, fmt.Sprintf("description has %d characters, the limit is %d", n, maxDescriptionLength)})
	}

	return append(out, claudeOnlyNames(body, bodyStart)...)
}

func splitFrontmatter(content string) (front, body string, bodyStartLine int, ok bool) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	rest, found := strings.CutPrefix(content, "---\n")
	if !found {
		return "", "", 0, false
	}
	front, body, found = strings.Cut(rest, "\n---")
	if !found {
		return "", "", 0, false
	}
	_, body, _ = strings.Cut(body, "\n")
	return front, body, strings.Count(content, "\n") - strings.Count(body, "\n"), true
}

// claudeOnlyNames reports every Claude Code-only name that sits outside a section whose
// heading starts with "In Claude Code". The section ends at the next heading of the
// same or a higher level.
func claudeOnlyNames(body string, linesBefore int) []Violation {
	var out []Violation
	inFence := false
	sectionLevel := 0
	for i, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
		} else if m := headingRe.FindStringSubmatch(line); !inFence && m != nil {
			level := len(m[1])
			switch {
			case strings.HasPrefix(m[2], claudeCodeHeading):
				sectionLevel = level
			case sectionLevel > 0 && level <= sectionLevel:
				sectionLevel = 0
			}
		}
		if sectionLevel > 0 {
			continue
		}
		for _, name := range ClaudeOnlyNames {
			if strings.Contains(line, name) {
				out = append(out, Violation{RuleClaudeOnlyName, fmt.Sprintf("%q on line %d is outside a section titled %q", name, linesBefore+i+1, claudeCodeHeading)})
			}
		}
	}
	return out
}
