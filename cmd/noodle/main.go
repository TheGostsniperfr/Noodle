// noodle turns an architecture spec (YAML) into a draw.io file in the house style,
// refusing to emit a diagram whose lines cross boxes or whose labels collide.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"github.com/TheGostsniperfr/Noodle/internal/house"
	"github.com/TheGostsniperfr/Noodle/internal/lint"
)

const slideMargin = 24.0

func main() {
	if len(os.Args) > 1 {
		commands := map[string]func([]string) error{"render": renderCommand, "migrate": migrateCommand}
		if cmd, ok := commands[os.Args[1]]; ok {
			if err := cmd(os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
	}
	out := flag.String("o", "", "output .drawio path (lint only when empty)")
	icons := flag.String("icons", "", "comma-separated directories of extra <name>.svg icons, searched before the built-in ones")
	themeName := flag.String("theme", "dark", "dark or light")
	listIcons := flag.Bool("list-icons", false, "print the available icon names and exit")
	slide := flag.Bool("slide", false, "drawing only, cropped to its content: no header, cards or notes")
	flag.Parse()

	set, err := newIconSet(splitList(*icons))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *listIcons {
		fmt.Println(strings.Join(set.names(), "\n"))
		return
	}
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: noodle render DIR [-view ID] [-o out.drawio] [-theme dark|light] [-icons DIR[,DIR]]\n       noodle [-o out.drawio] [-theme dark|light] [-icons DIR[,DIR]] spec.yaml   (v0 single file)\n       noodle migrate SPEC.yaml DIR   (v0 single file to a v1alpha1 system)\n       noodle -list-icons [-icons DIR]")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *out, *themeName, set, *slide); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func run(specPath, out, themeName string, icons *iconSet, slide bool) error {
	th, err := house.ThemeByName(themeName)
	if err != nil {
		return err
	}
	spec, err := loadSpec(specPath)
	if err != nil {
		return err
	}
	return lintAndRender(spec, out, th, icons, slide)
}

// lintAndRender lints the full diagram, so a slide crop never hides a finding.
func lintAndRender(spec *diagram.Spec, out string, th *house.Theme, icons *iconSet, slide bool) error {
	errors := 0
	for _, f := range lint.Lint(spec) {
		level := "lint"
		if f.Warn {
			level = "warn"
		} else {
			errors++
		}
		fmt.Fprintf(os.Stderr, "%s: %-28s %s\n", level, f.Where, f.Msg)
	}
	if errors > 0 {
		return fmt.Errorf("%d lint finding(s)", errors)
	}
	if out == "" {
		return nil
	}
	if slide {
		spec.ForSlide(slideMargin)
	}
	r := &renderer{spec: spec, th: th, icons: icons, cache: map[string]string{}, ports: house.PortBadges(spec)}
	xml, err := r.render()
	if err != nil {
		return err
	}
	return os.WriteFile(out, []byte(xml), 0o644)
}
