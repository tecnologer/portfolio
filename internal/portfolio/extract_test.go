//nolint:testpackage // white-box: this file exercises mapDoc/parseDoc, which is unexported.
package portfolio

import (
	"bytes"
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExtractDeterministic pins AC-1/FR-3 (§10.2): the same input produces
// byte-identical output on every run, and the output carries no TODO
// placeholder (FR-2).
func TestExtractDeterministic(t *testing.T) {
	t.Parallel()

	out1 := filepath.Join(t.TempDir(), "resume.yaml")
	out2 := filepath.Join(t.TempDir(), "resume.yaml")

	if err := Extract(srcMD, out1); err != nil {
		t.Fatalf("first Extract: %v", err)
	}

	if err := Extract(srcMD, out2); err != nil {
		t.Fatalf("second Extract: %v", err)
	}

	firstRun, err := os.ReadFile(out1)
	if err != nil {
		t.Fatalf("reading %s: %v", out1, err)
	}

	secondRun, err := os.ReadFile(out2)
	if err != nil {
		t.Fatalf("reading %s: %v", out2, err)
	}

	if !bytes.Equal(firstRun, secondRun) {
		t.Errorf("Extract is not deterministic: two runs on the same input produced different output")
	}

	if strings.Contains(string(firstRun), "TODO") {
		t.Errorf("output contains TODO, want none (FR-2)")
	}
}

// TestStoryJSONKeys pins FR-35/DR-4 (§10.4): a Go field rename must not
// change the json: tag reaching the browser. Render the template against
// a synthetic story and assert the emitted STORIES JSON carries exactly
// the five frozen keys.
func TestStoryJSONKeys(t *testing.T) {
	t.Parallel()

	stories := renderStoriesJSON(t, oneStoryResume())
	if len(stories) != 1 {
		t.Fatalf("got %d stories in STORIES, want 1", len(stories))
	}

	wantKeys := []string{"title", "story", "source", "href", "tags"}

	got := stories[0]
	if len(got) != len(wantKeys) {
		t.Errorf("story object has %d keys %v, want exactly %v", len(got), keysOf(got), wantKeys)
	}

	for _, k := range wantKeys {
		if _, ok := got[k]; !ok {
			t.Errorf("story JSON missing key %q, got keys %v", k, keysOf(got))
		}
	}
}

// TestBuildStoryStripsBoldMarkup pins the fix for a bold markdown link
// label (e.g. `[**Tempura**](url)`) leaking literal ** markers into the
// rendered story: unwrapLinks discards the link syntax but not the
// label's own emphasis, and template/index.html.tmpl's esc() renders
// story/title as plain text, so any leftover ** would show up literally
// on the page.
func TestBuildStoryStripsBoldMarkup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		bulletText string
		wantTitle  string
		wantStory  string
	}{
		{
			name:       "bold_link_label",
			bulletText: "Built [**Tempura**](https://github.com/tecnologer/tempura), an IoT monitor.",
			wantTitle:  "Built Tempura, an IoT monitor",
			wantStory:  "Built Tempura, an IoT monitor.",
		},
		{
			name:       "bold_prefix_title_unaffected",
			bulletText: "**The Translator Pattern:** every connector owns its parsing logic.",
			wantTitle:  "The Translator Pattern",
			wantStory:  "every connector owns its parsing logic.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := buildStory(test.bulletText, "example")
			if got.Title != test.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, test.wantTitle)
			}

			if got.Story != test.wantStory {
				t.Errorf("Story = %q, want %q", got.Story, test.wantStory)
			}

			if strings.Contains(got.Title, "*") || strings.Contains(got.Story, "*") {
				t.Errorf("story leaked markdown bold markers: title=%q story=%q", got.Title, got.Story)
			}
		})
	}
}

// TestExtractMalformedLeavesOutputIntact pins AC-9/FR-27 (§10.6): a
// validation failure must not touch a pre-existing output file.
func TestExtractMalformedLeavesOutputIntact(t *testing.T) {
	t.Parallel()

	// testdata/malformed/no-experience/in.md: the EXPERIENCE heading
	// renamed to something unrecognized. extract still parses (H-8, no
	// crash) but yields zero experience entries and zero stories, so
	// V2/V3 must fail the validation gate.
	const mdPath = "testdata/malformed/no-experience/in.md"

	dir := t.TempDir()

	before, err := os.ReadFile(dataYAML)
	if err != nil {
		t.Fatalf("reading %s: %v", dataYAML, err)
	}

	outPath := filepath.Join(dir, "resume.yaml")
	//nolint:gosec // outPath is filepath.Join(t.TempDir(), ...), not attacker-controlled.
	if err := os.WriteFile(outPath, before, 0o644); err != nil {
		t.Fatalf("seeding pre-existing output: %v", err)
	}

	err = Extract(mdPath, outPath)
	if err == nil {
		t.Fatal("Extract succeeded on malformed input, want a validation error")
	}

	if !strings.Contains(err.Error(), "V2") {
		t.Errorf("error = %q, want it to name the failed check (V2)", err)
	}

	after, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading %s after failed Extract: %v", outPath, err)
	}

	if !bytes.Equal(before, after) {
		t.Errorf("output file was modified despite the validation failure")
	}
}

// --- helpers ---

// oneStoryResume is the minimal Resume the STORIES JSON shape is asserted
// against: one story carrying every key the template emits.
func oneStoryResume() *Resume {
	return &Resume{
		Profile: Profile{
			Name:     "Test Person",
			RoleLine: "Role Line",
			BioHTML:  "Bio",
		},
		Search: Search{
			Suggestions: []string{tagPerformance},
			Stories: []Story{
				{
					Title:  "A Story",
					Story:  "Did a thing.",
					Source: sourceAcme,
					Href:   "#experience",
					Tags:   []string{tagPerformance},
				},
			},
		},
	}
}

// renderStoriesJSON renders the page template and decodes the `const
// STORIES = [...]` array it embeds.
func renderStoriesJSON(t *testing.T, resume *Resume) []map[string]json.RawMessage {
	t.Helper()

	tmpl, err := template.New("index.html.tmpl").Funcs(templateFuncs()).ParseFiles(tmplPath)
	if err != nil {
		t.Fatalf("parsing template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, PageData{resume, "vTest", "reydavid_experience.md"}); err != nil {
		t.Fatalf("executing template: %v", err)
	}

	const marker = "const STORIES = "

	_, rest, found := strings.Cut(buf.String(), marker)
	if !found {
		t.Fatal("STORIES assignment not found in rendered output")
	}

	end := strings.IndexByte(rest, ';')
	if end < 0 {
		t.Fatal("STORIES assignment has no terminating ';'")
	}

	var stories []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(rest[:end]), &stories); err != nil {
		t.Fatalf("unmarshalling STORIES JSON: %v", err)
	}

	return stories
}

func keysOf(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	return out
}
