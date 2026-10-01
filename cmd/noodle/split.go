package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
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
	slide := fs.Bool("slide", false, "drawing only, cropped to its content: no header, cards or notes")
	lens := fs.String("lens", "", "lens id of a topology view: lit for one subject (ADR-0020)")
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
	th, err := house.ThemeByName(*themeName)
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
	v, ok := s.Views[id]
	if !ok {
		return fmt.Errorf("resolve: no view %q in %s", id, dirs[0])
	}
	if v.Type == "sequence" {
		return renderSequenceView(s, id, *out, th, set)
	}
	var spec *diagram.Spec
	if *lens != "" {
		spec, err = resolve.Lens(s, id, *lens)
	} else {
		spec, err = resolve.View(s, id)
	}
	if err != nil {
		return err
	}
	return lintAndRender(spec, *out, th, set, *slide)
}

// renderSequenceView lints the participant boxes only: rows and columns are computed,
// so labels and arrows cannot collide (ADR-0007).
func renderSequenceView(s *model.System, id, out string, th *house.Theme, icons *iconSet) error {
	seq, err := resolve.Sequence(s, id)
	if err != nil {
		return err
	}
	findings := lint.Lint(&seq.Frame)
	for _, f := range findings {
		fmt.Fprintf(os.Stderr, "lint: %-28s %s\n", f.Where, f.Msg)
	}
	if len(findings) > 0 {
		return fmt.Errorf("%d lint finding(s)", len(findings))
	}
	if out == "" {
		return nil
	}
	r := &renderer{th: th, icons: icons, cache: map[string]string{}}
	xml, err := r.renderSequence(seq)
	if err != nil {
		return err
	}
	return os.WriteFile(out, []byte(xml), 0o644)
}
