package main

import (
	"fmt"

	"github.com/TheGostsniperfr/Noodle/internal/migrate"
	"github.com/TheGostsniperfr/Noodle/internal/model"
)

// migrateCommand is "noodle migrate SPEC DIR": a v0 single file to a v1alpha1 system. It
// writes nothing unless the result draws the same diagram as the v0 file.
func migrateCommand(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: noodle migrate SPEC.yaml DIR")
	}
	spec, err := loadSpec(args[0])
	if err != nil {
		return err
	}
	s, err := migrate.FromV0(spec)
	if err != nil {
		return err
	}
	if err := migrate.Verify(spec, s); err != nil {
		return err
	}
	if err := migrate.Write(s, args[1]); err != nil {
		return err
	}
	written, err := model.LoadSystem(args[1])
	if err != nil {
		return fmt.Errorf("migrate: reading back %s: %w", args[1], err)
	}
	fresh, err := loadSpec(args[0])
	if err != nil {
		return err
	}
	return migrate.Verify(fresh, written)
}
