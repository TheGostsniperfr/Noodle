package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/graph"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

const driftUsage = "usage: noodle drift DIR"

// driftCommand prints the gap between the curated model and the fragments in
// DIR/discovered. Drift is a report, so it exits 0 whatever it finds (ADR-0015).
func driftCommand(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf(driftUsage)
	}
	s, g, err := loadSystem(args[0])
	if err != nil {
		return err
	}
	if len(s.Fragments) == 0 {
		return fmt.Errorf("drift: %s has no fragment in discovered/; run noodle discover first", args[0])
	}
	_, err = io.WriteString(stdout, formatDrift(g.Drift()))
	return err
}

// formatDrift prints the four families, each under a heading at the start of the line
// and followed by its entries, indented.
func formatDrift(r graph.Report) string {
	var b strings.Builder
	for _, family := range []struct {
		heading string
		entries []graph.Entry
	}{{"unmatched", r.Unmatched}, {"gone", r.Gone}, {"planned", r.Planned}, {"unmodelled", r.Unmodelled}} {
		fmt.Fprintf(&b, "%s (%d)\n", family.heading, len(family.entries))
		for _, e := range family.entries {
			fields := []string{e.ID}
			if e.Detail != "" {
				fields = append(fields, e.Detail)
			}
			b.WriteString("  " + strings.Join(append(fields, location(e.Src)), "  ") + "\n")
		}
	}
	return b.String()
}

func location(s model.Src) string {
	switch {
	case s.File != "" && s.Line > 0:
		return fmt.Sprintf("%s:%d", s.File, s.Line)
	case s.File != "":
		return s.File
	}
	return s.Object
}
