package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/TheGostsniperfr/Noodle/internal/diagram"
	"gopkg.in/yaml.v3"
)

func loadSpec(path string) (*diagram.Spec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loadSpec: %w", err)
	}
	var s diagram.Spec
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("loadSpec %s: %w", path, err)
	}
	return &s, nil
}
