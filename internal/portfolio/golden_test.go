//nolint:testpackage // white-box: this file exercises mapDoc, parseDoc and stripLeadingComment, which are unexported.
package portfolio

// Golden-fixture suite, per ARCHITECTURE.md §10.1: each testdata/drift/<name>
// fixture pins exactly one FR-37 parser tolerance rule, and TestBaseline
// pins AC-2 against the real committed source. Run with -update to
// regenerate want.yaml after a deliberate parser change (AD-12: a fixture
// break is a contract break, not a test to update blindly).

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

//nolint:gochecknoglobals // A test flag must be registered at package level.
var update = flag.Bool("update", false, "regenerate testdata/drift/*/want.yaml")

// TestDrift walks every testdata/drift/<name> directory (never hardcoded,
// so adding a fixture per AD-12's rule is enough to pin it), extracts
// in.md, and compares the byte-exact output against want.yaml. want.yaml
// is the production marshaller's own output, never hand-written (AD-12).
func TestDrift(t *testing.T) {
	t.Parallel()

	dirs, err := filepath.Glob("testdata/drift/*")
	if err != nil {
		t.Fatalf("globbing testdata/drift: %v", err)
	}

	if len(dirs) == 0 {
		t.Fatal("no fixtures found under testdata/drift")
	}

	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()

			in := filepath.Join(dir, "in.md")
			want := filepath.Join(dir, "want.yaml")
			got := filepath.Join(t.TempDir(), "resume.yaml")

			if err := Extract(in, got); err != nil {
				t.Fatalf("Extract(%s) failed the V1..V5 gate: %v", in, err)
			}

			gotBytes, err := os.ReadFile(got)
			if err != nil {
				t.Fatalf("reading extract output: %v", err)
			}

			if *update {
				//nolint:gosec // want is a testdata path from the glob above, not attacker-controlled.
				if err := os.WriteFile(want, gotBytes, 0o644); err != nil {
					t.Fatalf("writing %s: %v", want, err)
				}

				return
			}

			wantBytes, err := os.ReadFile(want)
			if err != nil {
				t.Fatalf("reading %s (run `make golden` to generate it): %v", want, err)
			}

			if !bytes.Equal(gotBytes, wantBytes) {
				t.Errorf("%s drifted from want.yaml\n--- got ---\n%s\n--- want ---\n%s", dir, gotBytes, wantBytes)
			}
		})
	}
}

// The two counts AC-2 asserts are blocked (ARCHITECTURE.md §12 C-2/C-3):
// AC-2's own text says 7 entries and 18 stories, but the committed
// page/reydavid_experience.md has no Noodle entry (6 companies on disk) and
// FR-9 has no bullet-selection rule to prune 42 story bullets down to 18.
// Verified this session: `go run . extract` prints "6 entries, 42 stories".
// Ruling C-2/C-3 is a one-line change to these two constants.
const (
	wantEntries = 6
	wantStories = 42
)

// wantCompanies is the company list AC-2 names, in file order, minus
// Noodle (absent from the committed source, C-2).
//
//nolint:gochecknoglobals // Ordered baseline expectation, compared index by index.
var wantCompanies = []string{
	companyDefectDojo,
	companyPullerTech,
	"Pentalog",
	"Ubilogix Inc",
	companyAlesayi,
	"RedRabbit MX",
}

// TestBaseline pins AC-2 (§10.3) against the real committed source, not a
// copy: reading page/reydavid_experience.md directly means this fixture can
// never drift from the source. Asserts against the in-memory *Resume
// (parseDoc + mapDoc), the same path Extract takes before validation.
func TestBaseline(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(srcMD)
	if os.IsNotExist(err) {
		t.Skipf("%s absent; extract is optional", srcMD) // data/resume.yaml may be hand-maintained
	}

	if err != nil {
		t.Fatalf("reading %s: %v", srcMD, err)
	}

	resume := mapDoc(parseDoc(stripLeadingComment(string(raw))))

	if len(resume.Experience) != wantEntries {
		t.Errorf("len(Experience) = %d, want %d", len(resume.Experience), wantEntries)
	}

	if len(resume.Experience) == len(wantCompanies) {
		for i, want := range wantCompanies {
			if got := resume.Experience[i].Company; got != want {
				t.Errorf("Experience[%d].Company = %q, want %q", i, got, want)
			}
		}
	}

	assertStories(t, resume)
}

// assertStories checks the story and suggestion invariants the baseline
// guarantees, independent of the experience entries above.
func assertStories(t *testing.T, resume *Resume) {
	t.Helper()

	// Lower bound, not an equality: the markdown is hand-maintained and
	// data/resume.yaml no longer has to match it, so an added bullet is
	// not a regression.
	if len(resume.Search.Stories) < wantStories {
		t.Errorf("len(Search.Stories) = %d, want >= %d", len(resume.Search.Stories), wantStories)
	}

	for idx, story := range resume.Search.Stories {
		if story.Title == "" {
			t.Errorf("story[%d]: empty title", idx)
		}

		if len(story.Tags) == 0 {
			t.Errorf("story[%d] %q: no tags", idx, story.Title)
		}
	}

	if n := len(resume.Search.Suggestions); n < 1 || n > 8 {
		t.Errorf("len(Suggestions) = %d, want 1..8", n)
	} else if resume.Search.Suggestions[0] != tagPerformance {
		t.Errorf("Suggestions[0] = %q, want %q", resume.Search.Suggestions[0], tagPerformance)
	}
}
