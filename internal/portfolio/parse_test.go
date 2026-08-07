//nolint:testpackage // white-box: this file exercises parseDoc and classify, which are unexported.
package portfolio

import (
	"strings"
	"testing"
)

func TestClassify(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		line string
		want lineKind
	}{
		{"empty", "", kindBlank},
		{"whitespace only", "   \t  ", kindBlank},
		{"heading level 3", "### SUMMARY", kindHeading},
		{"heading level 1", "# Name — Role", kindHeading},
		{"heading trailing hashes", "## Homelab ##", kindHeading},
		{"bullet star", "* item text", kindBullet},
		{"bullet dash", "- item text", kindBullet},
		{"bullet plus indented", "  + nested item", kindBullet},
		{"bullet with bold prefix wins over boldish", "* **Advanced multithreaded performance:** delivered talks", kindBullet},
		{"italic date line", "*Apr 2024 – Present*", kindItalic},
		{"bold only", "**SKILLS**", kindBoldOnly},
		{"bold only education entry", "**B.S. Software Engineering — " + schoolITC + "**", kindBoldOnly},
		{"boldish entry line", "**Company** – Title – Location", kindBoldish},
		{"boldish story title", "**The Translator Pattern:** every security-vendor connector owns its parsing logic", kindBoldish},
		{"plain tech line", "Technologies: Go, PostgreSQL", kindPlain},
		{"plain prose", "some prose line with no markup at all", kindPlain},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := classify(testCase.line)
			if got != testCase.want {
				t.Errorf("classify(%q) = %v, want %v", testCase.line, got, testCase.want)
			}
		})
	}
}

func TestParseDoc_DateDashOrdering(t *testing.T) {
	t.Parallel()

	t.Run("defectdojo three-segment entry with italic date line", func(t *testing.T) {
		t.Parallel()

		src := "### EXPERIENCE\n\n" +
			"**DefectDojo Inc – Sr Software Engineer – Remote**\n" +
			"*Apr 2024 – Present*\n\n" +
			"* Did the work.\n\n" +
			"Technologies: Go, PostgreSQL\n"
		doc := parseDoc(src)
		entry := requireOneEntry(t, doc, SectionExperience)
		wantSegments(t, entry, companyDefectDojo, "Sr Software Engineer", locationRemote)

		if entry.Start != "Apr 2024" {
			t.Errorf("Start = %q, want %q", entry.Start, "Apr 2024")
		}

		if entry.End != datePresent {
			t.Errorf("End = %q, want %q", entry.End, datePresent)
		}
	})

	t.Run("puller tech: parenthetical discarded, single date", func(t *testing.T) {
		t.Parallel()

		src := "### EXPERIENCE\n\n" +
			"**Puller Tech – Freelance Software Consultant – Remote**\n" +
			"*Aug 2025 (1-week engagement)*\n\n" +
			"Technologies: Go, gRPC\n"
		doc := parseDoc(src)
		entry := requireOneEntry(t, doc, SectionExperience)
		wantSegments(t, entry, companyPullerTech, "Freelance Software Consultant", locationRemote)

		if entry.Start != "Aug 2025" || entry.End != "Aug 2025" {
			t.Errorf("Start/End = %q/%q, want %q/%q", entry.Start, entry.End, "Aug 2025", "Aug 2025")
		}
	})
}

func TestParseDoc_DateNormalization(t *testing.T) {
	t.Parallel()

	t.Run("education year span with em dash", func(t *testing.T) {
		t.Parallel()

		src := "### EDUCATION\n\n" +
			"**B.S. Software Engineering — " + schoolITC + "**\n" +
			"*2009 — 2014*\n"
		doc := parseDoc(src)

		entry := requireOneEntry(t, doc, SectionEducation)
		if entry.Start != "2009" || entry.End != "2014" {
			t.Errorf("Start/End = %q/%q, want %q/%q", entry.Start, entry.End, "2009", "2014")
		}
	})

	t.Run("present normalization from current and ongoing", func(t *testing.T) {
		t.Parallel()

		src := "### EXPERIENCE\n\n" +
			"**Alpha Co – Engineer – Remote**\n" +
			"*Jan 2020 – Current*\n\n" +
			"**Beta Co – Engineer – Remote**\n" +
			"*Jan 2021 – Ongoing*\n"

		doc := parseDoc(src)
		if len(doc.Sections) != 1 || len(doc.Sections[0].Entries) != 2 {
			t.Fatalf("got %d sections; want 1 section with 2 entries", len(doc.Sections))
		}

		for _, e := range doc.Sections[0].Entries {
			if e.End != datePresent {
				t.Errorf("entry %v: End = %q, want %q", e.Raw, e.End, datePresent)
			}
		}
	})
}

func TestParseDoc_Arity(t *testing.T) {
	t.Parallel()

	t.Run("arity 1: company only", func(t *testing.T) {
		t.Parallel()

		src := "### EXPERIENCE\n\n**OnlyCompany**\n*Jan 2020 – Feb 2021*\n"
		doc := parseDoc(src)
		entry := requireOneEntry(t, doc, SectionExperience)
		wantSegments(t, entry, "OnlyCompany", "", "")
	})

	t.Run("arity 2: company and title", func(t *testing.T) {
		t.Parallel()

		src := "### EXPERIENCE\n\n**Company Two – Some Title**\n*2019 – 2020*\n"
		doc := parseDoc(src)
		entry := requireOneEntry(t, doc, SectionExperience)
		wantSegments(t, entry, "Company Two", "Some Title", "")
	})

	t.Run("arity 3: company, title, location", func(t *testing.T) {
		t.Parallel()

		src := "### EXPERIENCE\n\n**A Co – A Title – A Place**\n*2019 – 2020*\n"
		doc := parseDoc(src)
		entry := requireOneEntry(t, doc, SectionExperience)
		wantSegments(t, entry, "A Co", "A Title", "A Place")
	})

	t.Run("arity 4+: middle segments rejoin with en dash and warn", func(t *testing.T) {
		t.Parallel()

		src := "### EXPERIENCE\n\n**A – B – C – D**\n*2019 – 2020*\n"
		doc := parseDoc(src)
		entry := requireOneEntry(t, doc, SectionExperience)
		wantSegments(t, entry, "A", "B – C", "D")

		if len(doc.Warnings) == 0 {
			t.Errorf("expected a warning for 4+ arity, got none")
		}
	})
}

func TestParseDoc_Identity(t *testing.T) {
	t.Parallel()

	t.Run("one-line D-25 shape", func(t *testing.T) {
		t.Parallel()

		src := "# Jane Doe — Senior Engineer\n\n### SUMMARY\n\nSome bio.\n"

		doc := parseDoc(src)
		if doc.Name != "Jane Doe" {
			t.Errorf("Name = %q, want %q", doc.Name, "Jane Doe")
		}

		if doc.RoleLine != "Senior Engineer" {
			t.Errorf("RoleLine = %q, want %q", doc.RoleLine, "Senior Engineer")
		}
	})

	t.Run("three-line shape from the real file: contact discarded, role verbatim", func(t *testing.T) {
		t.Parallel()
		// Verbatim first lines of page/reydavid_experience.md, minus the
		// leading HTML comment block.
		src := "**REY DAVID DOMÍNGUEZ SOTO**\n" +
			"Senior Go Engineer  •  Distributed Systems  •  Backend, Infrastructure & Developer Tooling\n\n" +
			"[rdominguez@tecnologer.net](mailto:rdominguez@tecnologer.net)  •  " +
			"[tecnologer.net](http://tecnologer.net)  •  Remote-first  •  " +
			"San José del Cabo, México (CST/MST)\n\n" +
			"### SUMMARY\n\nSome bio.\n"

		doc := parseDoc(src)
		if doc.Name != "REY DAVID DOMÍNGUEZ SOTO" {
			t.Errorf("Name = %q, want %q", doc.Name, "REY DAVID DOMÍNGUEZ SOTO")
		}

		wantRole := "Senior Go Engineer  •  Distributed Systems  •  Backend, Infrastructure & Developer Tooling"
		if doc.RoleLine != wantRole {
			t.Errorf("RoleLine = %q, want %q (verbatim)", doc.RoleLine, wantRole)
		}

		if strings.Contains(doc.RoleLine, "mailto:") || strings.Contains(doc.Name, "mailto:") {
			t.Errorf("contact line leaked into identity: Name=%q RoleLine=%q", doc.Name, doc.RoleLine)
		}
	})
}

func TestParseDoc_EducationBoldOnlyAsymmetry(t *testing.T) {
	t.Parallel()

	src := "### EDUCATION\n\n" +
		"**B.S. Software Engineering — " + schoolITC + "**\n" +
		"*2009 — 2014*\n"
	doc := parseDoc(src)

	if len(doc.Sections) != 1 {
		t.Fatalf("got %d sections, want 1 (unknown-keyword BoldOnly must NOT open a second section)", len(doc.Sections))
	}

	sec := doc.Sections[0]
	if sec.Kind != SectionEducation {
		t.Fatalf("section Kind = %v, want SectionEducation", sec.Kind)
	}

	if len(sec.Entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(sec.Entries))
	}

	entry := sec.Entries[0]
	if len(entry.Segments) != 2 {
		t.Fatalf("got %d segments, want 2 (institution, detail)", len(entry.Segments))
	}

	if entry.Segments[0] != schoolITC {
		t.Errorf("institution = %q, want %q", entry.Segments[0], schoolITC)
	}

	if entry.Segments[1] != "B.S. Software Engineering" {
		t.Errorf("detail = %q, want %q", entry.Segments[1], "B.S. Software Engineering")
	}
}

// --- helpers ---

func requireOneEntry(t *testing.T, doc Doc, kind SectionKind) Entry {
	t.Helper()

	for _, sec := range doc.Sections {
		if sec.Kind != kind {
			continue
		}

		if len(sec.Entries) != 1 {
			t.Fatalf("section %v: got %d entries, want 1", kind, len(sec.Entries))
		}

		return sec.Entries[0]
	}

	t.Fatalf("no section of kind %v found in %+v", kind, doc.Sections)

	return Entry{}
}

func wantSegments(t *testing.T, entry Entry, company, title, location string) {
	t.Helper()

	if len(entry.Segments) != 3 {
		t.Fatalf("got %d segments %v, want 3", len(entry.Segments), entry.Segments)
	}

	if entry.Segments[0] != company || entry.Segments[1] != title || entry.Segments[2] != location {
		t.Errorf("Segments = %v, want [%q %q %q]", entry.Segments, company, title, location)
	}
}
