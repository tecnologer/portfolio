# profilegen — Architecture: markdown-source-first pipeline

Status: revision 1 of the architecture under requirements revision 3. Written
2026-08-06. Supersedes the previous `.claude/ARCHITECTURE.md` in full.

Sole authoritative input: `.claude/REQUERIMENTS.md` (revision 3 final, D-01 to D-31,
zero open questions). Note the filename spelling: **REQUERIMENTS**, no "I". The
previous revision of this file cited `claude/REQUIREMENTS.md` at line 3, which was
wrong twice (missing dot, wrong spelling); corrected here per D-19.

This document specifies *what* to build and *where*. It contains no implementation
code, with one deliberate exception: the regex inventory of §4.7, which is normative
and must be copied verbatim.

Every repo-state claim below was read this session. Line citations are citations of
lines actually read. Where the requirements and the repo disagree, the conflict is
recorded in §12 and **not** silently resolved.

---

## 1. Decision table

Each row traces to the FR / D / H it implements. Decisions are numbered AD-nn.
References to the previous architecture's D1..D15 are marked **superseded**.

| # | Decision | Traces to | Rationale, and what was rejected |
|---|---|---|---|
| **AD-01a** | **`main.go` alone is `package main` at the repo root; `build.go`, `extract.go`, `parse.go`, `tags.go`, `model.go` and all tests move to `internal/portfolio`. One internal package, not several.** `Build` and `Extract` are the only identifiers crossing to `main`. | NFR-04, FR-32; **supersedes AD-01** | AD-01's premise (package boundaries buy nothing) still holds *between* the generator's files — `extract.go`/`tags.go` share seven of `parse.go`'s compiled regexes and `entryTechnologies` takes a `parse.Entry`, so splitting further would force exporting internals or duplicating tolerance rules. What it got wrong is the CLI boundary: nothing stopped `main.go` accreting parser logic, and 10 `.go` files sat interleaved with `data/`, `template/`, `page/` and repo metadata at the root. `internal/` (not `pkg/`) states the true constraint: no external consumers, ever. The Makefile, CI and `go.mod` are unchanged because `main.go` stays at the root. Rejected: `cmd/profilegen/` (one binary, breaks `go run .`); a 4-way split; `internal/resume` + `internal/extract` (closest call — revisit when a second output target appears, not on LOC growth); `internal/engine` (names no domain). |
| ~~**AD-01**~~ | ~~Single `package main` at the repo root. Files: `main.go`, `build.go`, `extract.go`, `parse.go`, `tags.go`, `model.go`. No sub-packages, no `internal/`.~~ **Superseded by AD-01a.** | NFR-04, FR-32 | The program is 576 LOC today (build.go 64, extract.go 200, main.go 74, mdout.go 98, model.go 140, verified) and lands near 1100 after the parser. One binary, one maintainer, no library consumers: package boundaries would buy import ceremony and nothing else. Rejected: `internal/parse` + `internal/model` (no second consumer to protect from). |
| **AD-02** | Exactly one third-party dependency: `github.com/goccy/go-yaml`, already at `go.mod:5`. `go mod tidy` to drop the wrong `// indirect` marker. `regexp`, `strings`, `bytes`, `html/template`, `unicode/utf8` cover everything else. | NFR-01, NFR-02 | Rejected: a markdown AST library. It gives inline emphasis and list normalization for free but cannot express "bold-ish line plus a nearby date line is an entry", which is the whole of H-3. It would cost a dependency and still leave the semantic layer hand-written. |
| **AD-03** | **`mdout.go` is deleted outright. The `mdFileName` constant moves into `build.go`.** | FR-40, D-08; **supersedes old D6** | FR-40 deletes `RenderMarkdown` (`mdout.go:16-98`), which would leave `mdout.go` an 11-line file holding one string constant (`mdout.go:8-11`). `build.go:50` is the constant's only supplier and `template/index.html.tmpl:443,551,560` its only consumers, all reached through build. The old D6 (`static.go` as the home for values rendered by both HTML and markdown) has no premise left: after FR-40 nothing is dual-rendered, so there is no shared-constant pressure and no `static.go`. Rejected: keeping `mdout.go` as a one-const file (file-count inflation); a new `static.go` (an abstraction with zero second consumer). `grep -r RenderMarkdown` returning nothing (FR-40 acceptance) holds either way. |
| **AD-04** | **The template data interface is a named struct in `build.go` with flat fields: `PageData{*Resume; Version string; MDName string}`.** No `.Static` sub-struct. | FR-12, FR-13, FR-34 | After the FR-13 / D-23 / D-24 reduction, build supplies exactly two values beyond the resume: `Version` and `MDName`. A `.Static` sub-struct would force edits to three working template lines (`:443`, `:551`, `:560`) to buy grouping for two scalars. Named rather than anonymous (today it is anonymous, `build.go:46-50`) so `build_test.go` and the FR-35 contract test can construct it and render the template against a fixture. `{{.MDName}}` and `{{.Version}}` keep resolving unchanged. Rejected: `.Static` (three edits, no benefit); keeping the anonymous struct (untestable from another function). |
| **AD-05** | Parser splits into two layers with one seam: `parse.go` (line classifier + section state machine → `Doc`, knows markdown shape only) and `extract.go` (`Doc` → `Resume`, knows resume semantics only). | FR-6, H-1..H-8, FR-37 | The seam is what makes the FR-37 fixture suite testable at two granularities: shape drift against `Doc`, semantics against the YAML. `Doc` carries only the fields the five section modes consume (§4.2); the previous revision's four overlapping accumulators and `KV map[string]string` are cut, the map also because map iteration contradicts determinism (FR-3). |
| **AD-06** | All inference (tags, per-entry technologies, roles, suggestions) runs at **extract** time and is stored in `data/resume.yaml`. Build performs zero inference. | FR-10, FR-11, FR-24, D-04 | The YAML is the committed contract CI builds from; storing inference results there keeps `build` a pure renderer, makes chip churn visible in `git diff data/resume.yaml`, and needs no template change (`template/index.html.tmpl:443,598` already consume `.Search.Suggestions`). |
| **AD-07** | Technology dictionary, theme keyword mapping and role keyword mapping live in `tags.go` as **ordered slices**, never maps. | FR-21, FR-3, N6 | Ordered slices give deterministic iteration by construction; Go map order is randomized. Rejected: a repo data file (FR-21 permits it, but it adds a loader, a schema and a parse-error path for content only one person edits, in the same commit either way). Rejected outright per D-14 / N6: extending the dictionary from the SKILLS section. |
| **AD-08** | Suggestion index 0 is the literal string `performance`, injected before frequency ranking, never subject to it. | FR-11, FR-22, FR-28, FR-30, D-04 | `template/index.html.tmpl:443` renders it as the visible banner keyword inside `grep -i "…"` and `:683` indexes it again in browser JS. Pinning it in the extract computation, rather than in the template, keeps `.Search.Suggestions` a single ordered list that both the banner and the palette chips read (`:598`, `:633`). |
| **AD-09** | Build renders into a `bytes.Buffer` and writes `page/index.html` with a single `os.WriteFile` only after `Execute` returns nil. Build writes **exactly one file** and reads exactly two. | FR-34, FR-12, FR-39, D-07, AC-3, AC-8 | Today `build.go:40-44` does `os.Create` **before** `tmpl.Execute` at `:52`, so a template error truncates the previous good output to zero bytes. Buffering is three lines and removes the failure mode entirely. |
| **AD-10** | Extract validates the in-memory `Resume` before marshalling. On failure: print the failed check to stderr, exit non-zero, **write nothing**. | FR-27, FR-28, D-09, AC-9 | H-8's per-line degradation is right at line granularity and dangerous in aggregate: a source whose entry shape changed yields zero entries and a written YAML that then fails at `template/index.html.tmpl:443` in CI. The gate is the aggregate backstop H-8 names. |
| **AD-11** | Date-range extraction runs on the **raw** entry text before any dash splitting; the matched span is stripped, then the remainder is split. **This ordering is normative.** | H-3, D-15 (P-1), D-31 | The previous architecture split first (its §3.2) and the review caught it: the split pattern `\s+[-–—]\s+` matches the dash *inside* `Apr 2024 – Present`, shattering the range into two segments and pushing `Apr 2024` into `location`. The same bug hits education spans (`2009 – 2014`). Ordering, not pattern, is the fix. |
| **AD-12** | Golden fixtures are the executable spec of the §6 Part B tolerance contract. Expected results are stored as the byte-exact YAML that extract produces. | FR-37, D-31 | Reusing the production marshaller as the golden format means one representation, no second serializer to keep in sync, and the run-twice byte test (AC-1) falls out of the same machinery. |
| **AD-13** | `.github/workflows/static.yml` is replaced **in place**. Exactly one workflow deploys to Pages. | FR-33, D-11, AC-5 | Adding a second workflow on the same trigger sharing `concurrency: group: pages` produces racing deployments. The existing file already has the correct Pages wiring (`static.yml:41` configure-pages, `:43` upload-pages-artifact@v4 with `path: './page'` at `:46`, `:48` `id: deployment`); only `:39` (`go run main.go`) is wrong. |
| **AD-14** | Decommissioning is the **last** step, after the FR-14 metadata port is verified in a rendered `page/index.html`. | FR-32, D-13, AC-6 | `template/template.html:21` is the only copy of the keywords string and `:8-15` the only copy of the GA snippet. Deleting the old template before the port is verified loses the originals. |
| **AD-15** | Superseded decisions from the previous architecture, recorded so no one re-applies them: **old D6** (`static.go`) → AD-03; **old D8** (project-url href fallback) → FR-45 + D-23, projects have no model presence to fall back to; **old D10** (drop `Experience.Roles`) → **reversed by D-22**, roles stay, derived from the title; **old D11** (`languages` list in the model, rendered) → D-26, LANGUAGES is parse-and-ignore in v1; **old D12** (gitignore `page/reydavid_experience.md`) → **reversed by FR-31**, the source is committed and must never be gitignored. | D-22, D-23, D-26, FR-31, FR-40 | — |

---

## 2. Data flow

```
                LOCAL ONLY                                    LOCAL + CI
 ┌──────────────────────────────────┐          ┌────────────────────────────────────┐
 │ page/reydavid_experience.md      │          │        profilegen build            │
 │  hand-maintained SOURCE          │          │                                    │
 │  committed, published verbatim   │          │  model.go  LoadResume + validate   │
 └───────────────┬──────────────────┘          │      │                             │
                 │  profilegen extract         │      ▼                             │
                 ▼                             │  build.go  render → bytes.Buffer   │
 ┌──────────────────────────────────┐          │      │   template/index.html.tmpl  │
 │ parse.go                         │          │      │   PageData{*Resume,         │
 │  line classifier                 │          │      │            Version, MDName} │
 │  section state machine → Doc     │          │      ▼                             │
 └───────────────┬──────────────────┘          │  os.WriteFile on success ONLY      │
                 ▼                             │      │                             │
 ┌──────────────────────────────────┐          │      ▼                             │
 │ extract.go  Doc → Resume         │          │  page/index.html  (the ONLY write) │
 │   tags.go  tech dict, themes,    │          └──────────────┬─────────────────────┘
 │            roles, slugs, chips   │                         │
 │   validation gate (V1..V5)       │                         │
 │   ── fail ─▶ stderr, exit≠0,     │                         │
 │              write NOTHING       │                         │
 └───────────────┬──────────────────┘                         │
                 ▼                                            │
 ┌──────────────────────────────────┐                         │
 │ data/resume.yaml                 │─────────────────────────┘
 │  machine-generated, committed    │
 └──────────────────────────────────┘

 push to main ──▶ .github/workflows/static.yml (REPLACED in place)
                  checkout → setup-go(go.mod) → go test ./... → go run . build
                  → configure-pages → upload-pages-artifact@v4 path ./page
                  → deploy-pages id: deployment  ──▶ https://tecnologer.net

 The uploaded artifact is the whole page/ directory, so CNAME, favicon.ico,
 cover.png AND reydavid_experience.md ship with it (FR-25, FR-41, AC-5).
```

Invariants, each enforced by a test in §10:

- Extract runs **locally only**; CI never invokes it (FR-24, N3).
- Neither command performs network I/O (NFR-02).
- No command creates, modifies or deletes `page/reydavid_experience.md` (FR-39).
- Build writes one file (AD-09).

---

## 3. Architecture drivers

Five factors from the requirements actually shape this design. Everything else is
detail.

| # | Driver | Traced to | Consequence |
|---|---|---|---|
| **DR-1** | The output directory contains the input. `page/` holds both the hand-maintained source and the build output. | §1 fact 1, FR-39, AC-8 | Every write path in the toolchain is enumerated and bounded (AD-09). `make clean` loses one `rm` argument. No wholesale `page/` regeneration anywhere. |
| **DR-2** | The source format is a **tolerance contract**, not a fixed shape, and it must stay deterministic and offline. | FR-6, D-31, NFR-02, AC-1 | The parser is a heuristic line classifier plus a section state machine (§4), not a grammar. Tolerance is pinned by fixtures (AD-12), not by scoring or by an LLM. |
| **DR-3** | Two consumers index `search.suggestions[0]` and one of them is a template action, so an empty list is a **build** failure, not a degraded page. | FR-28, FR-30, `template/index.html.tmpl:443`, `:683` | Validation exists in two places, extract (AD-10) and `LoadResume`, and `performance` is pinned at index 0 (AD-08). |
| **DR-4** | A story JSON key rename produces no build error, no runtime error, and zero search results. | FR-35, D-16 | The five keys are frozen independently of Go field names, and pinned by a test that renders the template and asserts on the emitted JSON (§10.4), because no compiler check exists. |
| **DR-5** | Single maintainer, one resume, order 10 entries and 20 stories, sub-second runtime. | NFR-03, NFR-04 | No caching, no concurrency, no streaming, no plugin points, no config file (N2). Read the whole file into memory, walk it once. |

---

## 4. Parser architecture (`parse.go`)

### 4.1 Shape

Three passes over an in-memory `[]string` of right-trimmed lines. No streaming
(DR-5, the file is 114 lines).

1. **Classify.** Every line gets a `lineKind`, from the line alone. Pure, testable
   in isolation, no state.
2. **Segment.** A section state machine walks the classified lines, opening and
   closing sections and entries. This pass owns the H-3 lookahead ("a date-looking
   line follows within two non-blank lines") and is therefore not purely per-line.
3. **Map** (`extract.go`, not `parse.go`). `Doc` → `Resume`, calling `tags.go`.

The parser **never returns a fatal error for content**. Only I/O errors are fatal.
Everything else is `skip, warn, count` (H-8), collected in `Doc.Warnings`.

### 4.2 The `Doc` model

Only the fields the five section modes consume. No maps anywhere (FR-3).

```
Doc      { Name, RoleLine string; Sections []Section; Warnings []Warning }
Section  { Kind SectionKind; Slug string; Entries []Entry; Bullets []Bullet; Prose []string }
Entry    { Raw string; Segments []string; Start, End string; Bullets []Bullet; Techs []string }
Bullet   { Line int; Text string }
Warning  { Line int; Msg string }
```

`SectionKind` is a closed enum: `Summary, Experience, Education, Certifications,
Skills, Languages, Speaking, OpenSource, Homelab, Unknown`. `Section.Slug` is set
per FR-7: `community` for Speaking, `open source` for OpenSource, and for Homelab it
is derived per-story from the project name in the bullet, not per-section (§4.6).

### 4.3 Line classifier

`classify(line) lineKind`, evaluated in this order. Order is normative: earlier rules
win.

| # | Kind | Test | Notes |
|---|---|---|---|
| 1 | `Blank` | line is empty after trim | closes prose accumulation only, never an entry (H-8) |
| 2 | `Heading` | `reHeading` matches | any level `#` to `######` (H-1) |
| 3 | `Bullet` | `reBullet` matches | any of `*`, `-`, `+`, any indentation (H-5). Indent depth is captured and then discarded: nested bullets flatten into the parent entry |
| 4 | `Italic` | `reItalic` matches | date-line candidate (H-3) |
| 5 | `BoldOnly` | `reBoldOnly` matches | a line that is a single bold span and nothing else |
| 6 | `Boldish` | `reBoldSpan` matches anywhere in the line | entry-line candidate, including partially bold (H-3) |
| 7 | `Plain` | otherwise | may be a date line, a `Technologies:` line, prose, or noise |

Rule 3 before rules 5 and 6 matters: `* **Advanced multithreaded performance:** …`
is a bullet with a bold prefix, not an entry line. `reBullet` requires whitespace
after a single marker character, so `**Bold**` (two `*` with no space) can never
classify as `Bullet`.

### 4.4 Section recognition (H-1)

Keyword table, matched by **prefix**, **case-insensitively**, after stripping `#`
markers and leading/trailing emphasis runs:

| Keyword prefix | SectionKind | Slug |
|---|---|---|
| `summary`, `profile`, `about` | Summary | — |
| `experience`, `employment`, `work history` | Experience | per-entry (company) |
| `education`, `academic` | Education | — |
| `certification`, `certificate`, `course` | Certifications | — |
| `skills` | Skills | — |
| `languages` | Languages | — |
| `speaking`, `talks`, `teaching` | Speaking | `community` |
| `open source`, `open-source` | OpenSource | `open source` |
| `homelab`, `home lab`, `projects`, `side projects` | Homelab | per-story (§4.6) |

Prefix matching is what makes `### HOMELAB / SELF-HOSTED INFRASTRUCTURE`,
`## Homelab`, `### **HOMELAB**` and `### Homelab & Lab Notes` all reach the same
state with no code change. Text after the keyword is discarded.

**Two acceptance paths, with different failure behaviour. This distinction is
normative:**

- A `Heading` line (rule 2) is **always** a section boundary. If its keyword is
  unknown, the state becomes `Unknown`: consume until the next heading, emit one
  warning, never crash (FR-8, H-1).
- A `BoldOnly` line (rule 5) is a section boundary **only if** its stripped text
  prefix-matches a known keyword **and** its letters are upper case. Otherwise it
  falls through to entry recognition. Without this asymmetry the bold-only education
  line on disk today (`page/reydavid_experience.md:102`,
  `**B.S. Software Engineering — Instituto Tecnológico de Culiacán**`) would be
  eaten as an unknown section and the EDUCATION section would render empty.

### 4.5 Identity block (H-2)

Everything before the first section header. Rules run in this order:

1. Any line matching `reContactish` (markdown link, `mailto:`, bare URL, or a bare
   email address) is **discarded before any other rule sees it** (D-24). Contacts
   are template-static chrome; the source carries no contact contract. The
   ordering matters because the real contact line on disk
   (`page/reydavid_experience.md:4`) also carries `•` separators and would otherwise
   be a role-line candidate.
2. `profile.name` is the first surviving non-empty line, with `#` markers and
   emphasis stripped. Bold, plain or H1 all work.
3. If that same line contains a **spaced** separator (`—`, `–`, `-`, `•`, `·`), it is
   split **once**: left is the name, right is `hero.role_line`. This is the D-25
   degenerate shape `# Name — Role Line`, and it terminates identity parsing.
4. Otherwise `hero.role_line` is the next surviving line containing `•`, `—`, `–` or
   `·`, taken **verbatim**: no splitting, no re-ordering, no whitespace collapsing.

Both shapes are pinned by fixtures (`identity-h1`, `identity-three-line`, FR-37).

### 4.6 Entry and bullet handling

**H-3, entry recognition.** A line is an entry line when it classifies as `Boldish`
or `BoldOnly` (and was not consumed as a section header per §4.4) **and** a
date-looking token is found in one of three places:

- inline on the entry line itself;
- on a following `Italic` line, within two non-blank lines;
- on a following `Plain` line, within two non-blank lines.

**Exception, EDUCATION only:** the date is optional. A bold-ish line inside an
Education section opens an entry with empty `Years`. Without this the yearless
education entry on disk today produces zero education entries.

**AD-11 ordering, normative.** For every entry candidate:

1. Run `reDateRange` on the **raw** text (entry line, or the date line if the date
   sits there).
2. Strip the matched span from that text.
3. Strip a trailing parenthetical from what remains (`reTrailingParen`) and
   **discard it**: `Aug 2025 (1-week engagement)` yields the range
   `Aug 2025` → `Aug 2025`, and `(1-week engagement)` is stored nowhere. Recorded
   explicitly so nobody re-adds it as a field later (D-31).
4. Only now split the remainder on spaced dashes (H-4).

Reversing steps 1 and 4 is the defect the previous architecture shipped: the split
pattern matches the dash inside the range itself.

Range endpoints tolerate all three dash variants. `present`, `current`, `now`,
`to date` and `ongoing` all normalize to **exactly** the string `Present`, because
`model.go:80` is `func (e Experience) Current() bool { return e.EndDate == "Present" }`
and `template/index.html.tmpl:458` keys the `current` CSS class off it. A single
date with no range yields `Start == End`.

**H-4, arity-tolerant splitting.** Split the remainder on `reDashSplit` (spaced
dashes only, so `open-source`, `C#-based` and `C#→Go` survive):

| Segments | Mapping |
|---|---|
| 3 | company, title, location |
| 2 | company, title; location empty |
| 1 | company; title and location empty |
| 4+ | first = company, last = location, middle segments rejoined with ` – ` = title |

Arity outside 1..3 warns and continues (FR-36). Dropping an entry is worse than
filing one field imperfectly; FR-27 is the aggregate backstop.

**Education-local rule E-1.** Education entry lines appear in both orderings
(`Institution — Degree` and `Degree — Institution`). After the H-4 split, if exactly
one segment matches `reSchool`, that segment is the institution and the remaining
segments joined with ` · ` are the detail. Otherwise segment 0 is the institution.
Deterministic, one regex, and it is the difference between the page rendering
`Instituto Tecnológico de Culiacán` and rendering `B.S. Software Engineering` as the
school name.

**H-5, bullets.** Any of `*`, `-`, `+`, any indentation. Nested bullets are flattened
into the parent entry: indentation is captured by `reBullet` and then dropped. A
bullet attaches to the most recent open entry; if the current section has no open
entry and is a bullets-only story section (Speaking, OpenSource, Homelab), it
attaches to the section. In Certifications a bullet is a certification string
(verbatim, emphasis stripped). In Skills and Languages bullets are consumed and
discarded (D-26).

**H-6, story titles.** In order:

1. `reBoldTitleInner` (`**Title:** rest`) — group 1 is the title, group 2 the story.
2. `reBoldTitleOuter` (`**Title**: rest`) — same.
3. Otherwise the condensed-first-clause rule of FR-9 / D-27: bullet text up to the
   first `.`, `:` or spaced `—`, trimmed.

The bold-prefix rules win over the condensed clause. Both story text and title have
markdown link syntax unwrapped to label text.

**H-7, technologies line.** `reTechLine`, case-insensitive on the key, optionally
preceded by a bullet marker and optionally bold. Value split on commas, each segment
trimmed, **order and casing preserved verbatim** (D-21). Closes the current entry.
When absent, `Entry.Techs` is empty and technologies are inferred from the entry's
bullets in the mapping pass (§5.3).

**H-8, degradation.** Every rule above, on failure, appends a `Warning{Line, Msg}`
and continues. `Doc.Warnings` is counted into the FR-36 summary line. A parser that
silently produces an empty section is a bug; a parser that warns and produces a
partial section is the contract.

### 4.7 Regex inventory

Normative. Go `regexp` syntax, all compiled once at package level in `parse.go`.
These blocks are fenced, not tabulated, because several contain `|` alternation and
must be copied into Go **without** markdown pipe-escaping. Do not add `\|`
anywhere.

Shared fragments, composed by string concatenation at package level:

```go
const dash     = `[-\x{2013}\x{2014}]`                 // hyphen, en dash, em dash
const monthPat = `(?:jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)[a-z]*\.?`
const datePat  = `(?:` + monthPat + `\s+)?\d{4}`       // "Apr 2024" or "2014"
const openPat  = `present|current|now|to date|ongoing`
```

Matches a markdown heading at any level. Must NOT match `#experience` (no space
after the hashes) or a `--- ` rule.

```go
var reHeading = regexp.MustCompile(`^\s{0,3}(#{1,6})\s+(.+?)\s*#*\s*$`)
```

Trims emphasis runs from the ends of a heading or entry text. Applied as a
replace-with-empty on both anchors. Must NOT touch interior underscores, so
`Pop!_OS` survives.

```go
var reTrimEmph = regexp.MustCompile(`^[*_\s]+|[*_\s]+$`)
```

Matches a line that is one bold span and nothing else, e.g. `**SKILLS**` or
`**B.S. Software Engineering — Instituto Tecnológico de Culiacán**`. Must NOT match
`**Company** – Title – Location` (trailing text) or `**The Translator Pattern:** …`.

```go
var reBoldOnly = regexp.MustCompile(`^\s*\*\*([^*]+)\*\*\s*$`)
```

Matches any bold span anywhere in a line: the partially-bold entry-line test of H-3.

```go
var reBoldSpan = regexp.MustCompile(`\*\*([^*]+)\*\*`)
```

Matches a wholly italic line, the classic date line `*Apr 2024 – Present*`. The
`[^*]` guard is what stops it swallowing bold lines.

```go
var reItalic = regexp.MustCompile(`^\s*\*([^*].*?)\*\s*$`)
```

Matches a date or date range anywhere in the text. Run this **before** any dash
splitting (AD-11). Group 1 is the start, group 2 the end (empty when the entry
carries a single date). Must match `Apr 2024 – Present`, `Mar 2017 - Dec 2019`,
`2009 – 2014`, `Aug 2025`. Must NOT be run against bullet prose, where four-digit
years appear as ordinary numbers.

```go
var reDateRange = regexp.MustCompile(
    `(?i)\b(` + datePat + `)(?:\s*` + dash + `+\s*(` + datePat + `|` + openPat + `))?\b`)
```

Matches a trailing parenthetical to be discarded after the date span is stripped:
`Aug 2025 (1-week engagement)` → the parenthetical. Must NOT match a parenthetical
in the middle of a title.

```go
var reTrailingParen = regexp.MustCompile(`\s*\([^)]*\)\s*$`)
```

Splits an entry-line remainder into segments. Spaced dashes only. Must split
`DefectDojo Inc – Sr Software Engineer – Remote` into three. Must NOT split
`open-source`, `C#-based`, `C#→Go`, `10x–40x`.

```go
var reDashSplit = regexp.MustCompile(`\s+` + dash + `+\s+`)
```

Matches a bullet with any marker and any indentation. Group 1 is the indent (used
only to detect nesting, then discarded), group 2 the text. Must NOT match `---`,
`***`, or a bold line.

```go
var reBullet = regexp.MustCompile(`^(\s*)[-*+]\s+(\S.*)$`)
```

The two H-6 bold-prefix story-title shapes, tried in this order. Must match
`**The Translator Pattern:** every security-vendor connector …` and
`**Advanced multithreaded performance:** delivered talks …`. Must NOT match
`Built [**Tempura**](https://…), an end-to-end IoT monitor …` (the bold span is not
at position 0).

```go
var reBoldTitleInner = regexp.MustCompile(`^\*\*([^*]+?)\s*:\s*\*\*\s*(\S.*)$`)
var reBoldTitleOuter = regexp.MustCompile(`^\*\*([^*]+?)\*\*\s*:\s*(\S.*)$`)
```

Matches an optional `Technologies:` line (H-7), case-insensitive on the key, with an
optional bullet marker and optional bold. Group 1 is the comma-separated value.

```go
var reTechLine = regexp.MustCompile(`(?i)^\s*(?:[-*+]\s+)?(?:\*\*)?technologies(?:\*\*)?\s*:\s*(\S.*?)\s*$`)
```

Marks a line as contact-bearing, so H-2 discards it. Must match
`[rdominguez@tecnologer.net](mailto:…) • [tecnologer.net](http://tecnologer.net) • …`.
Must NOT match a plain role line.

```go
var reContactish = regexp.MustCompile(`(?i)\]\(|mailto:|https?://|[\w.+%-]+@[\w-]+\.[\w.]+`)
```

Marks a line as a role-line candidate (H-2 rule 4): it contains at least one of the
separator characters.

```go
var reSepLine = regexp.MustCompile(`[\x{2022}\x{00b7}\x{2014}\x{2013}]`)
```

Splits the D-25 one-line identity shape `# Name — Role Line`. Split **once**, on the
first spaced separator.

```go
var reNameSplit = regexp.MustCompile(`\s+[-\x{2013}\x{2014}\x{2022}\x{00b7}]\s+`)
```

Identifies the institution segment in an education entry line (rule E-1).

```go
var reSchool = regexp.MustCompile(`(?i)universi|instituto|institut|college|school|tecnol[oó]gico|academy|faculty`)
```

Markdown link and bare URL, for FR-45 story hrefs. Group 1 of `reMDLink` is the
target. `reBareURL` deliberately excludes `)`, `]`, `<` and `>` from the match, and
the result is then right-trimmed of `.,;:!?)]` per P-6, so `see https://x.com/y.`
yields `https://x.com/y`.

```go
var reMDLink  = regexp.MustCompile(`\[[^\]]*\]\((https?://[^)\s]+)\)`)
var reBareURL = regexp.MustCompile(`https?://[^\s<>()\[\]]+`)
```

Matches an Nx multiplier for the `performance` theme (§5.2). Must match `40x`,
`10x–40x`, `2.5x`. Must NOT match `x` alone or a hex literal.

```go
var reNx = regexp.MustCompile(`(?i)\b\d+(?:\.\d+)?x\b`)
```

Matches inline bold inside SUMMARY prose, converted to `<strong>` for
`hero.bio_html` (FR-44). Everything outside these spans is HTML-escaped.

```go
var reBoldInline = regexp.MustCompile(`\*\*([^*]+)\*\*`)
```

---

## 5. Tag inference (`tags.go`)

All of it runs at extract time (AD-06). Dictionaries are ordered slices (AD-07).

### 5.1 Technology dictionary

```
techRule { Tag string; Display string; Aliases []string; Mode matchMode }
matchMode = modeWord | modeExact | modeSymbol
```

Seed set from §7 of the requirements, extended with the census of the
`technologies:` lists in today's `data/resume.yaml` (`:47,55-80,88,96-120,128,136,144`).

- **T-1 `modeWord`, the default.** Case-insensitive, word boundaries both sides:
  `(?i)\b` + `regexp.QuoteMeta(alias)` + `\b`.
- **T-2 `modeExact`, for `go`, `ai`, `c`, `r`.** Case **sensitive**, matched against
  the original-case bullet text, aliases written in their canonical casing (`Go`,
  `AI`, `C`, `R`): `\b` + `regexp.QuoteMeta(alias)` + `\b`. Case-insensitive `\bgo\b`
  tags every bullet containing the English verb ("go live", "go from 1K to 40K"),
  and under frequency ranking (FR-22) those false positives buy chip slots.
- **T-3 `modeSymbol`, for `c#`, `c++`, `.net`.** Custom boundaries on **both** sides:
  `(?i)(^|[^\w#+.])` + `regexp.QuoteMeta(alias)` + `($|[^\w#+.])`.

  Why `\b` fails on both ends, stated because the previous revision only got the
  right-hand side: `\b` asserts a word/non-word transition. On the **right** of `c#`
  or `c++`, the last character (`#`, `+`) is already a non-word character and the
  next is usually a space, so there is no transition and `\b` is false. On the
  **left** of `.net`, the preceding character is a space and the first character of
  the token is `.` — both non-word — so again no transition, and `\b\.net` therefore
  **never matches at all**, in any input. `.NET` must be an explicit test case, not
  just `.net`.

### 5.2 Theme keyword mapping

Ordered slice, exactly the nine rows of §7, in this declaration order (which is also
the emission order of the resulting tags):

| Trigger words | Tag |
|---|---|
| throughput, faster, optimized, plus `reNx` | `performance` |
| rewrote, replaced, migration | `migration` |
| CLI, TUI | `cli` |
| connector, integration | `integrations` |
| LLM, AI, agent | `ai` |
| manager, mentored, 1:1s, interview, speaker, talks | `leadership` |
| open source, contribution, merged | `open-source` |
| debugged, root-caused, traced, bug | `debugging` |
| goroutine, parallel, concurrency | `concurrency` |

Triggers use `modeWord` matching, except `AI` which uses `modeExact` (T-2). `reNx` is
evaluated as a special case inside the `performance` rule, not as a general
`Extra []*regexp.Regexp` field: one pattern in one rule is not a mechanism.

### 5.3 Roles, technologies, slugs

**Roles (P-8, D-22, reversing old D10).** `Experience.Roles` stays in the model.
A role-keyword mapping in `tags.go`, ordered slice, applied to the **title segment
only**: architect ← `architect`; full-stack ← `full stack`, `full-stack`; back-end ←
`software engineer`, `backend`, `back-end`, `go engineer`; people manager ←
`people manager`, `manager`; security tooling ← `security`; consultant ←
`consultant`, `freelance`; speaker ← `speaker`, `teaching`. `Software Architect /
Full Stack Developer` yields `[architect, full-stack]`. The approximation is
accepted: a title yielding no roles renders no badges, guarded by the
`{{if .Roles}}` that already exists at `template/index.html.tmpl:461-463`.

**Per-entry technologies.** If the entry carried an H-7 `Technologies:` line, use it
verbatim (order and casing preserved). Otherwise infer: union of `techRule` matches
across that entry's bullets and its title segment, emitted in dictionary declaration
order with the rule's `Display` casing (`PostgreSQL`, not `postgresql`). An entry
matching nothing yields an empty list, which the new `{{if .Technologies}}` guard
(§7) renders as no chip row.

**Slugs.** `slug(name)`: lowercase, trim, drop corporate stop-words (`inc`, `co`,
`llc`, `ltd`, `s.a.`, `development`, `mx`), join remaining words with `-`.
`DefectDojo Inc` → `defectdojo`, `Puller Tech` → `puller-tech`,
`Alesayi Development Co. (ADCO)` → `alesayi-adco`. Story `source` is the slug with
hyphens rendered as spaces, matching today's `source: puller tech`
(`data/resume.yaml:218`). AC-2 permits divergence from the hand-curated originals.

For Homelab stories the slug comes per-story from the first bold or linked proper
noun in the bullet (`Built [**Tempura**](…)` → `tempura`), falling back to `homelab`
when the bullet names no project (D-20, FR-7).

### 5.4 Per-story tag assembly

Fixed concatenation order, then dedup keeping first occurrence:

1. theme tags, in themeMap declaration order;
2. tech tags, in techDict declaration order;
3. the story's source slug (this is what guarantees the FR-20 floor of ≥1 tag).

### 5.5 Story title dedup (D-27, H-6, P-7)

The dedup map lives in the **mapping pass in `extract.go`**, not in `parse.go` and
not in `tags.go`: it is a property of the assembled story list, and it needs the
source slug, which only the mapper knows.

- **Rune-safe truncation (P-7).** Today's `truncate` at `extract.go:192-199` is
  byte-based: `len(s)`, `s[:n]`, `strings.LastIndex(s[:n], " ")`, and a hardcoded
  `i > 20` floor. Convert to `[]rune`: if `len(runes) <= 60` return as-is; otherwise
  find the last space at or below index 60 in the rune slice and cut there, appending
  `…`; if there is no space at or below 60, cut at 60. The data is full of multi-byte
  characters (`—`, `·`, `í`, `Instituto Tecnológico`) and byte slicing emits invalid
  UTF-8 into the YAML.
- **Dedup.** A `map[string]int` keyed by the lowercased final title, counting
  occurrences. On the second and later occurrence, the story's title becomes
  `Title (source-slug)`. If that composite also collides, append an ordinal:
  `Title (source-slug 2)`, `Title (source-slug 3)`.
- **Determinism.** The map is used only for lookup and counting, never iterated.
  Iteration order is document order, which `parse.go` preserves: sections in file
  order, entries in file order, bullets in file order. Given the same input file,
  "the later one" is always the same story. FR-3 holds.

### 5.6 Suggestions (FR-11, FR-22, FR-28, FR-30, AD-08)

1. `out := []string{"performance"}`. Pinned unconditionally, including when
   `performance` has frequency zero. This alone satisfies the FR-28 floor of ≥1.
2. Candidates: every tag on every story, **minus** every tag that is also a story
   source slug (chips are search themes; the slug is already visible on each story
   card at `template/index.html.tmpl:647`), **minus** `performance` itself.
3. Sort by frequency descending, ties broken by tag ascending in byte order. Total
   order, no unstable comparisons.
4. Append the first 7. Result length is 1..8 by construction.

The 1..8 bound is enforced twice: at extract by check V5 (§6.1) and at load by
`LoadResume` (§6.3). DR-3 is why: `template/index.html.tmpl:443` executes
`{{index .Search.Suggestions 0}}` during template execution, which is a *build*
failure on an empty slice, and `:683` reads `SUGGESTIONS[0].length` in browser JS,
which is a silent runtime failure of the hero typewriter.

---

## 6. The reduced model (`model.go`)

### 6.1 What leaves

Deleted from `model.go` and from `data/resume.yaml` entirely:

| Type or field | Now lives in | Ruling |
|---|---|---|
| `Site` (`Title`, `Description`, `URL`, `SourceURL`), all four fields | template literals | FR-13 |
| `Project`, `Projects []Project` | template literals | D-23, D-05 |
| `Contact`, `Contacts []Contact` | template literals | D-24 |
| `CTA`, `Hero.CTAs` | template literals | FR-13 |
| `Stat`, `Hero.Stats` | template literals | FR-13 |
| `Hero.PromptOutput` | template literal | FR-13, D-25 |
| `Hero.HeadlineHTML` | template literal | FR-13 |
| `Profile.Handle` | template literal (`:403`, `:551`) | FR-13 |

### 6.2 What stays

```
Resume { Profile; Hero; Experience []Experience; Education []Education
         Certifications []string; Search Search }
Profile    { Name }                                        // profile.name
Hero       { RoleLine; BioHTML }                           // hero.role_line, hero.bio_html
Experience { Company; Title; StartDate; EndDate; Location
             Roles []string; Technologies []string }        // Roles per D-22, Techs per D-21
Education  { Institution; Detail; Years }
Search     { Suggestions []string; Stories []Story }
Story      { Title; Story; Source; Href; Tags []string }
```

`func (e Experience) Current() bool { return e.EndDate == "Present" }` stays
unchanged at `model.go:80`; it is consumed by `template/index.html.tmpl:458`.

`Profile` collapses to one field but stays a struct: the YAML key `profile.name` and
the validation message at `model.go:131-133` both reference it, and flattening it to
`Resume.Name` changes the on-disk key for no gain.

### 6.3 Validation

`LoadResume` (`model.go:122-140`) keeps its two existing checks and gains one:

| # | Check | Source |
|---|---|---|
| L1 | `profile.name` non-empty | `model.go:131-133`, unchanged |
| L2 | every story has a title and ≥1 tag | `model.go:134-138`, unchanged |
| **L3** | `1 <= len(search.suggestions) <= 8` | **new**, FR-28, DR-3 |

L3 is what turns `search: {suggestions: []}` into a validation error with a message,
rather than an opaque template execution error from `:443` (FR-28 acceptance).

### 6.4 The frozen story JSON contract (FR-35, DR-4)

`template/index.html.tmpl:597` emits `const STORIES = {{.Search.Stories}};`.
`html/template` serializes the `[]Story` through `encoding/json`, so the JavaScript
at `:618-621` reads `s.tags`, `s.title`, `s.story`, `s.source` and `:664` reads
`s.href`. Those five key names are frozen.

Today's `model.go:114-118` already carries the correct `json:` tags on all five
fields. The freeze means: **a Go field rename must not change the `json:` tag.**

Why this needs a dedicated test and not a code review: renaming `Story.Source` to
`Story.Company` and letting the tag follow produces no build error (Go does not
check template field references at compile time in this direction), no template
error (the field still resolves), no runtime error in the browser
(`s.source` is simply `undefined`), and the only observable symptom is that
palette search silently returns fewer results. §10.4 pins it by rendering the
template against a fixture and asserting on the emitted JSON text.

`href` consumer contract (FR-45): `template/index.html.tmpl:664-665` branches on
`s.href.startsWith('#')`; anchors scroll via `scrollIntoView`, everything else opens
via `window.open(s.href, '_blank', 'noopener')`. So `#experience` is the correct
default and any non-anchor value must be a full absolute URL.

### 6.5 Template data interface (AD-04)

```
PageData {
    *Resume            // §6.2
    Version string     // "v2006.0102.1504", build.go:38, never in the YAML
    MDName  string     // mdFileName, const in build.go per AD-03
}
```

`:443` (`cat {{.MDName}} | grep -i "{{index .Search.Suggestions 0}}"`), `:551`
(footer `{{.Version}}` and `<a href="/{{.MDName}}">`) and `:560` (palette chrome
`cat {{.MDName}} | grep -i`) all keep resolving with zero edits.

Template `FuncMap` after the reduction: `sub` stays (used at `:470`), `html` stays
(used at `:424` for `BioHTML`), **`hasPrefix` is removed** — its only use is `:542`
inside the `{{- range .Contacts}}` block, which becomes static HTML under D-24.

---

## 7. Command design

### 7.1 `profilegen build`

`Build(dataPath, templatePath, outDir string) error`:

1. `LoadResume(dataPath)` — L1, L2, L3.
2. Parse the template with the reduced `FuncMap`.
3. `os.MkdirAll(outDir, 0o755)`.
4. `version := time.Now().Format("v2006.0102.1504")` (`build.go:38`, unchanged).
5. **`var buf bytes.Buffer; tmpl.Execute(&buf, PageData{…})`.** On error: return it,
   write nothing.
6. `os.WriteFile(filepath.Join(outDir, "index.html"), buf.Bytes(), 0o644)`.
7. Print one line: the path written and the version.

This replaces `build.go:40-52` (`os.Create` at `:40`, before `Execute` at `:52`) and
deletes `build.go:56-58` (the `mdPath` / `os.WriteFile` pair that overwrites the
source). Build touches **no other path** in `outDir` (FR-39, AC-8). Flags are
unchanged: `-data data/resume.yaml`, `-template template/index.html.tmpl`,
`-out page` (`main.go:30-32`).

### 7.2 `profilegen extract`

`Extract(inPath, outPath string) error`:

1. Read `inPath`. I/O error is fatal.
2. `parse.go` → `Doc` plus `Doc.Warnings`.
3. Map `Doc` → `Resume`, calling `tags.go` (§5).
4. **Validation gate, on the in-memory `Resume`, before any write** (FR-27, AD-10).
   Each check has a name printed on failure:

   | # | Check |
   |---|---|
   | V1 | `profile.name` is non-empty |
   | V2 | `len(experience) >= 1` |
   | V3 | `len(search.stories) >= 1` |
   | V4 | every story has a non-empty title and `len(tags) >= 1` |
   | V5 | `1 <= len(search.suggestions) <= 8` |

   On any failure: print `extract: validation failed: <check name>: <detail>` to
   stderr, return a non-zero exit, and **do not open `outPath` at all**. The
   previous `data/resume.yaml` stays byte-identical because nothing touched it
   (AC-9). Do not write to a temp file and rename: not opening the target is
   simpler and strictly safer.
5. Marshal with `yaml.Marshal`, prepend the fixed FR-5 header comment (no timestamp,
   no version stamp — it would break AC-1), write `outPath`.
6. Print the FR-36 summary line to stderr:
   `extract: <n> entries, <n> stories, <n> suggestions, <n> warnings → data/resume.yaml`.
   Today's summary at `extract.go:172-173` counts entries, projects and stories and
   reports **no** warnings; the warning count is the new part, and it is what makes a
   silently degraded parse visible.

Flags per D-10 / FR-1: `-in` defaults to `page/reydavid_experience.md` (today
`main.go:38` defaults to `""`), `-out` defaults to `data/resume.yaml` (unchanged).
**Removed:** the `-force` flag (`main.go:40`), its guard (`extract.go:61-66`), and
the "`-in` is required" exit-2 check (`main.go:42-45`). `go run . extract` with no
flags must succeed from a clean checkout.

### 7.3 Makefile

```make
.PHONY: site extract build test serve clean

site:    go run . build
extract: go run . extract          # no IN= variable; defaults are the contract (D-10)
build:   go build -o bin/profilegen .
test:    go test ./...
serve:   site, then python3 -m http.server 8080 in page/
clean:   rm -rf bin page/index.html
```

`clean` drops `page/reydavid_experience.md` from its `rm` list (`Makefile:21` today
destroys the source, FR-39, AC-8). The "One-time:" wording at `Makefile:7` is
removed (FR-19). `.PHONY` gains `test`.

---

## 8. Template change list

Line anchors are against `template/index.html.tmpl` as it stands today (708 lines).
Apply from the bottom up so earlier anchors stay valid.

### 8.1 Static conversions (FR-13, D-05, D-23, D-24)

| Line | Today | Becomes |
|---|---|---|
| `:6` | `<title>{{.Site.Title}}</title>` | literal `Rey David Domínguez · Tecnologer — Senior Go Engineer` (from `data/resume.yaml:7`) |
| `:7` | `content="{{.Site.Description}}"` | literal, from `data/resume.yaml:8-12` |
| `:8` | `og:title` `{{.Site.Title}}` | same literal as `:6` |
| `:9` | `og:description` `{{.Hero.RoleLine}}` | **frozen to the same literal as `:7`.** Decision: `og:description` is site metadata under FR-13, and it should agree with `<meta name="description">` rather than duplicate the role line. `.Hero.RoleLine` stays dynamic at `:423`, where it is resume content |
| `:10` | `og:url` `{{.Site.URL}}/` | literal `https://tecnologer.net/` |
| `:403` | `{{.Profile.Handle}}` | literal `tecnologer` |
| `:421` | `{{.Hero.PromptOutput}}` | literal `rey david domínguez soto` (from `data/resume.yaml:21`) |
| `:422` | `{{html .Hero.HeadlineHTML}}` | literal `I build backends<br>that hold the line<span class="dim">.</span>` (from `data/resume.yaml:22`) |
| `:425-429` | `{{- range .Hero.CTAs}}` block | three literal `<a class="btn …">` elements, from `data/resume.yaml:32-34` |
| `:430-434` | `{{- range .Hero.Stats}}` block | three literal `<div class="stat">` elements, from `data/resume.yaml:36-38` |
| `:485-501` | `{{- range .Projects}}` block inside `<section id="projects">` | three literal `<article class="project">` blocks, from `data/resume.yaml:147-174`, keeping the `status` / `status_label` markup shape at `:488` (`live` ● and `wip` ◐). There is no `{{if .Projects}}` guard today and none is added: the section becomes unconditional static HTML |
| `:540-544` | `{{- range .Contacts}}` block | five literal `<a href>` elements, from `data/resume.yaml:192-196`, each carrying `target="_blank" rel="noopener"` except the `mailto:` one. This is what makes the `hasPrefix` FuncMap entry dead (§6.5) |
| `:551` | `{{.Profile.Handle}}` and `{{.Site.SourceURL}}` | literals `tecnologer` and `https://github.com/tecnologer/portfolio`. **`{{.Version}}` and `{{.MDName}}` on this line stay dynamic** |

### 8.2 Guards

- **`{{if .Technologies}}` — ADD.** `:464-472` is today an unguarded
  `{{- $n := len .Technologies}}` followed by `<div class="techs">`, the range, and
  the `+N more` button. An entry with zero technologies currently renders an empty
  `<div class="techs">`. Wrap `:464-472` in `{{- if .Technologies}} … {{- end}}`
  (D-21, N4).
- **`{{if .Roles}}` — ALREADY EXISTS** at `:461-463`. D-22 says "the template gains
  an `{{if .Roles}}` guard"; it already has one, verified this session. **No work.**
  Listed here only so nobody adds a second one.

### 8.3 `<head>` metadata port (FR-14, AC-6)

Insert after `:11` (`<meta property="og:type" content="website">`), all ported from
`template/template.html`:

- the Google Analytics gtag snippet with id `G-2ZHSBVSQV5` (`template/template.html:7-15`);
- `<link rel="icon" href="favicon.ico">` (`:6`);
- `<meta property="og:image" content="cover.png">` (`:19`);
- `<link rel="canonical" href="https://tecnologer.net">` — **apex, not `www`**
  (D-03). `template/template.html:24` says `https://www.tecnologer.net`; that is the
  one value deliberately **not** ported verbatim, because `page/CNAME` and the served
  host are the apex;
- `<meta name="author" content="Tecnologer">` (`:22`);
- `<meta name="keywords" content="software, architect, golang, csharp, c#, mssql, postgres">`
  — verbatim from `template/template.html:21`, character for character.

### 8.4 Untouched

`:423` `.Hero.RoleLine`, `:424` `.Hero.BioHTML`, `:443` `.MDName` +
`{{index .Search.Suggestions 0}}`, `:457-474` `.Experience` (except the new guard),
`:514-520` `.Education`, `:525-527` `.Certifications`, `:551` `.Version` + `.MDName`,
`:560` `.MDName`, `:597` `.Search.Stories`, `:598` `.Search.Suggestions`. No CSS, no
layout, no JavaScript changes (N4).

---

## 9. CI (`.github/workflows/static.yml`, replaced in place)

Per FR-33 / D-11 / AD-13. Keep the file name and path; exactly one workflow deploys
to Pages.

Preserved from the current file: `on: push: branches: [main]` plus
`workflow_dispatch` (`:4-10`); `permissions: contents: read, pages: write,
id-token: write` (`:13-16`); `concurrency: group: "pages", cancel-in-progress: false`
(`:20-22`); `actions/checkout@v4` (`:33`); `actions/setup-go@v5` with
`go-version-file: 'go.mod'` (`:35-37`); `actions/configure-pages@v5` (`:41`);
`actions/upload-pages-artifact@v4` with `path: './page'` (`:43-46`);
`actions/deploy-pages@v4` with `id: deployment` (`:47-49`); the
`environment: {name: github-pages, url: ${{ steps.deployment.outputs.page_url }}}`
block (`:27-29`).

Two changes:

1. `:38-39` `- name: Generate pages / run: go run main.go` becomes `run: go run . build`.
   `go run main.go` is why CI is red: it names one file and skips the rest of
   the build, so the command fails on undefined symbols (pre-AD-01a, when the
   generator was also `package main` at the root). (`go build ./...`
   currently succeeds, exit 0 — the repo compiles; only the CI invocation is wrong.)
2. A `- name: Test / run: go test ./...` step is inserted **before** the generate
   step. A failing test or a failing build fails the workflow before any artifact is
   uploaded (FR-26, no partial deploys).

No extract step, ever (FR-24, N3). No secrets. No network access needed beyond the
actions themselves.

**FR-41 consequence, stated because it is easy to miss:** the artifact is the whole
`page/` directory, so `page/reydavid_experience.md` must be **committed** for it to
land in the artifact and serve at `https://tecnologer.net/reydavid_experience.md`.
It must never be added to `.gitignore`. `.gitignore:7` already reads `page/*.html`,
which covers `page/index.html` and nothing else — correct as-is, **not** a change.

One-time repo setting, outside this repo's files: Pages source must be
"GitHub Actions". Note it in the README.

---

## 10. Test architecture

All tests are in `package portfolio` alongside the code, under
`internal/portfolio` (AD-01a). `testdata/` sits beside them, because `go test`
runs with CWD set to the package directory; the three repo-root inputs
(`page/reydavid_experience.md`, `template/index.html.tmpl`, `data/resume.yaml`)
are reached through the `../../`-prefixed consts in `paths_test.go`. `t.Parallel()` on every
test; no shared mutable state exists.

### 10.1 Golden fixtures: the executable spec (FR-37, D-31, AD-12)

Location on disk:

```
testdata/
├── drift/
│   ├── heading-h2/          in.md  want.yaml
│   ├── heading-plain/       in.md  want.yaml
│   ├── section-suffix/      in.md  want.yaml
│   ├── date-plain/          in.md  want.yaml
│   ├── date-inline/         in.md  want.yaml
│   ├── date-parenthetical/  in.md  want.yaml
│   ├── bullet-markers/      in.md  want.yaml
│   ├── bullet-indented/     in.md  want.yaml
│   ├── entry-partial-bold/  in.md  want.yaml
│   ├── entry-two-segment/   in.md  want.yaml
│   ├── identity-h1/         in.md  want.yaml
│   ├── identity-three-line/ in.md  want.yaml
│   └── tech-lowercase/      in.md  want.yaml
└── malformed/
    └── no-experience/       in.md          (AC-9 input, no want.yaml by design)
```

`want.yaml` is the **byte-exact output of extract**, produced by the production
marshaller. A `-update` flag on the test regenerates them; regenerating is a
reviewable diff, which is the point.

The **baseline fixture is not a copy.** `TestBaseline` reads the committed
`page/reydavid_experience.md` directly. FR-37 requires the baseline in the suite, and
reading the real file rather than a snapshot means the fixture can never drift from
the source, and the same test doubles as the AC-2 ground-truth test (§10.3).

The nine FR-37 bullets map to the drift directories one to one:

| FR-37 bullet (REQUERIMENTS.md:341-348) | Fixture |
|---|---|
| `##` headings instead of `###` | `heading-h2` |
| a plain (non-bold) section heading | `heading-plain` |
| a section name with a suffix after the keyword (H-1) | `section-suffix` |
| a plain date line instead of an italic one | `date-plain` |
| a date on the entry line itself | `date-inline` |
| a range with a trailing parenthetical (`Aug 2025 (1-week engagement)`) | `date-parenthetical` |
| `-` and `+` bullet markers | `bullet-markers` |
| indented bullets | `bullet-indented` |
| a partially bold entry line (`**Company** – Title – Location`) | `entry-partial-bold` |
| a two-segment entry line | `entry-two-segment` |
| identity block in the D-25 H1 shape | `identity-h1` |
| identity block in the three-line shape (H-2) | `identity-three-line` |
| a lowercase `technologies:` line (H-7) | `tech-lowercase` |

**The rule, and it is a rule and not a guideline:** adding support for new drift
means adding a fixture. A tolerance the suite does not pin is not part of the
contract. A parser change that breaks a fixture is a **contract break**, not a test
to update: the fixture wins unless the requirements change first.

### 10.2 AC-1, determinism

`TestExtractDeterministic`: run `Extract` twice on `page/reydavid_experience.md` into
two temp paths, `bytes.Equal` the results, and assert `strings.Contains(out, "TODO")`
is false. Caveat carried from NFR-01: go-yaml's quoting and line-folding decisions
are part of the byte output, so this proves determinism **within** a dependency
version, not across upgrades. `go.mod:5` is a version pin, not an immutability
guarantee.

### 10.3 AC-2, ground truth

`TestBaseline`, against the real committed source:

- exactly the expected number of experience entries, with companies matching
  `Noodle, DefectDojo, Puller Tech, Pentalog, Ubilogix, Alesayi, RedRabbit MX`
  in file order;
- exactly the expected number of stories;
- every story has ≥1 tag and a non-empty title;
- `search.suggestions[0] == "performance"` and `1 <= len(suggestions) <= 8`.

**The two expected counts are blocked.** REQUERIMENTS.md AC-2 asserts 7 entries and
18 stories. The file on disk today yields 6 entries and 42 story-producing bullets.
See conflicts C-2 and C-3 in §12. The test is written with the counts as named
constants so the ruling is a one-line change; the assertion list above is otherwise
final.

### 10.4 FR-35, the frozen story JSON contract (DR-4)

`TestStoryJSONKeys`: render `template/index.html.tmpl` against a `PageData` carrying
one synthetic story, extract the `const STORIES = …;` line from the output, and
assert the emitted object contains exactly the keys `title`, `story`, `source`,
`href`, `tags`. This is the only mechanism that catches a Go field rename, because
such a rename produces no build error, no template error and no browser error
(§6.4).

### 10.5 AC-4, round trip

`TestExtractThenBuild`: `Extract` the real source into a temp YAML, then `Build` from
that YAML into a temp out-dir, asserting both succeed with no manual step. This is
also the regression test for L1/L2/L3 agreeing with V1..V5: if extract can emit a
YAML that `LoadResume` rejects, this test fails.

### 10.6 AC-9, no destructive extract

`TestExtractMalformedLeavesOutputIntact`: copy a known-good `data/resume.yaml` to a
temp path, run `Extract` with `testdata/malformed/no-experience/in.md` (the EXPERIENCE
heading renamed), assert a non-zero result, assert the failed check name appears on
stderr, and assert the temp YAML is **byte-identical** to what it was before.

### 10.7 AC-8 / FR-39, the source survives

`TestBuildWritesOneFile`: hash `page/reydavid_experience.md`, run `Build` into
`t.TempDir()`, assert the temp dir contains exactly one entry named `index.html`, and
assert the source hash is unchanged. Cheap, and it is the only automated guard
against the `build.go:56-58` class of defect ever returning.

### 10.8 Unit tests

`parse_test.go`: the line classifier table (§4.3 rules in order), H-3 date/dash
ordering with an inline range and an education year span, H-4 arity 1/2/3/4+.
`tags_test.go`: T-2 (`Go` matches, `go live` does not), T-3 with `.net`, `.NET`,
`C#`, `C++`, `reNx` on `40x` and `10x–40x`, suggestion tie-breaking, slug
stop-words, rune-safe truncation on a 60+ rune string containing `—` and `í`.

---

## 11. Sequencing and file-by-file plan

Execution order. Do not reorder S0-1 or the decommission step.

### S0-1 — commit the untracked baseline (D-12, FR-24, AC-5)

Verified `??` set this session. Commit **in**:

`.claude/ARCHITECTURE_REVIEW.md`, `.claude/REQUERIMENTS_REVIEW.md`, `CLAUDE.md`,
`build.go`, `data/` (contains only `resume.yaml`), `extract.go`, `mdout.go`,
`model.go`, `page/reydavid_experience.md`, `template/index.html.tmpl`.
Plus the already-staged-then-modified `.claude/ARCHITECTURE.md` and
`.claude/REQUERIMENTS.md`, and the modified tracked `Makefile`, `README.md`,
`go.mod`, `go.sum`, `main.go`.

Explicit rulings on the two ambiguous items:

- **`.claude/ARCHITECTURE_REVIEW.md` and `.claude/REQUERIMENTS_REVIEW.md`: IN.**
  `.claude/REQUERIMENTS.md:5-6` cites both by path as the source of D-01 to D-31. A
  cited document that is not in the repo is a dangling reference, and the decision
  log is the audit trail for this whole redesign.
- **`Profile.pdf`: OUT.** It is a binary rendering of the same content the source
  markdown carries, no command reads it, and it is not part of any artifact. Adding
  a binary blob to a repo whose `page/` directory is published buys nothing. If it
  is wanted as a downloadable CV, that is a separate decision with a separate
  destination (`page/`), and FR-41's publishable-content constraint applies.

`page/reydavid_experience.md` **must** be committed here: FR-31, and FR-41 needs it
in the Pages artifact.

### S0-2 — the hand-reformatted source (FR-43, D-29)

A **content** deliverable, not a code deliverable: authored and reviewed by David,
carrying the 18 curated story texts of `data/resume.yaml:211-395` over as achievement
bullets, grouped under their entries and their FR-7 sections. The parser is written
**against** this file, not the other way round.

REQUERIMENTS.md §6 Part A carries a status flag (`:386-390`) saying the reformat has
not landed and that Part A must be re-read against the committed file once it does.
**That flag is now stale: the reformat has substantially landed** (the file on disk is
recruiter-format, 114 lines, `### SECTION` headers, bold entry lines, italic date
lines, achievement bullets). It has **not** fully landed against AC-2. Part A must be
re-read and corrected, and conflicts C-1 to C-4 in §12 must be ruled on, before parser
work starts. A Part A mismatch is a documentation defect; Part B is what the parser
implements.

Also part of S0-2, per FR-31 / D-30: add the `<!-- … -->` block at the top of the
source carrying the source-of-truth notice and the `curl … | grep` hint. Note this is
an **addition**, not a replacement: the `Generated by profilegen …` line the
requirements describe at §1:66 is not present in the file today (verified: zero
matches for `Generated by profilegen`).

### S1 — model and build

| File | Change |
|---|---|
| `model.go` | Delete `Site`, `Project`, `Contact`, `CTA`, `Stat` and the fields listed in §6.1. Add check L3 to `LoadResume`. Keep the five `json:` tags on `Story` untouched (FR-35). Update the stale doc comment at `:10-12` ("hand-edited source of truth", "one-time scaffolder") |
| `build.go` | Add `const mdFileName = "reydavid_experience.md"` (moved from `mdout.go:8-11`, AD-03). Add the named `PageData` type (AD-04). Replace `:40-52` with buffer-then-write (AD-09). Delete `:56-58`. Drop `hasPrefix` from the FuncMap. Update the doc comment at `:12-15`, which still claims two artifacts |
| `mdout.go` | **DELETE the file** (AD-03, FR-40) |
| `template/index.html.tmpl` | §8 |
| `main.go` | Extract flag defaults per D-10; delete `-force` (`:40`) and the required-`-in` check (`:42-45`); rewrite the doc comment (`:1-12`) and `usage()` (`:58-67`), removing "one-time" at `:10` and `:63` and the "+ page/reydavid_experience.md" claim at `:12` and `:62` (FR-19) |

At this point `go build ./...` must still pass and `go run . build` must render from
the existing `data/resume.yaml`, minus the removed fields.

### S2 — parser and inference

| File | Change |
|---|---|
| `parse.go` | **NEW.** §4: regex inventory, classifier, section state machine, `Doc` |
| `tags.go` | **NEW.** §5: techDict, themeMap, roleMap, slug, normalize, per-story tags, per-entry technologies, suggestions |
| `extract.go` | **REWRITE.** `Doc` → `Resume` mapping, story assembly and dedup (§5.5), the V1..V5 gate (§7.2), fixed FR-5 header, deterministic marshal, FR-36 summary. Delete `reEntry`/`reDates`/`reSplit` (`:32-36`), `sectionKind` (`:38-56`), the `-force` guard (`:61-66`), the TODO scaffolding (`:73-87`), `orTODO` (`:177-182`). Convert `truncate` (`:192-199`) to rune-safe (P-7). Keep `titleCase` (`:184-190`) as the `Present` normalizer |
| `testdata/` | **NEW.** §10.1 |
| `parse_test.go`, `tags_test.go`, `extract_test.go`, `build_test.go` | **NEW.** §10 |
| `go.mod` | `go mod tidy` to drop the wrong `// indirect` at `:5` (NFR-01) |

### S3 — CI and docs

| File | Change |
|---|---|
| `.github/workflows/static.yml` | §9, replaced in place |
| `Makefile` | §7.3: `clean` loses `page/reydavid_experience.md` (`:21`), `extract` loses `IN=`, `:7` loses "One-time:", `.PHONY` gains `test` (FR-19, FR-39) |
| `README.md` | Rewrite the diagram at `:7-9` (it shows build emitting both artifacts), `:19` ("One-time bootstrap"), `:30` ("Edit `data/resume.yaml` — it is the source of truth from now on"), the layout block at `:44-53` (mentions `mdout.go`), and the deploy note at `:55-59`. Document the flow: edit `page/reydavid_experience.md` → `make extract` locally → commit both files → CI builds and deploys. Add the D-28 note that an inferred tag need not occur literally in the prose, so a chip and the advertised `grep` can disagree (FR-29 withdrawn, divergence accepted). Add the Pages-source repo setting note (AC-7) |
| `.gitignore` | **NO CHANGE.** `:7` already reads `page/*.html`. Do not add `page/reydavid_experience.md` (FR-31 forbids it) |

### S4 — decommission, LAST (FR-32, D-13, AD-14)

Only after a rendered `page/index.html` has been verified to contain the GA snippet,
the favicon link, `og:image`, the apex canonical, the author meta and the keywords
meta (AC-6).

| Target | Tracked? | How |
|---|---|---|
| `about/certification.go`, `about/contact.go`, `about/education.go`, `about/experience.go`, `about/me.go`, `about/projects.go` (**six** files, verified; the review said four and was wrong) | tracked | `git rm -r about/` |
| `template/template.html` | tracked | `git rm` |
| root `CNAME` (contains `tecnologer.net`, no trailing newline; the duplicate of `page/CNAME`, FR-17) | tracked | `git rm` |
| `page/certification.html`, `page/contact.html`, `page/education.html`, `page/experience.html`, `page/projects.html`, and `page/index.html` | **untracked**, matched by `.gitignore:7` | **plain `rm` on disk**, not `git rm`. There is nothing in the index to remove. `page/index.html` is regenerated by the next build |

FR-32 acceptance afterwards: `go build ./...` succeeds, and `page/` contains
`index.html`, `reydavid_experience.md`, `CNAME`, `favicon.ico`, `cover.png` and
nothing else.

---

## 12. Unresolved conflicts between the requirements and the repo

Reported, not guessed. Each blocks the acceptance criterion named.

**C-1 — REQUERIMENTS.md §6 Part A's status flag is stale, and so is §1:66.**
Part A (`:386-390`) says "the reformat of S0-2 / FR-43 has not landed yet, so the file
currently committed is still revision 2's machine-written KV format". It has landed:
`page/reydavid_experience.md` is 114 lines of recruiter format with `### SUMMARY`,
`### EXPERIENCE`, `### SPEAKING & TEACHING`, `### OPEN SOURCE`,
`### HOMELAB / SELF-HOSTED INFRASTRUCTURE`, `### SKILLS`, `### EDUCATION`,
`### CERTIFICATIONS`, `### LANGUAGES`, bold entry lines and italic date lines.
Consequently §1:66's claim that the file carries a `Generated by profilegen
v2026.0806.0639 …` header is false: the file has no such line (zero grep matches), so
FR-31 / D-30 is an **addition**, not a replacement. Part A needs re-reading and
correcting per its own instruction.

**C-2 — AC-2's 7 experience entries is unreachable: `Noodle` is absent.**
The committed source has six EXPERIENCE entries: DefectDojo Inc (`:12`), Puller Tech
(`:27`), Pentalog (`:33`), Ubilogix Inc (`:44`), Alesayi Development Co. (ADCO)
(`:53`), RedRabbit MX (`:61`). Zero occurrences of "Noodle" anywhere in the file, and
`data/resume.yaml:41-47` carries Noodle with `technologies: [Go]` and no stories.
Either the reformat adds a Noodle entry (a content fix, S0-2) or AC-2 drops to 6.
**Needs David's ruling.** No parser design can conjure the entry.

**C-3 — AC-2's 18 stories versus 42 story-producing bullets.**
FR-9 is unconditional: "Every achievement bullet under an EXPERIENCE entry or an FR-7
optional section becomes one story." The committed source has 57 bullets total
(verified), of which 30 are under EXPERIENCE entries, 4 under SPEAKING & TEACHING, 3
under OPEN SOURCE and 5 under HOMELAB: **42 story-producing bullets**. The remaining
15 are SKILLS (10) and CERTIFICATIONS (5), neither of which produces stories. So
FR-9 applied to the file yields 42 stories, not 18. Nothing in the requirements
defines a bullet-selection rule, and inventing one would be curatorial, not
mechanical, which D-31 forbids. Options, all needing a ruling: (a) AC-2's count
becomes 42; (b) the reformat prunes the source to 18 story bullets, losing resume
content; (c) FR-9 gains an explicit, mechanical selection rule (there is no obvious
candidate). **Needs David's ruling.** §10.3 holds the counts as constants pending it.

**C-4 — EDUCATION and `Technologies:` regressions in the committed source.**
Two, both content-side:
(a) EDUCATION (`:100-102`) has one entry,
`**B.S. Software Engineering — Instituto Tecnológico de Culiacán**`, with **no year
span** and degree-first ordering, and `Universidad de Guadalajara`
(`data/resume.yaml:180-182`) is gone. Rule E-1 (§4.6) recovers the institution from
the wrong-order line and the EDUCATION date exception keeps the entry alive, but
`template/index.html.tmpl:518` will render an empty `<small>{{.Years}}</small>` and
the second institution cannot be recovered.
(b) The source contains **zero** `Technologies:` lines (verified), so every entry
takes the H-7 inference path. Today's curated lists regress hard: DefectDojo's 25
entries (`data/resume.yaml:55-80`) will not be recovered from bullet text. D-21
sanctions the inference path and AC-2 says nothing about tech chips, so this is
accepted-by-omission rather than a contradiction, but it is a visible content
regression and should be an explicit acceptance, not a surprise.

**C-5 — Role badges will regress.** D-22 keeps `Experience.Roles` derived from the
title segment. The committed source's titles are terser than the YAML's: the source
says `Sr Software Engineer` for DefectDojo (`:12`) where `data/resume.yaml:50` says
`Senior Software Engineer, Security Tooling`. The `security tooling` and `open source`
role badges at `data/resume.yaml:54` are therefore not derivable. D-22 states "the
approximation is accepted", so this is in scope, recorded here so it is not read as a
parser bug.

**C-6 — a line-count claim I was given did not verify.** The task brief lists
`page/reydavid_experience.md` at 228 lines. It is **114**. Everything else in the
brief's verified list checked out against the repo, including the six `about/` files,
the `{{if .Roles}}` guard already existing at `:461-463`, the absence of any
`{{if .Technologies}}` or `{{if .Projects}}` guard, `.gitignore:7`, `go.mod:5`, and
every cited template line.

---

## 13. Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| The source drifts into a shape no heuristic covers, and extract silently ships a gutted site | medium | Two layers: H-8 warns and counts per line, the V1..V5 gate (§7.2) refuses to write in aggregate, and the FR-36 summary makes a degraded parse visible on every run. AC-9 pins the refusal |
| A tolerance is added in code without a fixture, and the next change silently removes it | medium | AD-12's rule: adding drift support means adding a fixture. Enforced socially, and cheaply visible because `testdata/drift/` is a flat list of named directories |
| `\b` boundary failures on `c#`, `c++`, `.net` produce silently missing tags | high without care | T-3 custom boundaries on both sides (§5.1), with `.NET`, `.net`, `C#`, `C++` as explicit named test cases. The left-side failure is the one that was missed last time |
| `go` matched case-insensitively inflates the tag on the English verb and buys chip slots | high without care | T-2 `modeExact` (§5.1). Also blunted by AD-08: chip 0 is pinned, so the worst case is a wrong chip at index 1-7, not a wrong hero banner |
| Nondeterminism sneaks in and AC-1 starts flapping | low | Ordered slices everywhere (AD-07), no map iteration anywhere in the output path (§4.2, §5.5), total-order sorts, fixed header with no timestamp, `go.mod:5` pinned. Locked by §10.2 |
| A future `build` change reintroduces a write into `page/` and destroys the source | low but catastrophic | §10.7 asserts build writes exactly one file into a temp dir and that the source hash is unchanged. This is the cheapest possible guard against the single worst failure mode in the project (DR-1) |
| A Go field rename silently kills palette search | low but invisible | §10.4. Nothing else catches it (DR-4, §6.4) |
| The FR-14 metadata is lost when `template/template.html` is deleted | low | AD-14: decommission is last, gated on AC-6 verification in a rendered page |
| C-2 / C-3 are ruled on late and the parser has to change shape | medium | The counts live in named constants in one test (§10.3), and neither ruling changes the parser: FR-9 stays one-bullet-one-story either way. A ruling that prunes the source is a content edit, not a code edit |

---

## 14. Evolution notes

Two changes, and only two, are worth pre-describing:

- **If C-3 is ruled as "42 stories".** Nothing structural changes. The palette gets
  more results per query, and `search.suggestions` frequency ranking becomes more
  meaningful because the sample is larger. The only tuning knob likely to matter is
  the theme trigger list (§5.2), which is a one-slice edit.
- **If the source format drifts past what H-1 to H-8 absorb.** The seam is
  `parse.go`'s classifier (§4.3): a new shape is a new classifier rule plus a new
  fixture, and `extract.go` does not change. If instead the drift is semantic (a new
  section that must render), it is a `SectionKind` value, a mapping branch and a
  template block, and `parse.go` barely changes. The split exists for exactly this.

No other future-proofing. There is no second site, no second resume, no second
maintainer, and no requirement that implies any of them (DR-5).
