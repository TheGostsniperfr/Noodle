// Command tokencost reads a Claude Code transcript and prints what it cost: weighted
// tokens and dollars per tool category and per model, the number of turns and the wall
// time (spec 009 FR-013). Weights follow the API price ratios: cache read x0.1, cache
// write x1.25, output x5, relative to uncached input.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const usage = "usage: tokencost [TRANSCRIPT.jsonl]   (stdin when no file is given)"

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	in := stdin
	switch len(args) {
	case 0:
	case 1:
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("run: %w", err)
		}
		defer f.Close()
		in = f
	default:
		return errors.New(usage)
	}
	t, err := readTranscript(in)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return t.print(stdout)
}
