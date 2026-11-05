package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/TheGostsniperfr/Noodle/internal/graph"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// loadSystem is how a command reads a system: the files, the check of the curated model
// alone, then the fragments merged into it (FR-010). The returned system's Model is the
// merged one.
func loadSystem(dir string) (*model.System, *graph.Graph, error) {
	s, err := model.LoadSystem(dir)
	if err != nil {
		return nil, nil, err
	}
	if findings := model.Check(s); len(findings) > 0 {
		for _, f := range findings {
			fmt.Fprintf(os.Stderr, "check: %s\n", f)
		}
		return nil, nil, fmt.Errorf("%d check finding(s) in %s", len(findings), dir)
	}
	fragments := make([]*model.Fragment, 0, len(s.Fragments))
	for _, id := range sortedFragmentIDs(s) {
		fragments = append(fragments, s.Fragments[id])
	}
	g := graph.Merge(s.Model, fragments...)
	s.Model = g.Model
	return s, g, nil
}

func sortedFragmentIDs(s *model.System) []string {
	ids := make([]string, 0, len(s.Fragments))
	for id := range s.Fragments {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
