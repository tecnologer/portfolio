package portfolio

import (
	"fmt"
	"os"
	"strings"

	yaml "github.com/goccy/go-yaml"
)

// Resume is the machine-generated data model (data/resume.yaml). It is
// produced by `profilegen extract` from page/reydavid_experience.md and
// consumed by `profilegen build`, which renders it into the landing page.
type Resume struct {
	Profile  Profile   `yaml:"profile"`
	Contacts []Contact `yaml:"contacts"`

	Experience     []Experience `yaml:"experience"`
	Education      []Education  `yaml:"education"`
	Certifications []string     `yaml:"certifications"`
	Search         Search       `yaml:"search"`
}

// Profile is the owner's identity and the hero copy.
type Profile struct {
	Name     string `yaml:"name"`
	RoleLine string `yaml:"role_line"`
	// HeroBioHTML is the short hero copy; BioHTML the full bio. Both may
	// contain <strong> emphasis.
	HeroBioHTML string `yaml:"hero_bio_html"`
	BioHTML     string `yaml:"bio_html"`
}

// HeroBio is the hero paragraph: the short copy when set, the full bio
// otherwise (extract only ever writes the latter).
func (p Profile) HeroBio() string {
	if p.HeroBioHTML != "" {
		return p.HeroBioHTML
	}

	return p.BioHTML
}

// Contact is one link in the contact section. A contact with a CTA is
// also rendered as a button in the hero row, the first one as primary.
type Contact struct {
	Label string `yaml:"label"`
	Href  string `yaml:"href"`
	CTA   string `yaml:"cta"`
}

// External reports whether the link leaves the site, i.e. needs
// target="_blank". mailto:/tel: links do not.
func (c Contact) External() bool { return strings.HasPrefix(c.Href, "http") }

// Experience is a work history entry.
type Experience struct {
	Company      string   `yaml:"company"`
	Title        string   `yaml:"title"`
	StartDate    string   `yaml:"start_date"`
	EndDate      string   `yaml:"end_date"` // datePresent marks the current role
	Location     string   `yaml:"location"`
	Roles        []string `yaml:"roles"`
	Technologies []string `yaml:"technologies"`
}

// Current reports whether this is the ongoing role.
func (e Experience) Current() bool { return e.EndDate == datePresent }

// Education is an academic entry.
type Education struct {
	Institution string `yaml:"institution"`
	Detail      string `yaml:"detail"` // "Engineer's degree · Computer Software Engineering"
	Years       string `yaml:"years"`  // "2009 — 2014"
}

// Search powers the career-search command palette.
type Search struct {
	Suggestions []string `yaml:"suggestions"`
	Stories     []Story  `yaml:"stories"`
}

// Story is one searchable challenge → action → impact entry.
type Story struct {
	Title  string   `yaml:"title" json:"title"`
	Story  string   `yaml:"story" json:"story"`
	Source string   `yaml:"source" json:"source"` // short slug shown on the right (company/project)
	Href   string   `yaml:"href" json:"href"`     // "#experience" or an external URL
	Tags   []string `yaml:"tags" json:"tags"`
}

// errEmptyData tells the reader how to fill in a missing or blank
// data file: by hand, or with the optional extract command.
func errEmptyData(path string) error {
	return fmt.Errorf("%s has no resume data — write it by hand, or run `profilegen extract` to generate it from page/%s", path, mdFileName)
}

// LoadResume reads and validates the YAML source of truth.
func LoadResume(path string) (*Resume, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, errEmptyData(path)
	}

	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var resume Resume
	if err := yaml.Unmarshal(raw, &resume); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if resume.Profile.Name == "" {
		return nil, errEmptyData(path)
	}

	for i, s := range resume.Search.Stories {
		if s.Title == "" || len(s.Tags) == 0 {
			return nil, fmt.Errorf("%s: search.stories[%d] needs a title and at least one tag", path, i)
		}
	}

	return &resume, nil
}
