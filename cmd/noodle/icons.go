package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/TheGostsniperfr/Noodle/assets"
)

// iconSet resolves icon names against project directories first, then the icons
// compiled into the binary, so a project can add or override a logo without a rebuild.
type iconSet struct {
	sources []fs.FS
}

func newIconSet(dirs []string) (*iconSet, error) {
	s := &iconSet{}
	for _, d := range dirs {
		if _, err := os.Stat(d); err != nil {
			return nil, fmt.Errorf("icons dir: %w", err)
		}
		s.sources = append(s.sources, os.DirFS(d))
	}
	embedded, err := fs.Sub(assets.Icons, "icons")
	if err != nil {
		return nil, fmt.Errorf("embedded icons: %w", err)
	}
	s.sources = append(s.sources, embedded)
	return s, nil
}

// read prefers <name>.<theme>.svg so a logo can swap to a variant that stays visible on the background.
func (s *iconSet) read(name, theme string) ([]byte, error) {
	for _, src := range s.sources {
		for _, file := range []string{name + "." + theme + ".svg", name + ".svg"} {
			raw, err := fs.ReadFile(src, file)
			if err == nil {
				return raw, nil
			}
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("icon %q: %w", name, err)
			}
		}
	}
	return nil, fmt.Errorf("icon %q not found: add %s.svg to a directory passed with -icons, or run noodle -list-icons", name, name)
}

func (s *iconSet) names() []string {
	seen := map[string]bool{}
	for _, src := range s.sources {
		entries, _ := fs.ReadDir(src, ".")
		for _, e := range entries {
			n, ok := strings.CutSuffix(e.Name(), ".svg")
			if !ok {
				continue
			}
			n, _, _ = strings.Cut(n, ".")
			seen[n] = true
		}
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
