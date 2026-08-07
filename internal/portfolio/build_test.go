//nolint:testpackage // white-box: this file exercises templateFuncs, which is unexported.
package portfolio

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

// TestExtractThenBuild pins AC-4 (§10.5): extract the real source into a
// temp YAML, then build from it, both succeeding with no manual step. It
// doubles as the regression test for L1/L2/L3 agreeing with V1..V5.
func TestExtractThenBuild(t *testing.T) {
	t.Parallel()

	yamlPath := filepath.Join(t.TempDir(), "resume.yaml")
	if err := Extract(srcMD, yamlPath); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	outDir := t.TempDir()
	if err := Build(yamlPath, tmplPath, outDir); err != nil {
		t.Fatalf("Build: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil {
		t.Errorf("expected index.html in %s: %v", outDir, err)
	}
}

// TestBuildWritesOneFile pins AC-8/FR-39 (§10.7): build writes exactly one
// file, into the given out-dir, and never touches
// page/reydavid_experience.md.
func TestBuildWritesOneFile(t *testing.T) {
	t.Parallel()

	before, err := os.ReadFile(srcMD)
	if err != nil {
		t.Fatalf("reading %s: %v", srcMD, err)
	}

	wantSum := sha256.Sum256(before)

	outDir := t.TempDir()
	if err := Build(dataYAML, tmplPath, outDir); err != nil {
		t.Fatalf("Build: %v", err)
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("reading %s: %v", outDir, err)
	}

	if len(entries) != 1 || entries[0].Name() != "index.html" {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}

		t.Fatalf("outDir contents = %v, want exactly [index.html]", names)
	}

	after, err := os.ReadFile(srcMD)
	if err != nil {
		t.Fatalf("reading %s after Build: %v", srcMD, err)
	}

	if gotSum := sha256.Sum256(after); gotSum != wantSum {
		t.Errorf("%s hash changed after Build; source must survive untouched", srcMD)
	}
}
