// Package assets holds the files compiled into the noodle binary, so a plugin or Nix
// install needs no icon directory on disk.
package assets

import "embed"

//go:embed icons/*.svg
var Icons embed.FS
