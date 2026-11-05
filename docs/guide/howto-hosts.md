# How to use noodle in OpenCode or Codex

noodle ships a skill, `noodle-diagram`, and three subagents: `fact-finder`, which reads
code and returns the facts a diagram needs, `diagram-reviewer`, which critiques a rendered
diagram, and `icon-curator`, which fetches a missing logo. Claude Code installs them from a
plugin. For OpenCode and Codex, the `noodle` binary writes them where those hosts look.

Install the binary first: `nix run github:TheGostsniperfr/Noodle`, or an archive from the
[releases](https://github.com/TheGostsniperfr/Noodle/releases) on your PATH.

## OpenCode

```bash
noodle setup opencode -project
```

Writes five files, listing each as `wrote <path>`:

- `.agents/skills/noodle-diagram/SKILL.md` and `reference.md`
- `.opencode/agents/diagram-reviewer.md`, `fact-finder.md` and `icon-curator.md`

The read-only agents have `edit: deny`. With `-user`, the files go under `~/.agents/skills/`
and `~/.config/opencode/agents/` and apply to every project.

## Codex

```bash
noodle setup codex -project
```

Writes the same skill to `.agents/skills/noodle-diagram/` and one TOML file per agent to
`.codex/agents/`, with `name`, `description` and `developer_instructions`. With `-user`,
the files go under `~/.agents/skills/` and `~/.codex/agents/`. Codex does not let an agent
file restrict file edits, so the read-only agents rely on their instructions.

## Rerunning and upgrading

Every file carries a first line starting with `noodle setup`. Rerunning the command changes
nothing when the files are current, and replaces files that carry the line when the noodle
version changed. A file without that line is yours: noodle keeps it as it is, names it on
standard error and exits with status 1. Move it away, or merge by hand, and run again.

## Without the subagents

The skill works without them. When no `fact-finder` is available, the agent reads the code
itself. It costs more context, so prefer running discovery first (see
[discover a system](howto-discover.md)).
