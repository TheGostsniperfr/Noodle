// noodle turns an architecture spec (YAML) into a draw.io file in the house style,
// refusing to emit a diagram whose lines cross boxes or whose labels collide.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
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
		fmt.Fprintln(os.Stderr, "usage: noodle [-o out.drawio] [-theme dark|light] [-icons DIR[,DIR]] spec.yaml\n       noodle -list-icons [-icons DIR]")
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
	r := &renderer{spec: spec, th: th, icons: icons, cache: map[string]string{}, ports: spec.ports()}
	xml, err := r.render()
	if err != nil {
		return err
	}
	return os.WriteFile(out, []byte(xml), 0o644)
}
