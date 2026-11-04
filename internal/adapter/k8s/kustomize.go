package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// kustomizeBuild renders a directory the way Argo CD does, without writing to it
// (ADR-0027). The tool is looked up before anything else is read, so a missing one
// fails the run instead of silently dropping the directory.
func kustomizeBuild(ctx context.Context, dir string, enableHelm bool) ([]byte, error) {
	tool, err := exec.LookPath("kustomize")
	if err != nil {
		return nil, errors.New("kustomize not found: nix shell nixpkgs#kustomize, or nix develop in this repository")
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("-kustomize %s: not a directory", dir)
	}
	args := []string{"build", dir}
	if enableHelm {
		args = append(args, "--enable-helm")
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, tool, args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("kustomize %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}
