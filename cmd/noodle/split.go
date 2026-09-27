package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/model"
	"github.com/TheGostsniperfr/Noodle/internal/resolve"
)

// renderCommand is "noodle render DIR [-view ID]": a system directory in the v1alpha1
// format (ADR-0007). Flags may come before or after DIR.
func renderCommand(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	viewID := fs.String("view", "", "view id, the file stem in views/ (optional when the system has one view)")
	out := fs.String("o", "", "output .drawio path (lint only when empty)")
	icons := fs.String("icons", "", "comma-separated directories of extra <name>.svg icons")
	themeName := fs.String("theme", "dark", "dark or light")
	var dirs []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return err
		}
		if fs.NArg() == 0 {
			break
		}
		dirs = append(dirs, fs.Arg(0))
		args = fs.Args()[1:]
	}
	if len(dirs) != 1 {
		return fmt.Errorf("usage: noodle render DIR [-view ID] [-o out.drawio] [-theme dark|light] [-icons DIR[,DIR]]")
	}
	set, err := newIconSet(splitList(*icons))
	if err != nil {
		return err
	}
	th, err := themeByName(*themeName)
	if err != nil {
		return err
	}
	s, err := model.LoadSystem(dirs[0])
	if err != nil {
		return err
	}
	if findings := model.Check(s); len(findings) > 0 {
		for _, f := range findings {
			fmt.Fprintf(os.Stderr, "check: %s\n", f)
		}
		return fmt.Errorf("%d check finding(s) in %s", len(findings), dirs[0])
	}
	id := *viewID
	if id == "" {
		if ids := s.ViewIDs(); len(ids) == 1 {
			id = ids[0]
		} else {
			return fmt.Errorf("%s has %d views, pick one with -view: %s", dirs[0], len(ids), strings.Join(ids, ", "))
		}
	}
	spec, err := resolve.Topology(s, id)
	if err != nil {
		return err
	}
	return lintAndRender(spec, *out, th, set)
}
