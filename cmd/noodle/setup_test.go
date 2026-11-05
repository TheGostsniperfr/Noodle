package main

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TheGostsniperfr/Noodle/internal/skillrules"
)

var (
	agentNames   = []string{"diagram-reviewer", "fact-finder", "icon-curator"}
	openCodeWant = []string{
		".agents/skills/noodle-diagram/SKILL.md",
		".agents/skills/noodle-diagram/reference.md",
		".opencode/agents/diagram-reviewer.md",
		".opencode/agents/fact-finder.md",
		".opencode/agents/icon-curator.md",
	}
	codexWant = []string{
		".agents/skills/noodle-diagram/SKILL.md",
		".agents/skills/noodle-diagram/reference.md",
		".codex/agents/diagram-reviewer.toml",
		".codex/agents/fact-finder.toml",
		".codex/agents/icon-curator.toml",
	}
)

func wroteLines(paths []string) string {
	var b strings.Builder
	for _, p := range paths {
		b.WriteString("wrote " + p + "\n")
	}
	return b.String()
}

func testEnv(stdout, stderr *bytes.Buffer, home string) setupEnv {
	return setupEnv{stdout: stdout, stderr: stderr, home: home, version: "dev"}
}

func setupProject(t *testing.T, hostName string) (dir, stdout string) {
	t.Helper()
	dir = t.TempDir()
	var out, errOut bytes.Buffer

	err := runSetup([]string{hostName, "-project", dir}, testEnv(&out, &errOut, ""))

	require.NoError(t, err)
	require.Empty(t, errOut.String())
	return dir, out.String()
}

func treeFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	require.NoError(t, filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := os.ReadFile(p)
		rel, _ := filepath.Rel(root, p)
		files[filepath.ToSlash(rel)] = string(raw)
		return err
	}))
	return files
}

func assertMatchesGolden(t *testing.T, hostName, written string) {
	t.Helper()
	golden := filepath.Join("testdata", "setup", hostName)
	if *update {
		require.NoError(t, os.RemoveAll(golden))
		for rel, content := range treeFiles(t, written) {
			target := filepath.Join(golden, filepath.FromSlash(rel))
			require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o755))
			require.NoError(t, os.WriteFile(target, []byte(content), 0o644))
		}
	}
	assert.Equal(t, treeFiles(t, golden), treeFiles(t, written))
}

func assertCarriesAgentInstructions(t *testing.T, written, fileExt string) {
	t.Helper()
	for _, name := range agentNames {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("..", "..", "agents", name+".md"))
			require.NoError(t, err)
			_, body, ok := skillrules.SplitFrontmatter(string(source))
			require.True(t, ok)
			require.NotEmpty(t, strings.TrimSpace(body))

			for rel, content := range treeFiles(t, written) {
				if strings.HasSuffix(rel, "/"+name+fileExt) {
					assert.Contains(t, content, body)
					return
				}
			}
			t.Fatalf("no written file for agent %s", name)
		})
	}
}

func TestSetup_OpenCode(t *testing.T) {
	t.Run("project", func(t *testing.T) {
		dir, stdout := setupProject(t, "opencode")

		assert.Equal(t, wroteLines(openCodeWant), stdout)
		assertMatchesGolden(t, "opencode", dir)
	})

	t.Run("instructions", func(t *testing.T) {
		dir, _ := setupProject(t, "opencode")

		assertCarriesAgentInstructions(t, dir, ".md")
	})

	t.Run("skill_rules", func(t *testing.T) {
		dir, _ := setupProject(t, "opencode")
		skill, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", "noodle-diagram", "SKILL.md"))
		require.NoError(t, err)

		assert.Empty(t, skillrules.Check("noodle-diagram", string(skill)))
	})
}

func TestSetup_Codex(t *testing.T) {
	t.Run("project", func(t *testing.T) {
		dir, stdout := setupProject(t, "codex")

		assert.Equal(t, wroteLines(codexWant), stdout)
		assertMatchesGolden(t, "codex", dir)
	})

	t.Run("instructions", func(t *testing.T) {
		dir, _ := setupProject(t, "codex")

		assertCarriesAgentInstructions(t, dir, ".toml")
	})
}

func TestSetup_User(t *testing.T) {
	tests := []struct {
		host string
		want []string
	}{
		{"opencode", []string{".agents/skills/noodle-diagram/SKILL.md", ".config/opencode/agents/fact-finder.md"}},
		{"codex", []string{".agents/skills/noodle-diagram/SKILL.md", ".codex/agents/fact-finder.toml"}},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			var out, errOut bytes.Buffer

			err := setupCommand([]string{tt.host, "-user"}, &out, &errOut)

			require.NoError(t, err)
			for _, rel := range tt.want {
				assert.FileExists(t, filepath.Join(home, filepath.FromSlash(rel)))
				assert.Contains(t, out.String(), "wrote ~/"+rel+"\n")
			}
			assert.Len(t, treeFiles(t, home), 5)
		})
	}
}

func TestSetup_Idempotent(t *testing.T) {
	for _, hostName := range []string{"opencode", "codex"} {
		t.Run(hostName, func(t *testing.T) {
			t.Parallel()
			dir, _ := setupProject(t, hostName)
			longAgo := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			for rel := range treeFiles(t, dir) {
				require.NoError(t, os.Chtimes(filepath.Join(dir, filepath.FromSlash(rel)), longAgo, longAgo))
			}
			before := treeFiles(t, dir)
			var out, errOut bytes.Buffer

			err := runSetup([]string{hostName, "-project", dir}, testEnv(&out, &errOut, ""))

			require.NoError(t, err)
			assert.Empty(t, out.String())
			assert.Equal(t, before, treeFiles(t, dir))
			for rel := range before {
				info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel)))
				require.NoError(t, err)
				assert.True(t, info.ModTime().Equal(longAgo), "%s was rewritten", rel)
			}
		})
	}
}

func TestSetup_ReadsFlagsAndDirInAnyOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args func(dir string) []string
	}{
		{"flag then dir", func(dir string) []string { return []string{"opencode", "-project", dir} }},
		{"dir then flag", func(dir string) []string { return []string{"opencode", dir, "-project"} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			var out, errOut bytes.Buffer

			err := runSetup(tt.args(dir), testEnv(&out, &errOut, ""))

			require.NoError(t, err)
			assert.Equal(t, wroteLines(openCodeWant), out.String())
			assert.FileExists(t, filepath.Join(dir, ".opencode", "agents", "fact-finder.md"))
		})
	}
}

func TestSetup_RejectsBadArguments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no host", nil, "usage: noodle setup"},
		{"unknown host", []string{"emacs", "-project"}, `unknown host "emacs"`},
		{"neither scope", []string{"codex"}, "exactly one of -project and -user"},
		{"both scopes", []string{"codex", "-project", "-user"}, "exactly one of -project and -user"},
		{"user with a directory", []string{"codex", "-user", "somewhere"}, "-user takes none"},
		{"two directories", []string{"codex", "-project", "a", "b"}, "at most one DIR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out, errOut bytes.Buffer

			err := runSetup(tt.args, testEnv(&out, &errOut, t.TempDir()))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
			assert.Empty(t, out.String())
		})
	}
}

func TestSetup_KeepsForeign(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		host    string
		foreign string
		wrote   []string
	}{
		{"an OpenCode agent file", "opencode", ".opencode/agents/fact-finder.md", []string{
			".agents/skills/noodle-diagram/SKILL.md", ".agents/skills/noodle-diagram/reference.md",
			".opencode/agents/diagram-reviewer.md", ".opencode/agents/icon-curator.md"}},
		{"a Codex agent file", "codex", ".codex/agents/fact-finder.toml", []string{
			".agents/skills/noodle-diagram/SKILL.md", ".agents/skills/noodle-diagram/reference.md",
			".codex/agents/diagram-reviewer.toml", ".codex/agents/icon-curator.toml"}},
		{"the shared skill", "codex", ".agents/skills/noodle-diagram/SKILL.md", []string{
			".agents/skills/noodle-diagram/reference.md",
			".codex/agents/diagram-reviewer.toml", ".codex/agents/fact-finder.toml", ".codex/agents/icon-curator.toml"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			target := filepath.Join(dir, filepath.FromSlash(tt.foreign))
			require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o755))
			require.NoError(t, os.WriteFile(target, []byte("my own text\n"), 0o644))
			var out, errOut bytes.Buffer

			err := runSetup([]string{tt.host, "-project", dir}, testEnv(&out, &errOut, ""))

			require.Error(t, err)
			raw, readErr := os.ReadFile(target)
			require.NoError(t, readErr)
			assert.Equal(t, "my own text\n", string(raw))
			assert.Equal(t, "kept "+tt.foreign+": it was not written by noodle setup\n", errOut.String())
			assert.Equal(t, wroteLines(tt.wrote), out.String())
		})
	}
}

func TestSetup_ExitsOneAndKeepsAForeignFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, ".opencode", "agents", "fact-finder.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o755))
	require.NoError(t, os.WriteFile(target, []byte("my own text\n"), 0o644))

	cmd := exec.Command(buildNoodle(t, ""), "setup", "opencode", "-project", dir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitCode())
	assert.Contains(t, stderr.String(), ".opencode/agents/fact-finder.md")
	raw, readErr := os.ReadFile(target)
	require.NoError(t, readErr)
	assert.Equal(t, "my own text\n", string(raw))
}

func TestSetup_ReplacesAFileItWroteWhenTheVersionChanges(t *testing.T) {
	t.Parallel()
	for _, hostName := range []string{"opencode", "codex"} {
		t.Run(hostName, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			var out, errOut bytes.Buffer
			require.NoError(t, runSetup([]string{hostName, "-project", dir}, setupEnv{stdout: &out, stderr: &errOut, version: "v1.0.0"}))
			out.Reset()

			err := runSetup([]string{hostName, "-project", dir}, setupEnv{stdout: &out, stderr: &errOut, version: "v1.1.0"})

			require.NoError(t, err)
			assert.Equal(t, 5, strings.Count(out.String(), "wrote "))
			for rel, content := range treeFiles(t, dir) {
				assert.Contains(t, content, "noodle setup", rel)
				assert.NotContains(t, content, "v1.0.0", rel)
				assert.Contains(t, content, "v1.1.0", rel)
			}
		})
	}
}

func TestSetup_MarkerDoesNotBreakTheSkillRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"with frontmatter", "---\nname: x\n---\n\n# Body\n", "---\nname: x\n---\n<!-- noodle setup skills dev, do not edit: rerun noodle setup -->\n\n# Body\n"},
		{"without frontmatter", "# Reference\n", "<!-- noodle setup skills dev, do not edit: rerun noodle setup -->\n\n# Reference\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := markMarkdown(tt.content, markerText("skills", "dev"))

			assert.Equal(t, tt.want, got)
			assert.True(t, markerRe.MatchString(got))
		})
	}
}

func TestCodexAgent_EscapesWhatTOMLTreatsAsSyntax(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		agent agent
		want  string
	}{
		{
			"plain",
			agent{name: "scout", description: "Finds things", body: "\nLook.\n"},
			"# marker\nname = \"scout\"\ndescription = \"Finds things\"\ndeveloper_instructions = \"\"\"\n\nLook.\n\"\"\"\n",
		},
		{
			"quotes, backslashes and newlines in the description",
			agent{name: "scout", description: "say \"hi\" \\ now\nthen", body: "x\n"},
			"# marker\nname = \"scout\"\ndescription = \"say \\\"hi\\\" \\\\ now\\nthen\"\ndeveloper_instructions = \"\"\"\nx\n\"\"\"\n",
		},
		{
			"backslashes and triple quotes in the instructions",
			agent{name: "scout", description: "d", body: "a \\n b \"\"\" c\n"},
			"# marker\nname = \"scout\"\ndescription = \"d\"\ndeveloper_instructions = \"\"\"\na \\\\n b \\\"\\\"\\\" c\n\"\"\"\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			file, content := codexAgent(tt.agent, "marker")

			assert.Equal(t, "scout.toml", file)
			assert.Equal(t, tt.want, content)
		})
	}
}
