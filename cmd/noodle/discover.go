package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/TheGostsniperfr/Noodle/internal/adapter"
	"github.com/TheGostsniperfr/Noodle/internal/adapter/k8s"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

const discoverUsage = "usage: noodle discover ADAPTER [PATH…|-] [-o fragment.yaml] [-namespace NS] [-ref REF] [-observed-at RFC3339]"

// adapters lists every adapter noodle ships.
func adapters() (*adapter.Registry, error) { return adapter.NewRegistry(k8s.Adapter{}) }

// discoverCommand is "noodle discover ADAPTER [PATH…|-]": no path or - reads stdin.
// Flags may come before or after the arguments.
func discoverCommand(ctx context.Context, reg *adapter.Registry, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("discover", flag.ContinueOnError)
	out := fs.String("o", "", "output fragment path, usually <system>/discovered/<adapter>-<source>.yaml (stdout when empty)")
	ref := fs.String("ref", "", "version of the source, e.g. a Git commit")
	namespace := fs.String("namespace", "", "namespace of objects that name none, as kubectl apply -n (default: default)")
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
	src.Namespace = *namespace
	at := time.Now()
	if *observed != "" {
		if at, err = time.Parse(time.RFC3339, *observed); err != nil {
			return fmt.Errorf("-observed-at: %w", err)
		}
	}
	f, stats, err := adapter.Run(ctx, a, src, *ref, at)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := model.EncodeFragment(&buf, f); err != nil {
		return err
	}
	settled := *f
	settled.Unresolved = nil
	var without bytes.Buffer
	if err := model.EncodeFragment(&without, &settled); err != nil {
		return err
	}
	fmt.Fprintln(stderr, summary(a.Name(), stats, f, buf.Len(), buf.Len()-without.Len()))
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

// summary is the line that tells whether discovery pays: what was read against what an
// agent will read instead. Tokens are estimated at four bytes each.
func summary(name string, st adapter.Stats, f *model.Fragment, fragmentBytes, unresolvedBytes int) string {
	skipped, reasons := 0, make([]string, 0, len(st.Skipped))
	for r, n := range st.Skipped {
		skipped += n
		reasons = append(reasons, fmt.Sprintf("%s %d", r, n))
	}
	sort.Strings(reasons)
	why := ""
	if len(reasons) > 0 {
		why = " (" + strings.Join(reasons, ", ") + ")"
	}
	noise := ""
	if st.NoiseBytes > 0 {
		noise = fmt.Sprintf(" (~%s without CRD schemas)", kTokens(st.InputBytes-st.NoiseBytes))
	}
	return fmt.Sprintf("%s: read %d objects, ~%s tokens%s; skipped %d%s; fragment %d elements, %d connections, %d references, %d unresolved, ~%s tokens (unresolved ~%s)",
		name, st.Objects, kTokens(st.InputBytes), noise, skipped, why,
		len(f.Elements), len(f.Connections), len(f.References), len(f.Unresolved), kTokens(fragmentBytes), kTokens(unresolvedBytes))
}

func kTokens(bytes int) string {
	t := float64(bytes) / 4
	if t < 1000 {
		return fmt.Sprintf("%.0f", t)
	}
	return fmt.Sprintf("%.1fk", t/1000)
}
