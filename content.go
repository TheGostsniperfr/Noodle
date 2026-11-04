// Package noodle carries the skill and the subagents inside the binary, so noodle setup
// can write them for hosts that have no plugin system.
package noodle

import "embed"

//go:embed skills agents
var Content embed.FS
