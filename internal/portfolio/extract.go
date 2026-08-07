package portfolio

// Doc -> Resume mapping, per ARCHITECTURE.md §7.2 (the extract command) and
// §5.5 (story title dedup). parse.go produces shape-only structure; this
// file is the only place that knows resume semantics: story assembly, tag
// calls into tags.go, the validation gate, and the on-disk YAML shape.

import (
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"

	yaml "github.com/goccy/go-yaml"
)

// storyTitleLimit is the P-7 rune-safe truncation limit for story titles
// (FR-9 / D-27).
const storyTitleLimit = 60

// yamlHeader is the fixed FR-5 header: no timestamp, no version stamp, so
// re-running extract on unchanged input reproduces it byte for byte (FR-3,
// AC-1).
const yamlHeader = "# resume.yaml — the data `profilegen build` renders. Hand-edit it freely.\n" +
	"# `profilegen extract` is optional: it rebuilds this file from\n" +
	"# page/reydavid_experience.md and overwrites any hand edits.\n\n"

// reMDLinkLabel unwraps a markdown link to its label text. Used for the
// H-6 rule that story title/text carry no markdown link syntax through to
// the rendered JSON. Distinct from parse.go's reMDLink (§4.7), which
// captures the URL target for FR-45 hrefs, not the label.
var reMDLinkLabel = regexp.MustCompile(`\[([^\]]*)\]\(https?://[^)\s]+\)`)

// reCondenseSplit finds the first '.', ':' or spaced em dash, marking
// where the H-6 condensed-clause story title (FR-9 / D-27) is cut.
var reCondenseSplit = regexp.MustCompile(`[.:]|\s+\x{2014}\s+`)

// reLeadingComment matches a single HTML comment block at the very start
// of the source (D-30: the source-of-truth notice added per FR-31). It
// carries no resume content, so extract.go — not parse.go's shape-only
// classifier — strips it before parsing; parse.go has no HTML-comment
// rule and would otherwise read "<!--" as the identity name.
var reLeadingComment = regexp.MustCompile(`(?s)^\s*<!--.*?-->\s*`)

// stripLeadingComment removes a leading HTML comment block, if present.
func stripLeadingComment(src string) string {
	return reLeadingComment.ReplaceAllString(src, "")
}

// Extract reads inPath, parses it, maps the result to a Resume, validates
// it, and writes outPath. On any validation failure it writes nothing
// (AD-10, FR-27): the previous outPath, if any, is left untouched.
func Extract(inPath, outPath string) error {
	raw, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", inPath, err)
	}

	doc := parseDoc(stripLeadingComment(string(raw)))
	resume := mapDoc(doc)

	if failMsgs := validate(resume); len(failMsgs) > 0 {
		for _, msg := range failMsgs {
			fmt.Fprintf(os.Stderr, "extract: validation failed: %s\n", msg)
		}

		return fmt.Errorf("validation failed: %s", strings.Join(failMsgs, "; "))
	}

	out, err := yaml.Marshal(resume)
	if err != nil {
		return fmt.Errorf("marshalling resume: %w", err)
	}

	if err := os.WriteFile(outPath, append([]byte(yamlHeader), out...), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	fmt.Fprintf(os.Stderr, "extract: %d entries, %d stories, %d suggestions, %d warnings → %s\n",
		len(resume.Experience), len(resume.Search.Stories), len(resume.Search.Suggestions), len(doc.Warnings), outPath)

	return nil
}

// ---------------------------------------------------------------------
// Doc -> Resume mapping
// ---------------------------------------------------------------------

// mapDoc walks doc.Sections in document order and builds the Resume,
// calling into tags.go for every inference step (AD-06). Story order and
// the title-dedup map (§5.5) both depend on this being a single
// document-order pass: never iterate a map to produce output.
func mapDoc(doc Doc) *Resume {
	resume := &Resume{
		Profile: Profile{Name: doc.Name, RoleLine: doc.RoleLine},
	}

	titleCounts := make(map[string]int)

	var (
		storyTagLists [][]string
		sourceSlugs   []string
	)

	addStory := func(bulletText, sourceSlug string) {
		st := buildStory(bulletText, sourceSlug)
		st.Title = dedupTitle(titleCounts, st.Title, sourceSlug)
		resume.Search.Stories = append(resume.Search.Stories, st)
		storyTagLists = append(storyTagLists, st.Tags)
		sourceSlugs = append(sourceSlugs, sourceSlug)
	}

	for _, sec := range doc.Sections {
		mapSection(resume, sec, addStory)
	}

	resume.Search.Suggestions = suggestions(storyTagLists, sourceSlugs)

	return resume
}

// mapSection folds one parsed Section into resume, routing its bullets to
// addStory where the section contributes stories.
func mapSection(resume *Resume, sec Section, addStory func(bulletText, sourceSlug string)) {
	switch sec.Kind {
	case SectionSummary:
		resume.Profile.BioHTML = renderBioHTML(sec.Prose)

	case SectionExperience:
		resume.Experience = append(resume.Experience, mapExperience(sec.Entries, addStory)...)

	case SectionEducation:
		resume.Education = append(resume.Education, mapEducation(sec.Entries)...)

	case SectionCertifications:
		for _, b := range sec.Bullets {
			resume.Certifications = append(resume.Certifications, b.Text)
		}

	case SectionSpeaking, SectionOpenSource, SectionHomelab,
		SectionSkills, SectionLanguages, SectionUnknown:
		mapStorySection(sec, addStory)
	}
}

// mapStorySection handles the bullets-only sections that contribute
// stories but no structured resume field.
func mapStorySection(sec Section, addStory func(bulletText, sourceSlug string)) {
	switch sec.Kind {
	case SectionSpeaking, SectionOpenSource:
		for _, b := range sec.Bullets {
			addStory(b.Text, sec.Slug)
		}

	case SectionHomelab:
		for _, b := range sec.Bullets {
			addStory(b.Text, homelabSlug(b.Text))
		}

	case SectionSummary, SectionExperience, SectionEducation, SectionCertifications,
		SectionSkills, SectionLanguages, SectionUnknown:
		// Parsed without error, no render target (FR-7, D-26).
	}
}

// mapExperience maps EXPERIENCE entries in document order, feeding each
// entry's bullets to addStory under the company's slug.
func mapExperience(entries []Entry, addStory func(bulletText, sourceSlug string)) []Experience {
	out := make([]Experience, 0, len(entries))

	for _, entry := range entries {
		company, title, location := entrySegment(entry, 0), entrySegment(entry, 1), entrySegment(entry, 2)
		out = append(out, Experience{
			Company:      company,
			Title:        title,
			StartDate:    entry.Start,
			EndDate:      titleCase(entry.End),
			Location:     location,
			Roles:        inferRoles(title),
			Technologies: entryTechnologies(entry),
		})

		srcSlug := slug(company)
		for _, b := range entry.Bullets {
			addStory(b.Text, srcSlug)
		}
	}

	return out
}

// mapEducation maps EDUCATION entries in document order. Education carries
// no bullets, so it contributes no stories.
func mapEducation(entries []Entry) []Education {
	out := make([]Education, 0, len(entries))

	for _, e := range entries {
		out = append(out, Education{
			Institution: entrySegment(e, 0),
			Detail:      entrySegment(e, 1),
			Years:       formatYears(e.Start, e.End),
		})
	}

	return out
}

// entrySegment safely reads Entry.Segments[i], returning "" past the end
// (arity-tolerant: an Experience entry with fewer than 3 segments, or an
// Education entry with fewer than 2, still maps cleanly).
func entrySegment(e Entry, i int) string {
	if i < len(e.Segments) {
		return e.Segments[i]
	}

	return ""
}

// formatYears renders an education Years string ("2009 — 2014", or a
// single "Aug 2025" when start == end, or "" when no date was found — the
// EDUCATION date exception, §4.6).
func formatYears(start, end string) string {
	if start == "" {
		return ""
	}

	end = titleCase(end)
	if end == "" || end == start {
		return start
	}

	return start + " — " + end
}

// renderBioHTML implements FR-44: SUMMARY prose becomes profile.bio_html,
// **bold** converted to <strong>, everything else HTML-escaped.
func renderBioHTML(prose []string) string {
	text := strings.TrimSpace(strings.Join(prose, " "))
	if text == "" {
		return ""
	}

	var builder strings.Builder

	last := 0
	for _, loc := range reBoldInline.FindAllStringSubmatchIndex(text, -1) {
		builder.WriteString(html.EscapeString(text[last:loc[0]]))
		builder.WriteString("<strong>")
		builder.WriteString(html.EscapeString(text[loc[2]:loc[3]]))
		builder.WriteString("</strong>")

		last = loc[1]
	}

	builder.WriteString(html.EscapeString(text[last:]))

	return builder.String()
}

// ---------------------------------------------------------------------
// §5.5 Story assembly and title dedup
// ---------------------------------------------------------------------

// buildStory turns one achievement bullet into a Story (FR-9). Title and
// story text are computed from the markdown-link-unwrapped bullet (H-6);
// href and tags are computed from the raw bullet, since FR-45's href needs
// the URL that unwrapping discards.
func buildStory(bulletText, sourceSlug string) Story {
	clean := unwrapLinks(bulletText)

	var title, story string

	switch {
	case reBoldTitleInner.MatchString(clean):
		m := reBoldTitleInner.FindStringSubmatch(clean)
		title, story = strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
	case reBoldTitleOuter.MatchString(clean):
		m := reBoldTitleOuter.FindStringSubmatch(clean)
		title, story = strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
	default:
		story = clean
		title = condensedTitle(clean)
	}

	// A bold link label (`[**Tempura**](url)`) survives unwrapLinks with
	// its ** markers intact, since unwrapLinks only discards the link
	// syntax, not the label's own emphasis. Story/title text is rendered
	// as escaped plain text (template/index.html.tmpl's esc()), which
	// does not interpret markdown, so any leftover ** would show up
	// literally on the page. Stripped here, after the H-6 bold-prefix
	// match above, so it never interferes with reBoldTitleInner/Outer.
	title = reBoldSpan.ReplaceAllString(title, "$1")
	story = reBoldSpan.ReplaceAllString(story, "$1")

	return Story{
		Title:  truncate(title, storyTitleLimit),
		Story:  story,
		Source: displaySource(sourceSlug),
		Href:   storyHref(bulletText),
		Tags:   storyTags(bulletText, sourceSlug),
	}
}

// condensedTitle implements the FR-9 / D-27 fallback: bullet text up to
// the first '.', ':' or spaced em dash, trimmed.
func condensedTitle(text string) string {
	text = strings.TrimSpace(text)
	if loc := reCondenseSplit.FindStringIndex(text); loc != nil {
		text = text[:loc[0]]
	}

	return strings.TrimSpace(text)
}

// unwrapLinks replaces every markdown link with its label text (H-6).
func unwrapLinks(s string) string {
	return reMDLinkLabel.ReplaceAllString(s, "$1")
}

// displaySource renders a source slug for the Story.Source field: hyphens
// become spaces (§5.3), matching today's `source: puller tech`. A no-op
// for slugs that already carry spaces (community, open source).
func displaySource(slug string) string {
	return strings.ReplaceAll(slug, "-", " ")
}

// storyHref implements FR-45: the first URL in the bullet prose (markdown
// link target or bare URL), right-trimmed of trailing punctuation per
// P-6, defaulting to "#experience". "First" means first by position in
// the raw text, so a bare URL preceding a markdown link still wins.
func storyHref(text string) string {
	mdLoc := reMDLink.FindStringSubmatchIndex(text)
	bareLoc := reBareURL.FindStringIndex(text)

	var url string

	switch {
	case mdLoc == nil && bareLoc == nil:
		return "#experience"
	case bareLoc == nil || (mdLoc != nil && mdLoc[0] <= bareLoc[0]):
		url = text[mdLoc[2]:mdLoc[3]]
	default:
		url = text[bareLoc[0]:bareLoc[1]]
	}

	return strings.TrimRight(url, ".,;:!?)]")
}

// dedupTitle implements §5.5: a single map[string]int, keyed by lowercased
// title text, used only for lookup and counting (never iterated — FR-3).
// First occurrence of a title passes through unchanged; the second gets
// "Title (source-slug)" appended; further collisions of that composite
// get an ordinal, "Title (source-slug 2)", "Title (source-slug 3)", ...
func dedupTitle(counts map[string]int, title, sourceSlug string) string {
	key := strings.ToLower(title)

	counts[key]++
	if counts[key] == 1 {
		return title
	}

	composite := title + " (" + sourceSlug + ")"
	compositeKey := strings.ToLower(composite)

	counts[compositeKey]++
	if counts[compositeKey] == 1 {
		return composite
	}

	return fmt.Sprintf("%s (%s %d)", title, sourceSlug, counts[compositeKey])
}

// ---------------------------------------------------------------------
// §7.2 Validation gate (V1..V5)
// ---------------------------------------------------------------------

// validate runs all five checks (never short-circuiting) and returns one
// "<check name>: <detail>" message per failing check, in V1..V5 order.
// nil/empty means the Resume is valid.
func validate(resume *Resume) []string {
	var fails []string

	if resume.Profile.Name == "" {
		fails = append(fails, "V1: profile.name is empty")
	}

	if len(resume.Experience) < 1 {
		fails = append(fails, fmt.Sprintf("V2: len(experience)=%d, need at least 1", len(resume.Experience)))
	}

	if len(resume.Search.Stories) < 1 {
		fails = append(fails, fmt.Sprintf("V3: len(search.stories)=%d, need at least 1", len(resume.Search.Stories)))
	}

	if !storiesValid(resume.Search.Stories) {
		fails = append(fails, "V4: every story needs a non-empty title and at least one tag")
	}

	if n := len(resume.Search.Suggestions); n < 1 || n > 8 {
		fails = append(fails, fmt.Sprintf("V5: search.suggestions has %d entries, need 1-8", n))
	}

	return fails
}

func storiesValid(stories []Story) bool {
	for _, s := range stories {
		if s.Title == "" || len(s.Tags) == 0 {
			return false
		}
	}

	return true
}

// ---------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------

// titleCase normalizes an endpoint token to exactly datePresent. parse.go's
// tryEntry already applies this during date-range extraction (AD-11), so
// by the time a Doc reaches this file every End is already normalized;
// kept here as the defensive, idempotent normalizer on the way into the
// model (P-5).
func titleCase(text string) string {
	lower := strings.ToLower(text)
	if lower == "present" || lower == "current" || lower == "now" {
		return datePresent
	}

	return text
}

// truncate is the P-7 rune-safe truncation: strings under the limit pass
// through unchanged; otherwise cut at the last space at or below limit
// (or at limit itself if there is none) and append "…". Never splits a
// multi-byte rune.
func truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}

	cut := limit
	for cut > 0 && runes[cut] != ' ' {
		cut--
	}

	if cut == 0 {
		cut = limit
	}

	return string(runes[:cut]) + "…"
}
