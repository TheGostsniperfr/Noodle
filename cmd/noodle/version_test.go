package main

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildNoodle(t *testing.T, ldflags string) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "noodle")
	args := []string{"build", "-mod=vendor"}
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}
	out, err := exec.Command("go", append(args, "-o", binary, ".")...).CombinedOutput()
	require.NoError(t, err, string(out))
	return binary
}

func TestVersion_PrintsTheBuildVariable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		ldflags string
		want    string
	}{
		{"a release build prints its tag", "-X main.version=v0.1.0", "v0.1.0"},
		{"a plain build prints dev", "", "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			binary := buildNoodle(t, tt.ldflags)

			out, err := exec.Command(binary, "version").Output()

			require.NoError(t, err)
			assert.Equal(t, tt.want+"\n", string(out))
		})
	}
}

func TestVersion_RejectsArguments(t *testing.T) {
	t.Parallel()

	err := versionCommand([]string{"extra"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "extra")
}
