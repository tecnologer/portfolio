package portfolio

// Tag inference, per ARCHITECTURE.md §5. All of it runs at extract time
// (AD-06) against the shape-only Doc/Section/Entry/Bullet types parse.go
// produces. Dictionaries and mappings are ordered slices, never maps, on
// any output path (AD-07, FR-3): maps here are for counting/lookup only.
//
// Not wired to extract.go yet (that mapping pass, and §5.5's title dedup,
// are a later step); this file only exposes the building blocks it will
// call.

import (
	"regexp"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------
// §5.1 Technology dictionary
// ---------------------------------------------------------------------

// matchMode selects how a techRule's aliases are turned into a pattern.
type matchMode int

const (
	// modeWord is T-1, the default: case-insensitive, word boundaries on
	// both sides.
	modeWord matchMode = iota
	// modeExact is T-2, for ambiguous short aliases (go, ai, c, r): case
	// sensitive, matched against the original-case text.
	modeExact
	// modeSymbol is T-3, for c#, c++, .net: custom boundaries on both
	// sides, because \b fails on the right of #/+ and on the left of a
	// leading dot (see ARCHITECTURE.md §5.1).
	modeSymbol
)

// techRule is one entry in the technology dictionary. Tag is the lowercase
// slug used for story/theme-style tags; Display is the human casing used
// verbatim in Experience.Technologies (PostgreSQL, not postgresql).
type techRule struct {
	Tag     string
	Display string
	Aliases []string
	Mode    matchMode
}

// techDict is the technology dictionary: the §7 seed set (in seed order),
// extended with the census of the `technologies:` lists in
// data/resume.yaml (experience entries for Noodle, DefectDojo, Puller
// Tech, Pentalog, Ubilogix, Alesayi Development Co., RedRabbit MX),
// appended in the order those entries are first seen. C++ is included
// alongside C# and .NET as the third modeSymbol example (ARCHITECTURE.md
// §5.1), even though it does not appear in the census.
//
//nolint:gochecknoglobals // AD-07: the tag dictionaries are ordered package-level tables; output order depends on their declaration order.
var techDict = []techRule{
	// --- §7 seed set ---
	{Tag: "go", Display: "Go", Aliases: []string{"Go"}, Mode: modeExact},
	{Tag: "grpc", Display: "gRPC", Aliases: []string{"gRPC"}},
	{Tag: "protobuf", Display: "Protobuf", Aliases: []string{"Protobuf"}},
	{Tag: "graphql", Display: "GraphQL", Aliases: []string{"GraphQL"}},
	{Tag: tagPostgreSQL, Display: displayPostgreSQL, Aliases: []string{displayPostgreSQL, "Postgres"}},
	{Tag: "sqlite", Display: "SQLite", Aliases: []string{"SQLite"}},
	{Tag: "mssql", Display: "MSSQL", Aliases: []string{"MSSQL"}},
	{Tag: "redis", Display: "Redis", Aliases: []string{"Redis"}},
	{Tag: "aws", Display: "AWS", Aliases: []string{"AWS"}},
	{Tag: "lambda", Display: "Lambda", Aliases: []string{"Lambda"}},
	{Tag: "appsync", Display: "AppSync", Aliases: []string{"AppSync"}},
	{Tag: "docker", Display: displayDocker, Aliases: []string{displayDocker}},
	{Tag: "kubernetes", Display: "Kubernetes", Aliases: []string{"Kubernetes"}},
	{Tag: "langchain", Display: "LangChain", Aliases: []string{"LangChain"}},
	{Tag: "openai", Display: "OpenAI", Aliases: []string{"OpenAI"}},
	{Tag: "ollama", Display: "Ollama", Aliases: []string{"Ollama"}},
	{Tag: "mcp", Display: "MCP", Aliases: []string{"MCP"}},
	{Tag: "zigbee", Display: "Zigbee", Aliases: []string{"Zigbee"}},
	{Tag: "thread", Display: "Thread", Aliases: []string{"Thread"}},
	{Tag: "csharp", Display: "C#", Aliases: []string{"C#"}, Mode: modeSymbol},
	{Tag: tagDotnet, Display: ".NET", Aliases: []string{".NET"}, Mode: modeSymbol},
	{Tag: "python", Display: "Python", Aliases: []string{"Python"}},
	{Tag: "typescript", Display: "TypeScript", Aliases: []string{"TypeScript"}},
	{Tag: "wpf", Display: "WPF", Aliases: []string{"WPF"}},
	{Tag: "vue", Display: "Vue", Aliases: []string{"Vue"}},
	{Tag: "bubbletea", Display: "Bubble Tea", Aliases: []string{"Bubble Tea"}},
	{Tag: "gorm", Display: "GORM", Aliases: []string{"GORM"}},
	{Tag: "entity-framework", Display: "Entity Framework", Aliases: []string{"Entity Framework"}},
	{Tag: "esp8266", Display: "ESP8266", Aliases: []string{"ESP8266"}},
	{Tag: "raspberry-pi", Display: "Raspberry Pi", Aliases: []string{"Raspberry Pi"}},
	// cpp is the third T-3 example (§5.1), not in the seed/census.
	{Tag: "cpp", Display: "C++", Aliases: []string{"C++"}, Mode: modeSymbol},

	// --- census additions: DefectDojo (data/resume.yaml:55-80) ---
	{Tag: "rest-api", Display: "REST API", Aliases: []string{"REST API", "REST"}},
	{Tag: "urfave-cli", Display: "urfave/cli", Aliases: []string{"urfave/cli"}},
	{Tag: "github-actions", Display: "GitHub Actions", Aliases: []string{"GitHub Actions"}},
	{Tag: "linux", Display: "Linux", Aliases: []string{"Linux"}},
	{Tag: "bash", Display: "Bash", Aliases: []string{"Bash"}},
	{Tag: "checkmarx-one", Display: "Checkmarx One", Aliases: []string{"Checkmarx One"}},
	{Tag: "wiz", Display: "Wiz", Aliases: []string{"Wiz"}},
	{Tag: "snyk", Display: "Snyk", Aliases: []string{"Snyk"}},
	{Tag: "sonarcloud", Display: "SonarCloud", Aliases: []string{"SonarCloud"}},
	{Tag: "sonarqube", Display: "SonarQube", Aliases: []string{"SonarQube"}},
	{Tag: "dependency-track", Display: "Dependency-Track", Aliases: []string{"Dependency-Track"}},
	{Tag: "iriusrisk", Display: "IriusRisk", Aliases: []string{"IriusRisk"}},
	{Tag: "akamai", Display: "Akamai", Aliases: []string{"Akamai"}},
	{Tag: "azure-devops", Display: "Azure DevOps", Aliases: []string{"Azure DevOps"}},
	{Tag: "servicenow", Display: "ServiceNow", Aliases: []string{"ServiceNow"}},
	{Tag: "slack", Display: "Slack", Aliases: []string{"Slack"}},
	{Tag: "telegram", Display: "Telegram", Aliases: []string{"Telegram"}},
	{Tag: "pagerduty", Display: "PagerDuty", Aliases: []string{"PagerDuty"}},
	{Tag: "git", Display: "Git", Aliases: []string{"Git"}},

	// --- census additions: Pentalog (data/resume.yaml:96-120) ---
	{Tag: "aws-cognito", Display: "AWS Cognito", Aliases: []string{"AWS Cognito", "Cognito"}},
	{Tag: "aws-s3", Display: "AWS S3", Aliases: []string{"AWS S3"}},
	{Tag: "aws-cloudformation", Display: "AWS CloudFormation", Aliases: []string{"AWS CloudFormation", "CloudFormation"}},
	{Tag: "aws-sam", Display: "AWS SAM", Aliases: []string{"AWS SAM"}},
	{Tag: "aws-secrets-manager", Display: "AWS Secrets Manager", Aliases: []string{"AWS Secrets Manager", "Secrets Manager"}},
	{Tag: "saml-sso", Display: "SAML SSO", Aliases: []string{"SAML SSO", "SAML"}},
	{Tag: "casbin", Display: "Casbin", Aliases: []string{"Casbin"}},

	// --- census additions: Ubilogix (data/resume.yaml:88) ---
	{Tag: "gitlab", Display: "GitLab", Aliases: []string{"GitLab"}},

	// --- census additions: Alesayi Development Co. (data/resume.yaml:128) ---
	{Tag: "angularjs", Display: "AngularJS", Aliases: []string{"AngularJS"}},
	{Tag: "javascript", Display: "JavaScript", Aliases: []string{"JavaScript"}},
	{Tag: "html5-css", Display: "HTML5/CSS", Aliases: []string{"HTML5/CSS"}},

	// --- census additions: RedRabbit MX (data/resume.yaml:136) ---
	{Tag: "coldfusion", Display: "ColdFusion", Aliases: []string{"ColdFusion"}},
	{Tag: "mongodb", Display: "MongoDB", Aliases: []string{"MongoDB"}},
}

// altPattern joins items into a QuoteMeta'd `a|b|c` alternation, shared by
// every rule kind below (tech, theme, role) so each is compiled once.
func altPattern(items []string) string {
	quoted := make([]string, len(items))
	for i, it := range items {
		quoted[i] = regexp.QuoteMeta(it)
	}

	return strings.Join(quoted, "|")
}

// compileTechRule builds the single pattern for one techRule per its Mode.
func compileTechRule(r techRule) *regexp.Regexp {
	alt := altPattern(r.Aliases)
	switch r.Mode {
	case modeExact:
		return regexp.MustCompile(`\b(?:` + alt + `)\b`)
	case modeSymbol:
		return regexp.MustCompile(`(?i)(^|[^\w#+.])(?:` + alt + `)($|[^\w#+.])`)
	case modeWord:
		return regexp.MustCompile(`(?i)\b(?:` + alt + `)\b`)
	default:
		return regexp.MustCompile(`(?i)\b(?:` + alt + `)\b`)
	}
}

// techPatterns is compiled once, in techDict declaration order, at package
// init.
//
//nolint:gochecknoglobals // Compiled once at init; recompiling per call would be a real cost.
var techPatterns = func() []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(techDict))
	for i, r := range techDict {
		out[i] = compileTechRule(r)
	}

	return out
}()

// matchTechRules returns every techDict rule matching text, in dictionary
// declaration order.
func matchTechRules(text string) []techRule {
	var out []techRule

	for i, rule := range techDict {
		if techPatterns[i].MatchString(text) {
			out = append(out, rule)
		}
	}

	return out
}

// matchTechTags returns the Tag (lowercase slug) of every matching rule,
// for story tag assembly (§5.4).
func matchTechTags(text string) []string {
	rules := matchTechRules(text)

	out := make([]string, len(rules))
	for i, r := range rules {
		out[i] = r.Tag
	}

	return out
}

// matchTechDisplays returns the Display (human casing) of every matching
// rule, for the per-entry Technologies field (§5.3).
func matchTechDisplays(text string) []string {
	rules := matchTechRules(text)

	out := make([]string, len(rules))
	for i, r := range rules {
		out[i] = r.Display
	}

	return out
}

// ---------------------------------------------------------------------
// §5.2 Theme keyword mapping
// ---------------------------------------------------------------------

// keywordRule is the shared shape of the theme and role mappings: a tag
// plus the modeWord keywords/phrases that trigger it.
type keywordRule struct {
	Tag      string
	Keywords []string
}

// themeMap is the nine theme rows of §5.2/§7, in declaration order (which
// is also emission order). The AI keyword is deliberately absent from the
// "ai" row's Keywords: it is modeExact (T-2) and handled as a special case
// in inferThemes, alongside reNx for tagPerformance — one pattern in one
// rule, not a general Extra field (§5.2).
//
//nolint:gochecknoglobals // AD-07: ordered theme table, see techDict.
var themeMap = []keywordRule{
	{Tag: tagPerformance, Keywords: []string{"throughput", "faster", "optimized"}},
	{Tag: tagMigration, Keywords: []string{"rewrote", "replaced", tagMigration}},
	{Tag: "cli", Keywords: []string{"CLI", "TUI"}},
	{Tag: "integrations", Keywords: []string{"connector", "integration"}},
	{Tag: "ai", Keywords: []string{"LLM", "agent"}},
	{Tag: "leadership", Keywords: []string{"manager", "mentored", "1:1s", "interview", tagSpeaker, "talks"}},
	{Tag: "open-source", Keywords: []string{tagOpenSource, "contribution", "merged"}},
	{Tag: "debugging", Keywords: []string{"debugged", "root-caused", "traced", "bug"}},
	{Tag: "concurrency", Keywords: []string{"goroutine", "parallel", "concurrency"}},
}

// aiExact is the T-2 modeExact pattern for the "ai" theme's AI keyword.
var aiExact = regexp.MustCompile(`\bAI\b`)

// compileKeywordRules builds one modeWord alternation pattern per rule.
func compileKeywordRules(rules []keywordRule) []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(rules))
	for i, r := range rules {
		out[i] = regexp.MustCompile(`(?i)\b(?:` + altPattern(r.Keywords) + `)\b`)
	}

	return out
}

//nolint:gochecknoglobals // Compiled once at init, see techPatterns.
var themePatterns = compileKeywordRules(themeMap)

// inferThemes returns every themeMap tag triggered anywhere in text, in
// themeMap declaration order. reNx (§4.7, parse.go) is the performance
// row's Nx special case; aiExact is the ai row's AI special case.
func inferThemes(text string) []string {
	var out []string

	for i, rule := range themeMap {
		matched := themePatterns[i].MatchString(text)

		switch rule.Tag {
		case tagPerformance:
			matched = matched || reNx.MatchString(text)
		case "ai":
			matched = matched || aiExact.MatchString(text)
		}

		if matched {
			out = append(out, rule.Tag)
		}
	}

	return out
}

// ---------------------------------------------------------------------
// §5.3 Roles, technologies, slugs
// ---------------------------------------------------------------------

// roleMap is the role-keyword mapping, in declaration order, applied to
// the title segment only.
//
//nolint:gochecknoglobals // AD-07: ordered role table, see techDict.
var roleMap = []keywordRule{
	{Tag: tagArchitect, Keywords: []string{tagArchitect}},
	{Tag: tagFullStack, Keywords: []string{"full stack", tagFullStack}},
	{Tag: tagBackEnd, Keywords: []string{"software engineer", "backend", tagBackEnd, "go engineer"}},
	{Tag: tagPeopleManager, Keywords: []string{tagPeopleManager, "manager"}},
	{Tag: "security tooling", Keywords: []string{"security"}},
	{Tag: "consultant", Keywords: []string{"consultant", "freelance"}},
	{Tag: tagSpeaker, Keywords: []string{tagSpeaker, "teaching"}},
}

//nolint:gochecknoglobals // Compiled once at init, see techPatterns.
var rolePatterns = compileKeywordRules(roleMap)

// inferRoles returns every roleMap tag triggered in titleSegment, in
// roleMap declaration order. "Software Architect / Full Stack Developer"
// yields [architect, full-stack].
func inferRoles(titleSegment string) []string {
	var out []string

	for i, rule := range roleMap {
		if rolePatterns[i].MatchString(titleSegment) {
			out = append(out, rule.Tag)
		}
	}

	return out
}

// entryTechnologies returns an entry's technology list: the verbatim H-7
// `Technologies:` line when the parser captured one (order and casing
// preserved), otherwise the union of techDict matches across the entry's
// bullets and its title segment (Segments[1]), in dictionary declaration
// order with each rule's Display casing.
func entryTechnologies(entry Entry) []string {
	if len(entry.Techs) > 0 {
		return entry.Techs
	}

	var builder strings.Builder
	if len(entry.Segments) > 1 {
		builder.WriteString(entry.Segments[1])
	}

	for _, b := range entry.Bullets {
		builder.WriteByte('\n')
		builder.WriteString(b.Text)
	}

	return matchTechDisplays(builder.String())
}

// slugStopWords are the corporate suffixes dropped by slug. Lookup only,
// never iterated.
//
//nolint:gochecknoglobals // Lookup-only set; never iterated on an output path (AD-07).
var slugStopWords = map[string]bool{
	"inc":         true,
	"co":          true,
	"llc":         true,
	"ltd":         true,
	"s.a.":        true,
	"development": true,
	"mx":          true,
}

// isSlugStopWord reports whether word (already lowercased) is a corporate
// stop-word, trying both the word as-is (for dotted abbreviations like
// "s.a.") and with one trailing "." stripped (for "co.", "inc.", "ltd.").
func isSlugStopWord(word string) bool {
	if slugStopWords[word] {
		return true
	}

	return slugStopWords[strings.TrimSuffix(word, ".")]
}

// reSlugSep turns parentheses into word separators, so a parenthesized
// abbreviation like "(ADCO)" becomes its own token instead of being
// discarded with the parens.
var reSlugSep = regexp.MustCompile(`[()]+`)

// slug lowercases name, drops corporate stop-words, and joins what's left
// with hyphens (§5.3). "DefectDojo Inc" -> "defectdojo", "Puller Tech" ->
// "puller-tech", "Alesayi Development Co. (ADCO)" -> "alesayi-adco".
func slug(name string) string {
	name = reSlugSep.ReplaceAllString(name, " ")
	fields := strings.Fields(name)

	words := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.Trim(field, ",;")

		key := strings.ToLower(field)
		if isSlugStopWord(key) {
			continue
		}

		words = append(words, strings.ToLower(strings.TrimSuffix(field, ".")))
	}

	return strings.Join(words, "-")
}

// reLinkedBold matches a bold span that is also a markdown link label,
// e.g. the "Tempura" in `[**Tempura**](https://…)`.
var reLinkedBold = regexp.MustCompile(`\[\*\*([^*]+)\*\*\]`)

// homelabSlug implements the Homelab per-story slug rule (§5.3, D-20,
// FR-7): the first bold or linked proper noun in the bullet, falling back
// to tagHomelab when the bullet names no project. reBoldSpan is parse.go's
// existing any-bold-span pattern (§4.7), reused here rather than
// redeclared.
func homelabSlug(bulletText string) string {
	linkedLoc := reLinkedBold.FindStringSubmatchIndex(bulletText)
	boldLoc := reBoldSpan.FindStringSubmatchIndex(bulletText)

	var name string

	switch {
	case linkedLoc == nil && boldLoc == nil:
		return tagHomelab
	case linkedLoc == nil:
		name = bulletText[boldLoc[2]:boldLoc[3]]
	case boldLoc == nil:
		name = bulletText[linkedLoc[2]:linkedLoc[3]]
	case linkedLoc[0] <= boldLoc[0]:
		name = bulletText[linkedLoc[2]:linkedLoc[3]]
	default:
		name = bulletText[boldLoc[2]:boldLoc[3]]
	}

	if s := slug(name); s != "" {
		return s
	}

	return tagHomelab
}

// ---------------------------------------------------------------------
// §5.4 Per-story tag assembly
// ---------------------------------------------------------------------

// dedupPreserveOrder drops repeats from items, keeping the first
// occurrence of each. The map is used only for membership testing.
func dedupPreserveOrder(items []string) []string {
	seen := make(map[string]bool, len(items))

	out := make([]string, 0, len(items))
	for _, item := range items {
		if seen[item] {
			continue
		}

		seen[item] = true
		out = append(out, item)
	}

	return out
}

// storyTags assembles one story's tags per §5.4: theme tags (themeMap
// order), then tech tags (techDict order), then sourceSlug — deduplicated
// keeping first occurrence. sourceSlug is always appended, which is what
// guarantees the FR-20 floor of >=1 tag.
func storyTags(text, sourceSlug string) []string {
	all := make([]string, 0, 8)
	all = append(all, inferThemes(text)...)
	all = append(all, matchTechTags(text)...)
	all = append(all, sourceSlug)

	return dedupPreserveOrder(all)
}

// ---------------------------------------------------------------------
// §5.6 Suggestions
// ---------------------------------------------------------------------

// suggestions computes search.suggestions from every story's tag list and
// source slug (§5.6): tagPerformance pinned at index 0 unconditionally,
// then the top 7 remaining tags by frequency descending (ties broken by
// tag ascending in byte order), excluding tagPerformance itself and any
// tag that is also some story's source slug. Result length is 1..8 by
// construction.
func suggestions(storyTags [][]string, sourceSlugs []string) []string {
	isSlug := make(map[string]bool, len(sourceSlugs))
	for _, s := range sourceSlugs {
		isSlug[s] = true
	}

	freq := make(map[string]int)

	var candidates []string

	for _, tags := range storyTags {
		for _, tag := range tags {
			if tag == tagPerformance || isSlug[tag] {
				continue
			}

			if _, ok := freq[tag]; !ok {
				candidates = append(candidates, tag)
			}

			freq[tag]++
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if freq[a] != freq[b] {
			return freq[a] > freq[b]
		}

		return a < b
	})

	if len(candidates) > 7 {
		candidates = candidates[:7]
	}

	out := make([]string, 0, 1+len(candidates))
	out = append(out, tagPerformance)

	return append(out, candidates...)
}
