package resources

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// developmentTreePresent reports whether the checkout carries the ORIGEN/IC_SDR
// payload that Path falls back to during development. That directory is
// gitignored, so it is absent from a fresh clone and from CI.
func developmentTreePresent(t *testing.T) bool {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat(filepath.Join(root, "ORIGEN", "IC_SDR"))
	return err == nil
}

func TestDevelopmentResourceCanBeLocatedOutsideWorkingDirectory(t *testing.T) {
	path := Path("tools", "aprs", "config", "direwolf-rx.conf")
	// Whether or not the resource exists, the path handed to a decoder or
	// printed in an error message has to be absolute: the tools IC-SDR starts
	// do not inherit its working directory.
	if !filepath.IsAbs(path) {
		t.Fatalf("resource path must be absolute: %q", path)
	}
	if !developmentTreePresent(t) {
		t.Skip("ORIGEN/IC_SDR is gitignored and absent from this checkout, so there is no development resource to locate")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("resource path %q is not usable: %v", path, err)
	}
}

func TestWritablePathStaysInsideProjectData(t *testing.T) {
	got := WritablePath("recordings", "capture.wav")
	if !filepath.IsAbs(got) {
		t.Fatalf("writable path must be absolute: %q", got)
	}
	// Mutable state always lands under a DATA directory, whichever root is in
	// use: DATA beside the executable in a release build, or the checkout's
	// DATA during development.
	want := filepath.Join("DATA", "recordings", "capture.wav")
	if !strings.HasSuffix(got, string(filepath.Separator)+want) {
		t.Fatalf("writable path escaped DATA: %q; want it to end in %q", got, want)
	}

	if !developmentTreePresent(t) {
		return
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	// With the development tree present the root is pinned to the checkout.
	prefix := filepath.Join(root, "DATA") + string(filepath.Separator)
	if !strings.HasPrefix(got, prefix) {
		t.Fatalf("writable path escaped the checkout: %q; want prefix %q", got, prefix)
	}
}
