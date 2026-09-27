package model

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// System is one directory: a model, its views and the layouts of its topology views.
type System struct {
	Dir     string
	Model   *Model
	Views   map[string]*View
	Layouts map[string]*Layout
}

func (s *System) ViewIDs() []string {
	ids := make([]string, 0, len(s.Views))
	for id := range s.Views {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func LoadSystem(dir string) (*System, error) {
	s := &System{Dir: dir, Model: &Model{}, Views: map[string]*View{}, Layouts: map[string]*Layout{}}
	if err := decodeFile(filepath.Join(dir, "model.yaml"), "Model", s.Model); err != nil {
		return nil, fmt.Errorf("LoadSystem: %w", err)
	}
	views, err := yamlFiles(filepath.Join(dir, "views"))
	if err != nil {
		return nil, fmt.Errorf("LoadSystem: %w", err)
	}
	for _, path := range views {
		v := &View{ID: stem(path)}
		if err := decodeFile(path, "View", v); err != nil {
			return nil, fmt.Errorf("LoadSystem: %w", err)
		}
		s.Views[v.ID] = v
	}
	layouts, err := yamlFiles(filepath.Join(dir, "layouts"))
	if err != nil {
		return nil, fmt.Errorf("LoadSystem: %w", err)
	}
	for _, path := range layouts {
		l := &Layout{ID: stem(path)}
		if err := decodeFile(path, "Layout", l); err != nil {
			return nil, fmt.Errorf("LoadSystem: %w", err)
		}
		s.Layouts[l.ID] = l
	}
	return s, nil
}

// header is decoded first so a file with the wrong kind fails on its kind, not on the
// first field that kind does not know.
type header struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

func decodeFile(path, kind string, out any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var h header
	if err := yaml.Unmarshal(raw, &h); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if h.APIVersion != APIVersion {
		return fmt.Errorf("%s: apiVersion %q, want %q", path, h.APIVersion, APIVersion)
	}
	if h.Kind != kind {
		return fmt.Errorf("%s: kind %q, want %q", path, h.Kind, kind)
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func yamlFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	return paths, nil
}

func stem(path string) string { return strings.TrimSuffix(filepath.Base(path), ".yaml") }
