package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skillFile(frontmatter, body string) string {
	return "---\n" + frontmatter + "\n---\n\n" + body
}

func validSkill(name string) string {
	return skillFile("name: "+name+"\ndescription: Draws things.\nlicense: Apache-2.0", "# Title\n\nUse it.\n")
}

func TestSkills_ChecksEachRule(t *testing.T) {
	t.Parallel()
	const path = "skills/demo/SKILL.md"
	tests := []struct {
		name    string
		path    string
		content string
		rule    string
	}{
		{"name_not_folder", path, validSkill("other"), "name-folder"},
		{"name_uppercase", "skills/Demo/SKILL.md", validSkill("Demo"), "name-format"},
		{"name_too_long", "skills/" + strings.Repeat("a", 65) + "/SKILL.md", validSkill(strings.Repeat("a", 65)), "name-length"},
		{"description_empty", path, skillFile("name: demo\ndescription: \"\"", "# T\n"), "description-empty"},
		{"description_too_long", path, skillFile("name: demo\ndescription: "+strings.Repeat("x", 1025), "# T\n"), "description-length"},
		{"unknown_key", path, skillFile("name: demo\ndescription: Draws.\nallowed-tools: Bash", "# T\n"), "frontmatter-key"},
		{"no_frontmatter", path, "# Just a title\n", "frontmatter"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := fixture(t, map[string]string{tt.path: tt.content})

			findings, err := repo.CheckSkills()

			require.NoError(t, err)
			require.Len(t, findings, 1)
			assert.Contains(t, findings[0].String(), tt.path)
			assert.Contains(t, findings[0].String(), tt.rule+":")
		})
	}

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		repo := fixture(t, map[string]string{path: validSkill("demo"), ".agents/skills/proc/SKILL.md": validSkill("proc")})

		findings, err := repo.CheckSkills()

		require.NoError(t, err)
		assert.Empty(t, findings)
	})

	t.Run("a_description_of_exactly_1024_characters_is_valid", func(t *testing.T) {
		t.Parallel()
		repo := fixture(t, map[string]string{path: skillFile("name: demo\ndescription: "+strings.Repeat("x", 1024), "# T\n")})

		findings, err := repo.CheckSkills()

		require.NoError(t, err)
		assert.Empty(t, findings)
	})
}

func TestSkills_KeepsClaudeCodeNamesUnderTheirHeading(t *testing.T) {
	t.Parallel()
	const path = "skills/demo/SKILL.md"
	for _, literal := range []string{"noodle:fact-finder", "noodle:diagram-reviewer", "noodle:icon-curator", "the Agent tool", "Task tool", "/plugin"} {
		t.Run("claude_name_outside_heading/"+literal, func(t *testing.T) {
			t.Parallel()
			repo := fixture(t, map[string]string{path: skillFile("name: demo\ndescription: Draws.", "# T\n\n## Workflow\n\nAsk "+literal+" first.\n")})

			findings, err := repo.CheckSkills()

			require.NoError(t, err)
			require.Len(t, findings, 1)
			assert.Contains(t, findings[0].String(), path)
			assert.Contains(t, findings[0].String(), "claude-only-name:")
			assert.Contains(t, findings[0].String(), "line 10")
		})
	}

	sections := []struct {
		name string
		body string
		want int
	}{
		{"claude_name_under_heading", "# T\n\n## In Claude Code\n\nUse the Agent tool.\n\n### Detail\n\nAlso /plugin.\n", 0},
		{"claude_name_after_the_section_ended", "# T\n\n## In Claude Code\n\nUse the Agent tool.\n\n## Next\n\nAgain the Agent tool.\n", 1},
		{"claude_name_in_a_code_comment_does_not_end_the_section", "## In Claude Code\n\n```bash\n# note\nrun the Agent tool\n```\n", 0},
	}
	for _, tt := range sections {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := fixture(t, map[string]string{path: skillFile("name: demo\ndescription: Draws.", tt.body)})

			findings, err := repo.CheckSkills()

			require.NoError(t, err)
			assert.Len(t, findings, tt.want)
		})
	}
}

func TestSkills_FailsAllForABrokenProcessSkill(t *testing.T) {
	t.Parallel()
	t.Run("process_skill_broken", func(t *testing.T) {
		t.Parallel()
		const path = ".agents/skills/broken/SKILL.md"
		repo := fixture(t, map[string]string{path: skillFile("name: broken\ndescription: "+strings.Repeat("x", 1025), "# T\n")})
		var stdout, stderr strings.Builder

		err := run([]string{"-root", repo.Root, "all"}, &stdout, &stderr)

		assert.ErrorIs(t, err, errFindings)
		lines := strings.Split(strings.TrimSpace(stderr.String()), "\n")
		require.Len(t, lines, 1)
		assert.Contains(t, lines[0], path)
	})
}
