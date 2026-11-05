package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const wrapperTools = "readlink dirname find sort xargs sha256sum cut mkdir mv tar gzip chmod rm cat mktemp sed awk grep tr head basename ln cp"

type wrapperEnv struct {
	t        *testing.T
	pathDir  string
	served   string
	cache    string
	curlLog  string
	stubsLog string
}

func lookBash(t *testing.T) string {
	t.Helper()
	bash, err := exec.LookPath("bash")
	require.NoError(t, err)
	return bash
}

func newWrapperEnv(t *testing.T) *wrapperEnv {
	t.Helper()
	root := t.TempDir()
	env := &wrapperEnv{
		t:        t,
		pathDir:  filepath.Join(root, "path"),
		served:   filepath.Join(root, "served"),
		cache:    filepath.Join(root, "cache"),
		curlLog:  filepath.Join(root, "curl.log"),
		stubsLog: filepath.Join(root, "stubs.log"),
	}
	require.NoError(t, os.MkdirAll(env.pathDir, 0o755))
	require.NoError(t, os.MkdirAll(env.served, 0o755))
	for _, tool := range strings.Fields(wrapperTools) {
		if real, err := exec.LookPath(tool); err == nil {
			require.NoError(t, os.Symlink(real, filepath.Join(env.pathDir, tool)))
		}
	}
	env.stub("uname", "case \"$1\" in -s) echo Linux;; -m) echo x86_64;; esac")
	env.stub("curl", `out=""; url=""
while [ $# -gt 0 ]; do case "$1" in -o) out="$2"; shift 2;; -*) shift;; *) url="$1"; shift;; esac; done
echo "$url" >> `+env.curlLog+`
cp "`+env.served+`/$(basename "$url")" "$out"`)
	return env
}

func (e *wrapperEnv) stub(name, body string) {
	e.t.Helper()
	script := "#!" + lookBash(e.t) + "\n" + body + "\n"
	require.NoError(e.t, os.WriteFile(filepath.Join(e.pathDir, name), []byte(script), 0o755))
}

func pluginVersion(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("../../.claude-plugin/plugin.json")
	require.NoError(t, err)
	var manifest struct {
		Version string `json:"version"`
	}
	require.NoError(t, json.Unmarshal(raw, &manifest))
	require.NotEmpty(t, manifest.Version)
	return manifest.Version
}

func tarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))}))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

// serveRelease publishes the checksums of genuine, and the bytes of served as the archive.
func (e *wrapperEnv) serveRelease(genuine, served []byte) {
	version := pluginVersion(e.t)
	archive := "noodle_" + version + "_linux_amd64.tar.gz"
	sum := sha256.Sum256(genuine)
	require.NoError(e.t, os.WriteFile(filepath.Join(e.served, archive), served, 0o644))
	checksums := hex.EncodeToString(sum[:]) + "  " + archive + "\n"
	require.NoError(e.t, os.WriteFile(filepath.Join(e.served, "noodle_"+version+"_checksums.txt"), []byte(checksums), 0o644))
}

func (e *wrapperEnv) run(args ...string) (stdout, stderr string, exitCode int) {
	e.t.Helper()
	cmd := exec.Command(lookBash(e.t), append([]string{"../../bin/noodle"}, args...)...)
	cmd.Env = []string{"PATH=" + e.pathDir, "HOME=" + e.t.TempDir(), "CLAUDE_PLUGIN_DATA=" + e.cache}
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return out.String(), errOut.String(), exitErr.ExitCode()
	}
	require.NoError(e.t, err)
	return out.String(), errOut.String(), 0
}

func (e *wrapperEnv) downloads() []string {
	raw, err := os.ReadFile(e.curlLog)
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(e.t, err)
	return strings.Fields(string(raw))
}

func (e *wrapperEnv) cachedBinaries() []string {
	matches, err := filepath.Glob(filepath.Join(e.cache, "noodle-*"))
	require.NoError(e.t, err)
	return matches
}

func TestWrapper(t *testing.T) {
	t.Parallel()
	bash := lookBash(t)
	stubNoodle := "#!" + bash + "\necho \"downloaded noodle ran: $*\"\n"

	t.Run("download_verified", func(t *testing.T) {
		t.Parallel()
		env := newWrapperEnv(t)
		archive := tarGz(t, map[string]string{"noodle": stubNoodle, "LICENSE": "license text"})
		env.serveRelease(archive, archive)

		stdout, stderr, code := env.run("version", "--now")

		assert.Equal(t, 0, code, stderr)
		assert.Equal(t, "downloaded noodle ran: version --now\n", stdout)
		version := pluginVersion(t)
		assert.Equal(t, []string{
			"https://github.com/TheGostsniperfr/Noodle/releases/download/v" + version + "/noodle_" + version + "_linux_amd64.tar.gz",
			"https://github.com/TheGostsniperfr/Noodle/releases/download/v" + version + "/noodle_" + version + "_checksums.txt",
		}, env.downloads())
		assert.Len(t, env.cachedBinaries(), 1)
	})

	t.Run("download_cached_is_not_downloaded_twice", func(t *testing.T) {
		t.Parallel()
		env := newWrapperEnv(t)
		archive := tarGz(t, map[string]string{"noodle": stubNoodle})
		env.serveRelease(archive, archive)
		_, _, firstCode := env.run("one")

		stdout, _, secondCode := env.run("two")

		assert.Equal(t, 0, firstCode)
		assert.Equal(t, 0, secondCode)
		assert.Equal(t, "downloaded noodle ran: two\n", stdout)
		assert.Len(t, env.downloads(), 2)
	})

	t.Run("download_altered", func(t *testing.T) {
		t.Parallel()
		env := newWrapperEnv(t)
		genuine := tarGz(t, map[string]string{"noodle": stubNoodle})
		altered := tarGz(t, map[string]string{"noodle": "#!" + bash + "\necho tampered\n"})
		env.serveRelease(genuine, altered)

		stdout, stderr, code := env.run("version")

		assert.NotEqual(t, 0, code)
		assert.Empty(t, stdout)
		assert.Contains(t, stderr, "checksum mismatch")
		assert.Empty(t, env.cachedBinaries())
		leftovers, err := filepath.Glob(filepath.Join(env.cache, "download.*"))
		require.NoError(t, err)
		assert.Empty(t, leftovers)
	})

	t.Run("go_on_path", func(t *testing.T) {
		t.Parallel()
		env := newWrapperEnv(t)
		env.stub("go", `echo "go $*" >> `+env.stubsLog+`
while [ $# -gt 0 ]; do if [ "$1" = -o ]; then out="$2"; fi; shift; done
printf '#!`+bash+`\necho built by go\n' > "$out"; chmod +x "$out"`)

		stdout, stderr, code := env.run("version")

		assert.Equal(t, 0, code, stderr)
		assert.Equal(t, "built by go\n", stdout)
		assert.Empty(t, env.downloads())
	})

	t.Run("nix_on_path", func(t *testing.T) {
		t.Parallel()
		env := newWrapperEnv(t)
		storePath := filepath.Join(t.TempDir(), "store")
		require.NoError(t, os.MkdirAll(filepath.Join(storePath, "bin"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(storePath, "bin", "noodle"), []byte("#!"+bash+"\necho built by nix\n"), 0o755))
		env.stub("nix", `echo "nix $*" >> `+env.stubsLog+`
echo `+storePath)

		stdout, stderr, code := env.run("version")

		assert.Equal(t, 0, code, stderr)
		assert.Equal(t, "built by nix\n", stdout)
		assert.Empty(t, env.downloads())
	})

	t.Run("noodle_on_path", func(t *testing.T) {
		t.Parallel()
		env := newWrapperEnv(t)
		env.stub("noodle", `echo "noodle on path: $*"`)

		stdout, stderr, code := env.run("version")

		assert.Equal(t, 0, code, stderr)
		assert.Equal(t, "noodle on path: version\n", stdout)
		assert.Empty(t, env.downloads())
		assert.Empty(t, env.cachedBinaries())
	})

	t.Run("unsupported_platform_does_not_download", func(t *testing.T) {
		t.Parallel()
		env := newWrapperEnv(t)
		env.stub("uname", "case \"$1\" in -s) echo FreeBSD;; -m) echo x86_64;; esac")

		_, stderr, code := env.run("version")

		assert.NotEqual(t, 0, code)
		assert.Contains(t, stderr, "no release binary for FreeBSD")
		assert.Empty(t, env.downloads())
	})
}
