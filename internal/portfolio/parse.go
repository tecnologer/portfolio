package portfolio

// Parser for page/reydavid_experience.md. Per ARCHITECTURE.md §4: a line
// classifier plus a section state machine, producing a Doc that knows
// markdown shape only. Doc -> Resume mapping happens in extract.go (not
// implemented here, and not wired to this file yet).
//
// The parser never returns a fatal error for content (H-8): every rule
// failure appends a Warning and continues. Only I/O errors would be fatal,
// and this file does none - parseDoc takes the source as a string.

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ---------------------------------------------------------------------
// §4.7 Regex inventory. Normative: these blocks are copied verbatim from
// ARCHITECTURE.md. Do not rename, do not "improve" a pattern.
// ---------------------------------------------------------------------

const dash = `[-\x{2013}\x{2014}]` // hyphen, en dash, em dash
const monthPat = `(?:jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)[a-z]*\.?`
const datePat = `(?:` + monthPat + `\s+)?\d{4}` // "Apr 2024" or "2014"
const openPat = `present|current|now|to date|ongoing`

// Matches a markdown heading at any level. Must NOT match `#experience` (no
// space after the hashes) or a `--- ` rule.
var reHeading = regexp.MustCompile(`^\s{0,3}(#{1,6})\s+(.+?)\s*#*\s*$`)

// Trims emphasis runs from the ends of a heading or entry text. Applied as
// a replace-with-empty on both anchors. Must NOT touch interior
// underscores, so `Pop!_OS` survives.
var reTrimEmph = regexp.MustCompile(`^[*_\s]+|[*_\s]+$`)

// Matches a line that is one bold span and nothing else, e.g. `**SKILLS**`
// or a bold-only education entry naming a school.
// Must NOT match `**Company** – Title – Location` (trailing text) or
// `**The Translator Pattern:** …`.
var reBoldOnly = regexp.MustCompile(`^\s*\*\*([^*]+)\*\*\s*$`)

// Matches any bold span anywhere in a line: the partially-bold entry-line
// test of H-3.
var reBoldSpan = regexp.MustCompile(`\*\*([^*]+)\*\*`)

// Matches a wholly italic line, the classic date line
// `*Apr 2024 – Present*`. The `[^*]` guard is what stops it swallowing
// bold lines.
var reItalic = regexp.MustCompile(`^\s*\*([^*].*?)\*\s*$`)

// Matches a date or date range anywhere in the text. Run this before any
// dash splitting (AD-11). Group 1 is the start, group 2 the end (empty
// when the entry carries a single date). Must match `Apr 2024 – Present`,
// `Mar 2017 - Dec 2019`, `2009 – 2014`, `Aug 2025`. Must NOT be run against
// bullet prose, where four-digit years appear as ordinary numbers.
var reDateRange = regexp.MustCompile(
	`(?i)\b(` + datePat + `)(?:\s*` + dash + `+\s*(` + datePat + `|` + openPat + `))?\b`)

// Matches a trailing parenthetical to be discarded after the date span is
// stripped: `Aug 2025 (1-week engagement)` -> the parenthetical. Must NOT
// match a parenthetical in the middle of a title.
var reTrailingParen = regexp.MustCompile(`\s*\([^)]*\)\s*$`)

// Splits an entry-line remainder into segments. Spaced dashes only. Must
// split `DefectDojo Inc – Sr Software Engineer – Remote` into three. Must
// NOT split `open-source`, `C#-based`, `C#→Go`, `10x–40x`.
var reDashSplit = regexp.MustCompile(`\s+` + dash + `+\s+`)

// Matches a bullet with any marker and any indentation. Group 1 is the
// indent (used only to detect nesting, then discarded), group 2 the text.
// Must NOT match `---`, `***`, or a bold line.
var reBullet = regexp.MustCompile(`^(\s*)[-*+]\s+(\S.*)$`)

// The two H-6 bold-prefix story-title shapes, tried in this order. Must
// match `**The Translator Pattern:** every security-vendor connector …`
// and `**Advanced multithreaded performance:** delivered talks …`. Must
// NOT match `Built [**Tempura**](https://…), an end-to-end IoT monitor …`
// (the bold span is not at position 0).
var reBoldTitleInner = regexp.MustCompile(`^\*\*([^*]+?)\s*:\s*\*\*\s*(\S.*)$`)
var reBoldTitleOuter = regexp.MustCompile(`^\*\*([^*]+?)\*\*\s*:\s*(\S.*)$`)

// Matches an optional `Technologies:` line (H-7), case-insensitive on the
// key, with an optional bullet marker and optional bold. Group 1 is the
// comma-separated value.
var reTechLine = regexp.MustCompile(`(?i)^\s*(?:[-*+]\s+)?(?:\*\*)?technologies(?:\*\*)?\s*:\s*(\S.*?)\s*$`)

// Marks a line as contact-bearing, so H-2 discards it. Must match
// `[rdominguez@tecnologer.net](mailto:…) • [tecnologer.net](http://tecnologer.net) • …`.
// Must NOT match a plain role line.
var reContactish = regexp.MustCompile(`(?i)\]\(|mailto:|https?://|[\w.+%-]+@[\w-]+\.[\w.]+`)

// Marks a line as a role-line candidate (H-2 rule 4): it contains at least
// one of the separator characters.
var reSepLine = regexp.MustCompile(`[\x{2022}\x{00b7}\x{2014}\x{2013}]`)

// Splits the D-25 one-line identity shape `# Name — Role Line`. Split
// once, on the first spaced separator.
var reNameSplit = regexp.MustCompile(`\s+[-\x{2013}\x{2014}\x{2022}\x{00b7}]\s+`)

// Identifies the institution segment in an education entry line (rule
// E-1).
//
//nolint:misspell // "instituto" is Spanish, not a misspelling of "institution".
var reSchool = regexp.MustCompile(`(?i)universi|instituto|institut|college|school|tecnol[oó]gico|academy|faculty`)

// Markdown link and bare URL, for FR-45 story hrefs. Group 1 of reMDLink
// is the target. reBareURL deliberately excludes `)`, `]`, `<` and `>`
// from the match, and the result is then right-trimmed of `.,;:!?)]` per
// P-6, so `see https://x.com/y.` yields `https://x.com/y`.
var reMDLink = regexp.MustCompile(`\[[^\]]*\]\((https?://[^)\s]+)\)`)
var reBareURL = regexp.MustCompile(`https?://[^\s<>()\[\]]+`)

// Matches an Nx multiplier for the `performance` theme (§5.2). Must match
// `40x`, `10x–40x`, `2.5x`. Must NOT match `x` alone or a hex literal.
var reNx = regexp.MustCompile(`(?i)\b\d+(?:\.\d+)?x\b`)

// Matches inline bold inside SUMMARY prose, converted to `<strong>` for
// `hero.bio_html` (FR-44). Everything outside these spans is
// HTML-escaped.
var reBoldInline = regexp.MustCompile(`\*\*([^*]+)\*\*`)

// ---------------------------------------------------------------------
// Helper regexes not in the §4.7 inventory. Needed for prose behaviour
// the inventory documents but doesn't provide a pattern for; see the
// final report for why each exists.
// ---------------------------------------------------------------------

// reItalicSpan matches an inline italic span (single asterisks). Used
// together with reBoldSpan to fully strip emphasis from certification
// text (H-5: "verbatim, emphasis stripped"), which needs interior
// stripping, not just the whole-line anchoring reTrimEmph provides.
var reItalicSpan = regexp.MustCompile(`\*([^*]+)\*`)

// reOpenEnded recognizes an endpoint token that must normalize to exactly
// datePresent (H-3 / D-31), composed from the shared openPat fragment.
var reOpenEnded = regexp.MustCompile(`(?i)^(?:` + openPat + `)$`)

// ---------------------------------------------------------------------
// §4.2 The Doc model
// ---------------------------------------------------------------------

// SectionKind is a closed enum. Prefixed Section* to avoid colliding with
// model.go's Experience and Education resume types.
type SectionKind int

const (
	SectionSummary SectionKind = iota
	SectionExperience
	SectionEducation
	SectionCertifications
	SectionSkills
	SectionLanguages
	SectionSpeaking
	SectionOpenSource
	SectionHomelab
	SectionUnknown
)

// Doc is the shape-only parse result: markdown structure, no resume
// semantics.
type Doc struct {
	Name     string
	RoleLine string
	Sections []Section
	Warnings []Warning
}

// Section groups entries, story bullets and prose under one SectionKind.
type Section struct {
	Kind    SectionKind
	Slug    string
	Entries []Entry
	Bullets []Bullet
	Prose   []string
}

// Entry is a work-history or education entry. Segments holds the
// AD-11/H-4 (or E-1, for education) normalized fields: for non-education
// sections it is always [company, title, location]; for education it is
// always [institution, detail].
type Entry struct {
	Raw      string
	Segments []string
	Start    string
	End      string
	Bullets  []Bullet
	Techs    []string
}

// Bullet is one list item's text, attached to whichever entry or section
// was open when it was read.
type Bullet struct {
	Line int
	Text string
}

// Warning records a non-fatal parse degradation (H-8).
type Warning struct {
	Line int
	Msg  string
}

// ---------------------------------------------------------------------
// §4.3 Line classifier
// ---------------------------------------------------------------------

type lineKind int

const (
	kindBlank lineKind = iota
	kindHeading
	kindBullet
	kindItalic
	kindBoldOnly
	kindBoldish
	kindPlain
)

// classify assigns a lineKind from the line alone, in the normative order
// 1-7. Earlier rules win: rule 3 (Bullet) before 5/6 is load-bearing, see
// ARCHITECTURE.md §4.3.
func classify(line string) lineKind {
	switch {
	case strings.TrimSpace(line) == "":
		return kindBlank
	case reHeading.MatchString(line):
		return kindHeading
	case reBullet.MatchString(line):
		return kindBullet
	case reItalic.MatchString(line):
		return kindItalic
	case reBoldOnly.MatchString(line):
		return kindBoldOnly
	case reBoldSpan.MatchString(line):
		return kindBoldish
	default:
		return kindPlain
	}
}

// ---------------------------------------------------------------------
// §4.4 Section recognition
// ---------------------------------------------------------------------

type sectionKeyword struct {
	Prefix string
	Kind   SectionKind
	Slug   string
}

// Ordered slice, not a map (FR-3): deterministic iteration.
//
//nolint:gochecknoglobals // AD-07: prefix-matched in declaration order, so the table is ordered and package-level.
var sectionKeywords = []sectionKeyword{
	{"summary", SectionSummary, ""},
	{"profile", SectionSummary, ""},
	{"about", SectionSummary, ""},
	{"experience", SectionExperience, ""},
	{"employment", SectionExperience, ""},
	{"work history", SectionExperience, ""},
	{"education", SectionEducation, ""},
	{"academic", SectionEducation, ""},
	{"certification", SectionCertifications, ""},
	{"certificate", SectionCertifications, ""},
	{"course", SectionCertifications, ""},
	{"skills", SectionSkills, ""},
	{"languages", SectionLanguages, ""},
	{"speaking", SectionSpeaking, tagCommunity},
	{"talks", SectionSpeaking, tagCommunity},
	{"teaching", SectionSpeaking, tagCommunity},
	{tagOpenSource, SectionOpenSource, tagOpenSource},
	{"open-source", SectionOpenSource, tagOpenSource},
	{tagHomelab, SectionHomelab, ""},
	{"home lab", SectionHomelab, ""},
	{"projects", SectionHomelab, ""},
	{"side projects", SectionHomelab, ""},
}

// matchSectionKeyword prefix-matches text (case-insensitively) against
// the keyword table.
func matchSectionKeyword(text string) (SectionKind, string, bool) {
	lower := strings.ToLower(strings.TrimSpace(text))
	for _, kw := range sectionKeywords {
		if strings.HasPrefix(lower, kw.Prefix) {
			return kw.Kind, kw.Slug, true
		}
	}

	return SectionUnknown, "", false
}

// isAllUpper reports whether every letter in s is upper case. A string
// with no letters at all is not "all upper".
func isAllUpper(s string) bool {
	hasLetter := false

	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true

			if !unicode.IsUpper(r) {
				return false
			}
		}
	}

	return hasLetter
}

// sectionBoundary reports whether line opens a new section, per the two
// acceptance paths of §4.4:
//
//   - Heading is always a boundary (known=false for an unrecognized
//     keyword, but isBoundary is still true).
//   - BoldOnly is a boundary only if its stripped text prefix-matches a
//     known keyword and its letters are upper case; otherwise
//     isBoundary is false and the line falls through to entry
//     recognition.
func sectionBoundary(line string, kind lineKind) (SectionKind, string, bool, bool) {
	switch kind {
	case kindHeading:
		m := reHeading.FindStringSubmatch(line)

		text := reTrimEmph.ReplaceAllString(m[2], "")
		if sk, sl, ok := matchSectionKeyword(text); ok {
			return sk, sl, true, true
		}

		return SectionUnknown, "", true, false
	case kindBoldOnly:
		m := reBoldOnly.FindStringSubmatch(line)

		text := strings.TrimSpace(m[1])
		if sk, sl, ok := matchSectionKeyword(text); ok && isAllUpper(text) {
			return sk, sl, true, true
		}

		return SectionUnknown, "", false, false
	case kindBlank, kindBullet, kindItalic, kindBoldish, kindPlain:
		return SectionUnknown, "", false, false
	}

	return SectionUnknown, "", false, false
}

// ---------------------------------------------------------------------
// §4.5 Identity block
// ---------------------------------------------------------------------

// extractIdentity walks lines from the start of the document up to (but
// not including) the first section boundary, applying the H-2 rules in
// order: contact-line discard first, then name, then the D-25 one-line
// shape or the verbatim role-line fallback. next is the index of the
// first line not consumed by identity parsing.
//
// A recognized section heading always ends identity parsing immediately.
// An unrecognized Heading is let through as a name candidate exactly
// once, before name is captured: H-2 rule 2 explicitly allows the name
// line to be an H1 heading (the D-25 shape "# Name — Role Line" matches
// reHeading but prefix-matches no known keyword), so it must not be
// preempted by the general "Heading is always a boundary" rule of §4.4.
// headingText returns a heading line's text with its `#` markers removed,
// or the line unchanged when it is not a heading.
func headingText(line string, kind lineKind) string {
	if kind != kindHeading {
		return line
	}

	if m := reHeading.FindStringSubmatch(line); m != nil {
		return m[2]
	}

	return line
}

// endsIdentity reports whether the line closes the identity block: a
// section boundary, once a name has been read or the heading keyword is
// a known section.
func endsIdentity(line string, kind lineKind, name string) bool {
	_, _, isBoundary, known := sectionBoundary(line, kind)

	return isBoundary && (name != "" || known)
}

// skipInIdentity reports whether the line carries no identity content:
// blank, or contact-bearing and discarded per H-2.
func skipInIdentity(line string, kind lineKind) bool {
	return kind == kindBlank || reContactish.MatchString(line)
}

// identityFromLine reads a name-bearing line. It returns the emphasis- and
// heading-stripped name, plus a role when the line carries the D-25
// one-line `Name — Role` shape (empty role otherwise). An empty name means
// the line held nothing usable.
func identityFromLine(line string, kind lineKind) (string, string) {
	stripped := reTrimEmph.ReplaceAllString(headingText(line, kind), "")
	if stripped == "" {
		return "", ""
	}

	if parts := reNameSplit.Split(stripped, 2); len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	return stripped, ""
}

func extractIdentity(lines []string, kinds []lineKind) (string, string, int) {
	var name, role string

	idx := 0
	for idx < len(lines) {
		if endsIdentity(lines[idx], kinds[idx], name) {
			break
		}

		line := lines[idx]
		kind := kinds[idx]
		idx++

		if skipInIdentity(line, kind) {
			continue
		}

		if name == "" {
			candidate, splitRole := identityFromLine(line, kind)
			if candidate == "" {
				continue
			}

			if splitRole != "" {
				// D-25: "# Name — Role Line", terminates identity parsing.
				return candidate, splitRole, idx
			}

			name = candidate

			continue
		}

		if role == "" && reSepLine.MatchString(line) {
			// Verbatim: no splitting, no re-ordering, no whitespace
			// collapsing.
			role = line
			return name, role, idx
		}
	}

	return name, role, idx
}

// ---------------------------------------------------------------------
// §4.6 Entry and bullet handling
// ---------------------------------------------------------------------

// stripEmphasis removes both bold and italic markers from s, keeping the
// inner text (used for Certifications bullets, which must be verbatim
// with emphasis stripped).
func stripEmphasis(s string) string {
	s = reBoldSpan.ReplaceAllString(s, "$1")
	s = reItalicSpan.ReplaceAllString(s, "$1")

	return strings.TrimSpace(s)
}

// normalizeDateToken normalizes an open-ended endpoint
// (present|current|now|to date|ongoing, any case) to exactly datePresent;
// any other token passes through unchanged.
func normalizeDateToken(tok string) string {
	tok = strings.TrimSpace(tok)
	if reOpenEnded.MatchString(tok) {
		return datePresent
	}

	return tok
}

// dateLineText extracts the date-bearing text from a lookahead candidate
// line: unwraps a whole-line italic span, otherwise the line as-is.
func dateLineText(line string, kind lineKind) string {
	if kind == kindItalic {
		if m := reItalic.FindStringSubmatch(line); m != nil {
			return strings.TrimSpace(m[1])
		}
	}

	return strings.TrimSpace(line)
}

// splitSegments applies H-4's reDashSplit to text, trimming each segment.
func splitSegments(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	parts := reDashSplit.Split(text, -1)

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}

	return out
}

// applyEducationSplit implements E-1: if exactly one segment matches
// reSchool, it is the institution and the rest join with " · " as the
// detail; otherwise segment 0 is the institution.
func applyEducationSplit(segments []string) []string {
	if len(segments) == 0 {
		return []string{"", ""}
	}

	matchIdx, matches := -1, 0

	for idx, seg := range segments {
		if reSchool.MatchString(seg) {
			matchIdx = idx
			matches++
		}
	}

	if matches == 1 {
		rest := make([]string, 0, len(segments)-1)
		for idx, seg := range segments {
			if idx != matchIdx {
				rest = append(rest, seg)
			}
		}

		return []string{segments[matchIdx], strings.Join(rest, " · ")}
	}

	return []string{segments[0], strings.Join(segments[1:], " · ")}
}

// applyExperienceSplit implements the H-4 arity table, always producing a
// [company, title, location] triple. Arity outside 1..3 warns (FR-36) but
// still produces a best-effort triple: first segment is company, last is
// location, the rest rejoin with " – " as the title.
func applyExperienceSplit(segments []string, rawText string) ([]string, string) {
	switch count := len(segments); count {
	case 0:
		return []string{"", "", ""}, fmt.Sprintf("entry %q has no parsable segments", rawText)
	case 1:
		return []string{segments[0], "", ""}, ""
	case 2:
		return []string{segments[0], segments[1], ""}, ""
	case 3:
		return []string{segments[0], segments[1], segments[2]}, ""
	default:
		company := segments[0]
		location := segments[count-1]
		title := strings.Join(segments[1:count-1], " – ")

		return []string{company, title, location},
			fmt.Sprintf("entry %q has %d dash-separated segments (expected 1-3); merged middle segments into title", rawText, count)
	}
}

// entryResult is the outcome of attempting to open an entry at a given
// line.
type entryResult struct {
	entry    Entry
	consumed int // index of a lookahead date line consumed, or -1
	isEntry  bool
	warnMsg  string
}

// tryEntry implements H-3 entry recognition plus the AD-11 ordering
// (date-range extraction, then trailing-parenthetical discard, then
// H-4/E-1 splitting) for the Boldish/BoldOnly line at lines[i].
// datesFrom reads a reDateRange submatch into start/end. A range with no
// second token ("Aug 2025 (1-week engagement)") collapses to start == end.
func datesFrom(match []string) (string, string) {
	start := normalizeDateToken(match[1])
	if match[2] == "" {
		return start, start
	}

	return start, normalizeDateToken(match[2])
}

// lookaheadDate scans up to the next 2 non-blank lines after i for a date
// range, per the italic-date-line shape (§4.6). It stops at the first line
// that is neither italic nor plain. consumed is the index of the line the
// date came from, or -1 when no date was found.
func lookaheadDate(lines []string, kinds []lineKind, idx int) (string, string, int, bool) {
	checked := 0

	for next := idx + 1; next < len(lines) && checked < 2; next++ {
		if kinds[next] == kindBlank {
			continue
		}

		checked++

		if kinds[next] != kindItalic && kinds[next] != kindPlain {
			break
		}

		if sm := reDateRange.FindStringSubmatch(dateLineText(lines[next], kinds[next])); sm != nil {
			start, end := datesFrom(sm)

			return start, end, next, true
		}
	}

	return "", "", -1, false
}

// tryBoldEntry runs tryEntry only for a bold line inside an open section;
// anything else is not an entry candidate.
func tryBoldEntry(lines []string, kinds []lineKind, i int, kind lineKind, cur *Section) entryResult {
	if cur == nil || (kind != kindBoldOnly && kind != kindBoldish) {
		return entryResult{}
	}

	return tryEntry(lines, kinds, i, cur.Kind)
}

func tryEntry(lines []string, kinds []lineKind, idx int, section SectionKind) entryResult {
	line := lines[idx]
	text := strings.TrimSpace(reBoldSpan.ReplaceAllString(line, "$1"))

	var (
		start, end string
		dateFound  bool
	)

	consumed := -1

	if m := reDateRange.FindStringSubmatchIndex(text); m != nil {
		start, end = datesFrom(reDateRange.FindStringSubmatch(text))
		text = text[:m[0]] + text[m[1]:]
		dateFound = true
	} else {
		start, end, consumed, dateFound = lookaheadDate(lines, kinds, idx)
	}

	if !dateFound && section != SectionEducation {
		return entryResult{}
	}

	text = strings.TrimSpace(reTrailingParen.ReplaceAllString(text, ""))
	segments := splitSegments(text)

	var (
		fields  []string
		warnMsg string
	)

	if section == SectionEducation {
		fields = applyEducationSplit(segments)
	} else {
		fields, warnMsg = applyExperienceSplit(segments, text)
	}

	return entryResult{
		entry: Entry{
			Raw:      line,
			Segments: fields,
			Start:    start,
			End:      end,
		},
		consumed: consumed,
		isEntry:  true,
		warnMsg:  warnMsg,
	}
}

// attachBullet implements H-5: any bullet attaches to the most recent
// open entry; failing that, to the section, for the bullets-only story
// sections. Certifications bullets are verbatim with emphasis stripped.
// Skills and Languages bullets are discarded (D-26).
func attachBullet(cur *Section, openEntry *Entry, lineNo int, text string) {
	if cur == nil {
		return
	}

	switch cur.Kind {
	case SectionCertifications:
		cur.Bullets = append(cur.Bullets, Bullet{Line: lineNo, Text: stripEmphasis(text)})
	case SectionSkills, SectionLanguages:
		// consumed and discarded, D-26
	case SectionSummary, SectionExperience, SectionEducation,
		SectionSpeaking, SectionOpenSource, SectionHomelab, SectionUnknown:
		if openEntry != nil {
			openEntry.Bullets = append(openEntry.Bullets, Bullet{Line: lineNo, Text: text})
			return
		}
		// No open entry: attaches to the section for Speaking/OpenSource/
		// Homelab per H-5. Any other kind falls back here too, rather
		// than silently dropping the line (H-8).
		cur.Bullets = append(cur.Bullets, Bullet{Line: lineNo, Text: text})
	}
}

// splitTechs implements H-7's comma split, order and casing preserved
// verbatim (D-21).
func splitTechs(value string) []string {
	parts := strings.Split(value, ",")

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}

	return out
}

// ---------------------------------------------------------------------
// Segmentation pass and entry point
// ---------------------------------------------------------------------

// splitLines splits src into right-trimmed lines (§4.1).
func splitLines(src string) []string {
	raw := strings.Split(src, "\n")

	lines := make([]string, len(raw))
	for i, l := range raw {
		lines[i] = strings.TrimRight(l, " \t\r")
	}

	return lines
}

// docParser carries the segment-pass state (§4.1 pass 2) that the line
// handlers below mutate as they walk the document: the open section, the
// open entry, and the lines a lookahead has already claimed.
type docParser struct {
	lines     []string
	kinds     []lineKind
	consumed  []bool
	sections  []Section
	warnings  []Warning
	cur       *Section
	openEntry *Entry
}

// closeEntry attaches the open entry to the open section, if both exist.
func (p *docParser) closeEntry() {
	if p.cur != nil && p.openEntry != nil {
		p.cur.Entries = append(p.cur.Entries, *p.openEntry)
	}

	p.openEntry = nil
}

// closeSection closes any open entry, then appends the open section.
func (p *docParser) closeSection() {
	p.closeEntry()

	if p.cur != nil {
		p.sections = append(p.sections, *p.cur)
	}

	p.cur = nil
}

// warnf records a non-fatal degradation (H-8).
func (p *docParser) warnf(line int, format string, args ...any) {
	p.warnings = append(p.warnings, Warning{Line: line, Msg: fmt.Sprintf(format, args...)})
}

// openSection closes the current section and starts a new one.
func (p *docParser) openSection(kind SectionKind, slug string) {
	p.closeSection()
	p.cur = &Section{Kind: kind, Slug: slug}
}

// handleLine routes one line by its classified kind. Blank lines and the
// kinds fully handled by their own case return here; the rest fall through
// to handleContentLine.
func (p *docParser) handleLine(idx int) {
	if p.consumed[idx] {
		return
	}

	line, kind, lineNo := p.lines[idx], p.kinds[idx], idx+1

	switch kind {
	case kindBlank:
		return
	case kindHeading:
		p.openHeading(line, kind, lineNo)

		return
	case kindBullet:
		m := reBullet.FindStringSubmatch(line)
		attachBullet(p.cur, p.openEntry, lineNo, m[2])

		return
	case kindItalic, kindBoldOnly, kindBoldish, kindPlain:
		// Handled below: these open a section, an entry, or a tech line.
	}

	p.handleContentLine(idx, line, kind, lineNo)
}

// openHeading starts the section a `#` heading names, warning when the
// keyword is not in the table.
func (p *docParser) openHeading(line string, kind lineKind, lineNo int) {
	skind, slug, _, known := sectionBoundary(line, kind)
	p.openSection(skind, slug)

	if !known {
		m := reHeading.FindStringSubmatch(line)
		text := reTrimEmph.ReplaceAllString(m[2], "")
		p.warnf(lineNo, "unknown section heading %q", text)
	}
}

// handleContentLine tries, in order: an all-caps bold section boundary, a
// `Technologies:` line for the open entry, a new entry, and finally
// section prose.
func (p *docParser) handleContentLine(idx int, line string, kind lineKind, lineNo int) {
	skind, slug, isBoundary, _ := sectionBoundary(line, kind)
	if kind == kindBoldOnly && isBoundary {
		p.openSection(skind, slug)

		return
	}

	if m := reTechLine.FindStringSubmatch(line); m != nil && p.openEntry != nil {
		p.openEntry.Techs = splitTechs(m[1])
		p.closeEntry()

		return
	}

	if res := tryBoldEntry(p.lines, p.kinds, idx, kind, p.cur); res.isEntry {
		p.startEntry(res, lineNo)

		return
	}

	if p.cur != nil {
		p.cur.Prose = append(p.cur.Prose, line)
	}
}

// startEntry closes the previous entry and opens the one tryEntry built,
// marking any date line its lookahead claimed.
func (p *docParser) startEntry(res entryResult, lineNo int) {
	p.closeEntry()

	entry := res.entry
	p.openEntry = &entry

	if res.consumed >= 0 {
		p.consumed[res.consumed] = true
	}

	if res.warnMsg != "" {
		p.warnf(lineNo, "%s", res.warnMsg)
	}
}

// parseDoc runs the classify and segment passes over src (§4.1 passes 1
// and 2) and returns the resulting Doc. It never fails for content: every
// degradation is recorded in Doc.Warnings (H-8).
func parseDoc(src string) Doc {
	lines := splitLines(src)

	kinds := make([]lineKind, len(lines))
	for i, l := range lines {
		kinds[i] = classify(l)
	}

	name, role, start := extractIdentity(lines, kinds)

	parser := &docParser{lines: lines, kinds: kinds, consumed: make([]bool, len(lines))}
	for idx := start; idx < len(lines); idx++ {
		parser.handleLine(idx)
	}

	parser.closeSection()

	return Doc{Name: name, RoleLine: role, Sections: parser.sections, Warnings: parser.warnings}
}
