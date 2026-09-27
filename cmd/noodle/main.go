// noodle turns an architecture spec (YAML) into a draw.io file in the house style,
// refusing to emit a diagram whose lines cross boxes or whose labels collide.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "render" {
		if err := renderCommand(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	out := flag.String("o", "", "output .drawio path (lint only when empty)")
	icons := flag.String("icons", "", "comma-separated directories of extra <name>.svg icons, searched before the built-in ones")
	themeName := flag.String("theme", "dark", "dark or light")
	listIcons := flag.Bool("list-icons", false, "print the available icon names and exit")
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
		fmt.Fprintln(os.Stderr, "usage: noodle render DIR [-view ID] [-o out.drawio] [-theme dark|light] [-icons DIR[,DIR]]\n       noodle [-o out.drawio] [-theme dark|light] [-icons DIR[,DIR]] spec.yaml   (v0 single file)\n       noodle -list-icons [-icons DIR]")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *out, *themeName, set); err != nil {
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

func run(specPath, out, themeName string, icons *iconSet) error {
	th, err := themeByName(themeName)
	if err != nil {
		return err
	}
	spec, err := loadSpec(specPath)
	if err != nil {
		return err
	}
	return lintAndRender(spec, out, th, icons)
}

func lintAndRender(spec *diagram.Spec, out string, th *Theme, icons *iconSet) error {
	findings := lint(spec)
	for _, f := range findings {
		fmt.Fprintf(os.Stderr, "lint: %-28s %s\n", f.where, f.msg)
	}
	if len(findings) > 0 {
		return fmt.Errorf("%d lint finding(s)", len(findings))
	}
	if out == "" {
		return nil
	}
	r := &renderer{spec: spec, th: th, icons: icons, cache: map[string]string{}, ports: portBadges(spec)}
	xml, err := r.render()
	if err != nil {
		return err
	}
	return os.WriteFile(out, []byte(xml), 0o644)
}
