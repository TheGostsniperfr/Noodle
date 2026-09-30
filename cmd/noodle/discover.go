package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/TheGostsniperfr/Noodle/internal/adapter"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

const discoverUsage = "usage: noodle discover ADAPTER [PATH…|-] [-o fragment.yaml] [-ref REF] [-observed-at RFC3339]"

// adapters lists every adapter noodle ships.
func adapters() (*adapter.Registry, error) { return adapter.NewRegistry() }

// discoverCommand is "noodle discover ADAPTER [PATH…|-]": no path or - reads stdin.
// Flags may come before or after the arguments.
func discoverCommand(ctx context.Context, reg *adapter.Registry, args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("discover", flag.ContinueOnError)
	out := fs.String("o", "", "output fragment path, usually <system>/discovered/<adapter>-<source>.yaml (stdout when empty)")
	ref := fs.String("ref", "", "version of the source, e.g. a Git commit")
	observed := fs.String("observed-at", "", "observation time, RFC 3339 (now when empty)")
	var pos []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return err
		}
		if fs.NArg() == 0 {
			break
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
	if len(pos) == 0 {
		return fmt.Errorf("%s\nadapters: %v", discoverUsage, reg.Names())
	}
	a, err := reg.Get(pos[0])
	if err != nil {
		return err
	}
	src, err := source(pos[1:], stdin)
	if err != nil {
		return err
	}
	at := time.Now()
	if *observed != "" {
		if at, err = time.Parse(time.RFC3339, *observed); err != nil {
			return fmt.Errorf("-observed-at: %w", err)
		}
	}
	f, err := adapter.Run(ctx, a, src, *ref, at)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := model.EncodeFragment(&buf, f); err != nil {
		return err
	}
	if *out == "" {
		_, err = stdout.Write(buf.Bytes())
		return err
	}
	return os.WriteFile(*out, buf.Bytes(), 0o644)
}

func source(paths []string, stdin io.Reader) (adapter.Source, error) {
	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "-") {
		return adapter.Source{Reader: stdin}, nil
	}
	for _, p := range paths {
		if p == "-" {
			return adapter.Source{}, fmt.Errorf("discover: - reads stdin and cannot be mixed with paths")
		}
	}
	return adapter.Source{Paths: paths}, nil
}
