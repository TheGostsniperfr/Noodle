package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	noodle "github.com/TheGostsniperfr/Noodle"
	"github.com/TheGostsniperfr/Noodle/internal/skillrules"
)

const (
	embeddedSkillDir = "skills/noodle-diagram"
	sharedSkillDir   = ".agents/skills/noodle-diagram"
	setupUsage       = "usage: noodle setup opencode|codex (-project [DIR] | -user)"
)

// agent is one of the subagents shipped in agents/, read from its Claude Code file.
type agent struct {
	name        string
	description string
	readOnly    bool
	body        string
}

// host knows where one coding agent expects subagents and how it spells them.
type host struct {
	projectAgentDir string
	userAgentDir    string
	agentFile       func(a agent, marker string) (name string, content string)
}

var hosts = map[string]host{
	"opencode": {projectAgentDir: ".opencode/agents", userAgentDir: ".config/opencode/agents", agentFile: openCodeAgent},
	"codex":    {projectAgentDir: ".codex/agents", userAgentDir: ".codex/agents", agentFile: codexAgent},
}

type setupFlags struct {
	set           *flag.FlagSet
	project, user *bool
}

func newSetupFlags() setupFlags {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	return setupFlags{
		set:     fs,
		project: fs.Bool("project", false, "write under DIR (default .)"),
		user:    fs.Bool("user", false, "write under the home directory"),
	}
}

// setupEnv is what runSetup needs from its surroundings, so tests can run it anywhere.
type setupEnv struct {
	stdout, stderr io.Writer
	home           string
	version        string
}

// setupCommand is "noodle setup HOST (-project|-user) [DIR]": it writes the noodle-diagram
// skill and the subagents in the format the host reads.
func setupCommand(args []string, stdout, stderr io.Writer) error {
	home, _ := os.UserHomeDir()
	return runSetup(args, setupEnv{stdout: stdout, stderr: stderr, home: home, version: version})
}

func runSetup(args []string, env setupEnv) error {
	if len(args) == 0 {
		return errors.New(setupUsage)
	}
	hostName := args[0]
	h, ok := hosts[hostName]
	if !ok {
		return fmt.Errorf("setup: unknown host %q\n%s", args[0], setupUsage)
	}
	opts := newSetupFlags()
	flags, project, user := opts.set, opts.project, opts.user
	flags.SetOutput(env.stderr)
	var dirs []string
	rest := args[1:]
	for {
		if err := flags.Parse(rest); err != nil {
			return err
		}
		if flags.NArg() == 0 {
			break
		}
		dirs = append(dirs, flags.Arg(0))
		rest = flags.Args()[1:]
	}
	if *project == *user {
		return errors.New("setup: give exactly one of -project and -user\n" + setupUsage)
	}
	if len(dirs) > 1 || (*user && len(dirs) > 0) {
		return errors.New("setup: -project takes at most one DIR and -user takes none\n" + setupUsage)
	}

	if *user {
		if env.home == "" {
			return errors.New("setup: cannot find the home directory")
		}
		return writeSetup(h, hostName, env, env.home, h.userAgentDir, "~/")
	}
	base := "."
	if len(dirs) == 1 {
		base = dirs[0]
	}
	return writeSetup(h, hostName, env, base, h.projectAgentDir, "")
}

func writeSetup(h host, hostName string, env setupEnv, base, agentDir, display string) error {
	files, err := setupFiles(h, hostName, env.version, agentDir)
	if err != nil {
		return err
	}
	rels := make([]string, 0, len(files))
	for rel := range files {
		rels = append(rels, rel)
	}
	sort.Strings(rels)

	kept := 0
	for _, rel := range rels {
		target := filepath.Join(base, filepath.FromSlash(rel))
		existing, err := os.ReadFile(target)
		switch {
		case err == nil && string(existing) == files[rel]:
			continue
		case err == nil && !markerRe.Match(existing):
			fmt.Fprintf(env.stderr, "kept %s: it was not written by noodle setup\n", display+rel)
			kept++
			continue
		case err != nil && !errors.Is(err, fs.ErrNotExist):
			return fmt.Errorf("setup: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		if err := os.WriteFile(target, []byte(files[rel]), 0o644); err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		fmt.Fprintf(env.stdout, "wrote %s\n", display+rel)
	}
	if kept > 0 {
		return fmt.Errorf("setup: %d file(s) kept", kept)
	}
	return nil
}

// setupFiles maps each file to write, by path relative to the base directory, to its content.
func setupFiles(h host, hostName, version, agentDir string) (map[string]string, error) {
	files := map[string]string{}
	err := fs.WalkDir(noodle.Content, embeddedSkillDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := fs.ReadFile(noodle.Content, p)
		if err != nil {
			return err
		}
		files[path.Join(sharedSkillDir, strings.TrimPrefix(p, embeddedSkillDir+"/"))] = markMarkdown(string(raw), markerText("skills", version))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("setup: read embedded skill: %w", err)
	}
	agents, err := loadAgents()
	if err != nil {
		return nil, err
	}
	for _, a := range agents {
		name, content := h.agentFile(a, markerText(hostName, version))
		files[path.Join(agentDir, name)] = content
	}
	return files, nil
}

// The marker is the first line of the body of a Markdown file and the first comment of a
// TOML file, never a frontmatter key, so the skill still follows the Agent Skills rules.
// Rerunning setup replaces a file that has it and keeps one that does not.
var markerRe = regexp.MustCompile(`(?m)^(?:<!-- |# )noodle setup \S+ \S+, do not edit: rerun noodle setup(?: -->)?$`)

func markerText(target, version string) string {
	return "noodle setup " + target + " " + version + ", do not edit: rerun noodle setup"
}

func markMarkdown(content, text string) string {
	_, body, hasFrontmatter := skillrules.SplitFrontmatter(content)
	if !hasFrontmatter {
		return "<!-- " + text + " -->\n\n" + content
	}
	return content[:len(content)-len(body)] + "<!-- " + text + " -->\n" + body
}

func loadAgents() ([]agent, error) {
	entries, err := fs.ReadDir(noodle.Content, "agents")
	if err != nil {
		return nil, fmt.Errorf("setup: read embedded agents: %w", err)
	}
	var out []agent
	for _, e := range entries {
		raw, err := fs.ReadFile(noodle.Content, "agents/"+e.Name())
		if err != nil {
			return nil, fmt.Errorf("setup: %w", err)
		}
		front, body, ok := skillrules.SplitFrontmatter(string(raw))
		if !ok {
			return nil, fmt.Errorf("setup: agents/%s has no frontmatter", e.Name())
		}
		var meta struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
			Tools       string `yaml:"tools"`
		}
		if err := yaml.Unmarshal([]byte(front), &meta); err != nil {
			return nil, fmt.Errorf("setup: agents/%s: %w", e.Name(), err)
		}
		readOnly := true
		for _, tool := range strings.Split(meta.Tools, ",") {
			if t := strings.TrimSpace(tool); t == "Write" || t == "Edit" {
				readOnly = false
			}
		}
		out = append(out, agent{name: meta.Name, description: meta.Description, readOnly: readOnly, body: body})
	}
	return out, nil
}

func openCodeAgent(a agent, marker string) (string, string) {
	type permission struct {
		Edit string `yaml:"edit"`
	}
	meta := struct {
		Description string      `yaml:"description"`
		Mode        string      `yaml:"mode"`
		Permission  *permission `yaml:"permission,omitempty"`
	}{Description: a.description, Mode: "subagent"}
	if a.readOnly {
		meta.Permission = &permission{Edit: "deny"}
	}
	front, _ := yaml.Marshal(meta)
	return a.name + ".md", markMarkdown("---\n"+string(front)+"---\n"+a.body, marker)
}

func codexAgent(a agent, marker string) (string, string) {
	instructions := strings.ReplaceAll(strings.ReplaceAll(a.body, `\`, `\\`), `"""`, `\"\"\"`)
	content := "# " + marker + "\nname = " + tomlString(a.name) + "\n" +
		"description = " + tomlString(a.description) + "\n" +
		"developer_instructions = \"\"\"\n" + instructions + "\"\"\"\n"
	return a.name + ".toml", content
}

func tomlString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"' || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, `\u%04X`, r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
