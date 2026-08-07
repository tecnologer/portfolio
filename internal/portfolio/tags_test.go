//nolint:testpackage // white-box: this file exercises matchTechTags, inferThemes, inferRoles and slug, which are unexported.
package portfolio

import (
	"slices"
	"testing"
)

func TestMatchTechTags_ModeExactGo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want bool
	}{
		{"capitalized Go language matches", "Built a Go microservice", true},
		{"go live is the English verb, not the language", "we need to go live tomorrow", false},
		{"go from 1K to 40K is the English verb too", "scaled go from 1K to 40K", false},
		{"lowercase go inside a word does not match", "mango season", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Contains(matchTechTags(test.text), "go")
			if got != test.want {
				t.Errorf("matchTechTags(%q) contains %q = %v, want %v", test.text, "go", got, test.want)
			}
		})
	}
}

func TestMatchTechTags_ModeSymbol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		tag  string
		want bool
	}{
		{"lowercase dotnet matches", "built on .net", tagDotnet, true},
		{"uppercase DOTNET matches", "migrated to .NET", tagDotnet, true},
		{"C# matches", "wrote it in C#", "csharp", true},
		{"C++ matches", "legacy C++ codebase", "cpp", true},
		{".NETWORK must not match: right boundary is a word char", "the .NETWORK protocol", tagDotnet, false},
		{"asp.netcore adjacency must not match: left boundary is a word char", "built on asp.netcore", tagDotnet, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Contains(matchTechTags(test.text), test.tag)
			if got != test.want {
				t.Errorf("matchTechTags(%q) contains %q = %v, want %v", test.text, test.tag, got, test.want)
			}
		})
	}
}

func TestInferThemes_PerformanceNx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want bool
	}{
		{"plain 40x with no other performance keyword", "shipped a 40x improvement", true},
		{"en-dash range 10x-40x", "performance profiling & optimization (10x–40x gains)", true},
		{"no Nx and no keyword: not tagged", "shipped a new feature", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Contains(inferThemes(test.text), tagPerformance)
			if got != test.want {
				t.Errorf("inferThemes(%q) contains %q = %v, want %v", test.text, tagPerformance, got, test.want)
			}
		})
	}
}

func TestInferThemes_AIModeExact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want bool
	}{
		{"AI upper case matches", "built an AI agent", true},
		{"lowercase ai inside a word does not match", "an air-gapped build", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Contains(inferThemes(test.text), "ai")
			if got != test.want {
				t.Errorf("inferThemes(%q) contains %q = %v, want %v", test.text, "ai", got, test.want)
			}
		})
	}
}

func TestInferRoles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		want  []string
	}{
		{"architect and full-stack in declaration order", "Software Architect / Full Stack Developer", []string{tagArchitect, tagFullStack}},
		{"back-end from software engineer", "Senior Software Engineer", []string{tagBackEnd}},
		{"bare manager keyword still matches people manager", "Office Manager of Snacks", []string{tagPeopleManager}},
		{"title matching nothing yields no roles", "Chief Wizard", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := inferRoles(test.title)
			if !slices.Equal(got, test.want) {
				t.Errorf("inferRoles(%q) = %v, want %v", test.title, got, test.want)
			}
		})
	}
}

func TestEntryTechnologies(t *testing.T) {
	t.Parallel()

	t.Run("verbatim H-7 line wins, order and casing preserved", func(t *testing.T) {
		t.Parallel()

		entry := Entry{
			Segments: []string{sourceAcmeTitle, "Go Engineer", locationRemote},
			Techs:    []string{tagPostgreSQL, "Go", "REDIS"},
		}

		got := entryTechnologies(entry)
		if !slices.Equal(got, entry.Techs) {
			t.Errorf("entryTechnologies() = %v, want verbatim %v", got, entry.Techs)
		}
	})

	t.Run("inferred from title and bullets when no Technologies line", func(t *testing.T) {
		t.Parallel()

		entry := Entry{
			Segments: []string{sourceAcmeTitle, "Go Engineer", locationRemote},
			Bullets: []Bullet{
				{Text: "Rebuilt the API on PostgreSQL and Docker."},
			},
		}
		got := entryTechnologies(entry)

		want := []string{"Go", displayPostgreSQL, displayDocker}
		if !slices.Equal(got, want) {
			t.Errorf("entryTechnologies() = %v, want %v", got, want)
		}
	})

	t.Run("no matches yields an empty list", func(t *testing.T) {
		t.Parallel()

		entry := Entry{
			Segments: []string{sourceAcmeTitle, "Office Manager", locationRemote},
			Bullets:  []Bullet{{Text: "Kept the snack drawer stocked."}},
		}

		got := entryTechnologies(entry)
		if len(got) != 0 {
			t.Errorf("entryTechnologies() = %v, want empty", got)
		}
	})
}

func TestSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{companyDefectDojo, companyDefectDojo, "defectdojo"},
		{companyPullerTech, companyPullerTech, "puller-tech"},
		{companyAlesayi, companyAlesayi, "alesayi-adco"},
		{"trailing Co. stop-word with a period", "Widgets Co.", "widgets"},
		{"s.a. stop-word", "Acme S.A.", sourceAcme},
		{"llc and ltd stop-words", "Acme LLC", sourceAcme},
		{"mx stop-word", "Acme MX", sourceAcme},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := slug(test.in)
			if got != test.want {
				t.Errorf("slug(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}

func TestHomelabSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		bullet string
		want   string
	}{
		{
			name:   "linked bold proper noun",
			bullet: "Built [**Tempura**](https://github.com/tecnologer/tempura), an end-to-end IoT monitor.",
			want:   "tempura",
		},
		{
			name:   "plain bold proper noun, no link",
			bullet: "Built **Spoolman Fork** for bulk filament loading.",
			want:   "spoolman-fork",
		},
		{
			name:   "no bold or linked noun falls back to homelab",
			bullet: "Linux daily driver since ~2010 across several distros.",
			want:   tagHomelab,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := homelabSlug(test.bullet)
			if got != test.want {
				t.Errorf("homelabSlug(%q) = %q, want %q", test.bullet, got, test.want)
			}
		})
	}
}

func TestStoryTags(t *testing.T) {
	t.Parallel()

	t.Run("assembly order: themes, then techs, then source slug", func(t *testing.T) {
		t.Parallel()

		text := "Rewrote the connector for 40x throughput on Go and PostgreSQL."
		got := storyTags(text, "puller-tech")

		want := []string{tagPerformance, tagMigration, "integrations", "go", tagPostgreSQL, "puller-tech"}
		if !slices.Equal(got, want) {
			t.Errorf("storyTags() = %v, want %v", got, want)
		}
	})

	t.Run("source slug guarantees at least one tag", func(t *testing.T) {
		t.Parallel()

		got := storyTags("nothing notable here", sourceAcme)
		if len(got) == 0 {
			t.Fatal("storyTags() returned no tags, want at least the source slug")
		}

		if got[len(got)-1] != sourceAcme {
			t.Errorf("storyTags() = %v, want last element %q", got, sourceAcme)
		}
	})

	t.Run("dedup keeps first occurrence when the slug repeats a tech tag", func(t *testing.T) {
		t.Parallel()

		got := storyTags("Built with Go", "go")

		want := []string{"go"}
		if !slices.Equal(got, want) {
			t.Errorf("storyTags() = %v, want %v", got, want)
		}
	})
}

func TestSuggestions(t *testing.T) {
	t.Parallel()

	t.Run("performance pinned at index 0 even at zero frequency", func(t *testing.T) {
		t.Parallel()

		got := suggestions([][]string{{"go", sourceAcme}}, []string{sourceAcme})
		if len(got) == 0 || got[0] != tagPerformance {
			t.Fatalf("suggestions()[0] = %v, want %q first", got, tagPerformance)
		}
	})

	t.Run("source slugs are excluded from candidacy", func(t *testing.T) {
		t.Parallel()

		got := suggestions([][]string{{"go", sourceAcme}}, []string{sourceAcme})
		if slices.Contains(got, sourceAcme) {
			t.Errorf("suggestions() = %v, must not contain the source slug %q", got, sourceAcme)
		}
	})

	t.Run("frequency descending, ties broken by tag ascending byte order", func(t *testing.T) {
		t.Parallel()

		storyTags := [][]string{
			{tagPerformance, "zeta", sourceAlpha, sourceOne},
			{"beta", sourceAlpha, "source2"},
		}
		got := suggestions(storyTags, []string{sourceOne, "source2"})

		want := []string{tagPerformance, sourceAlpha, "beta", "zeta"}
		if !slices.Equal(got, want) {
			t.Errorf("suggestions() = %v, want %v", got, want)
		}
	})
}

func TestSuggestionsBounds(t *testing.T) {
	t.Parallel()

	t.Run("length is 1..8: performance plus at most the first 7 candidates", func(t *testing.T) {
		t.Parallel()

		storyTags := [][]string{
			{"t1", "t2", "t3", "t4", "t5", "t6", "t7", "t8", "t9", sourceOne},
		}

		got := suggestions(storyTags, []string{sourceOne})
		if len(got) != 8 {
			t.Fatalf("len(suggestions()) = %d, want 8", len(got))
		}

		if got[0] != tagPerformance {
			t.Errorf("suggestions()[0] = %q, want %q", got[0], tagPerformance)
		}
	})

	t.Run("empty input still yields performance alone", func(t *testing.T) {
		t.Parallel()

		got := suggestions(nil, nil)

		want := []string{tagPerformance}
		if !slices.Equal(got, want) {
			t.Errorf("suggestions(nil, nil) = %v, want %v", got, want)
		}
	})
}
