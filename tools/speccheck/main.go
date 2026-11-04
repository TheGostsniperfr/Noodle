// Command speccheck enforces the engineering process of docs/PROCESS.md (ADR-0024) with
// no model in the loop: spec traceability, ADR shape and index, the command reference,
// and the evidence gate on docs/features.json.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const usage = `usage: speccheck [-root DIR] COMMAND
  all                 spec + adr + commands + skills (what make spec-check runs)
  spec [SPEC_DIR...]  traceability of every spec, or of the given ones
  adr [-write]        ADR shape, and docs/adr/README.md index (rewritten with -write)
  commands            Makefile targets and docs/COMMANDS.md list the same commands
  skills              every SKILL.md follows the Agent Skills rules
  features [-base REF]  features.json history and evidence for every new pass
  locate FEATURE_ID   print the spec directory that owns a feature`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if !errors.Is(err, errFindings) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

var errFindings = errors.New("findings")

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("speccheck", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return errors.New(usage)
	}
	repo := &Repo{Root: *root}
	cmd, rest := fs.Arg(0), fs.Args()[1:]

	var findings []Finding
	var err error
	switch cmd {
	case "all":
		for _, check := range []func() ([]Finding, error){
			func() ([]Finding, error) { return repo.CheckSpecs(nil) },
			func() ([]Finding, error) { return repo.CheckADRs(false) },
			repo.CheckCommands,
			repo.CheckSkills,
		} {
			f, cerr := check()
			if cerr != nil {
				return cerr
			}
			findings = append(findings, f...)
		}
	case "spec":
		findings, err = repo.CheckSpecs(rest)
	case "adr":
		afs := flag.NewFlagSet("adr", flag.ContinueOnError)
		write := afs.Bool("write", false, "rewrite the index in docs/adr/README.md")
		if err := afs.Parse(rest); err != nil {
			return err
		}
		findings, err = repo.CheckADRs(*write)
	case "commands":
		findings, err = repo.CheckCommands()
	case "skills":
		findings, err = repo.CheckSkills()
	case "features":
		ffs := flag.NewFlagSet("features", flag.ContinueOnError)
		base := ffs.String("base", "HEAD", "git ref to compare docs/features.json against")
		if err := ffs.Parse(rest); err != nil {
			return err
		}
		findings, err = repo.CheckFeatures(*base, os.Getenv("ALLOW_FEATURE_EDIT") == "1")
	case "locate":
		if len(rest) != 1 {
			return errors.New("usage: speccheck locate FEATURE_ID")
		}
		dir, lerr := repo.Locate(rest[0])
		if lerr != nil {
			return lerr
		}
		fmt.Fprintln(stdout, filepath.ToSlash(dir))
		return nil
	default:
		return fmt.Errorf("unknown command %q\n%s", cmd, usage)
	}
	if err != nil {
		return err
	}
	for _, f := range findings {
		fmt.Fprintln(stderr, f)
	}
	if len(findings) > 0 {
		return errFindings
	}
	return nil
}

// Finding is one process violation, printed as "file: message".
type Finding struct {
	File string
	Msg  string
}

func (f Finding) String() string { return f.File + ": " + f.Msg }
