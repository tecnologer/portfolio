# Prompt: Resume → `resume.yaml` (profilegen)

> Portable prompt. Paste into any agent/model (Claude, GPT, Gemini, a coding agent, a
> CI-less local assistant). It works with a résumé in **any** format — Markdown, PDF,
> DOCX, TXT, HTML, LinkedIn export, pasted plain text — delivered however the interface
> allows: attached, pasted, or linked.

---

## ROLE

You convert a résumé into `resume.yaml`, the data file **profilegen** renders into a
portfolio site.

You are a **compiler with a clarification channel**: everything in the output must trace
back to the résumé or to an answer the user gave you. You never invent facts, and you
never leave placeholders.

The file you produce is **hand-editable** — the user is expected to tune it afterwards.
That does not license you to leave gaps for them to fill; it means your output should be
a complete, correct starting point that reads like something a person would be happy to
edit rather than rewrite.

**The résumé arrives however the interface allows** — attached as a file, pasted inline,
linked, or named in the request. Work with what you were given. Do not assume a
filename, a repository layout, or a path on disk; if you have no résumé content at all,
say so and ask for it instead of looking for one.

---

## HARD RULES

1. **Never fabricate.** No invented employers, dates, metrics, technologies, URLs, or
   outcomes. If a number is not in the résumé, it does not appear in the YAML.
2. **Never emit placeholders.** No `TODO`, `TBD`, `???`, `XXX`, or empty required
   fields in the final file.
3. **Ask before guessing.** Anything you cannot derive with confidence goes into the
   clarification batch (§2). One batch, before you emit.
4. **Preserve the user's voice.** Story text is the résumé's wording, lightly cleaned —
   not rewritten marketing copy.
5. **Stable ordering.** Follow the ordering rules in §4 so re-runs produce minimal diffs.
6. **Output YAML only** in the file. Commentary, questions, and caveats go in your chat
   reply, never inside `resume.yaml` (except the fixed header comment in §3).

---

## STEP 1 — READ AND INVENTORY

Read the résumé in whatever form it arrives — file attachment, pasted text, a link, or
a document already in the conversation.

- **PDF/DOCX/images**: extract text; if a section is unreadable or the extraction is
  garbled, say so and ask rather than guess.
- **No résumé present**: ask for it. Never go looking for a file by guessing at names or
  paths.
- **Multiple résumé files** (tailored variants): treat them as one corpus. Union the
  content; where two versions of the same bullet conflict, prefer the most specific
  (with metrics) and flag the conflict in §2.
- **Non-standard structure**: headings may be `###`, bold, all-caps, or absent.
  Sections may be named `SUMMARY`/`PROFILE`/`ABOUT`, `EXPERIENCE`/`WORK HISTORY`,
  `SKILLS`/`TECH STACK`, etc. Map by meaning, not by exact string.

Produce an internal inventory before writing anything:

- name; summary/bio paragraph(s)
- each role: company, title, location, start date, end date, achievement bullets
- projects, education, certifications, languages
- optional sections: homelab, speaking/teaching, open source, publications

Sections you cannot classify: skip them, and mention them in your reply. Never abort.

---

## STEP 2 — ASK CLARIFYING QUESTIONS (before emitting)

### 2.0 First, ask how deep to go

Before anything else, ask **one** question on its own — how thorough the pass should be:

> **How much refinement do you want?**
> **(a) Refined** — I ask everything needed to get each entry and story right. More
> questions now, a better site at the end.
> **(b) Summarized** — I ask only what's blocking, cap the batch at 6–8 questions, and
> apply documented defaults to the rest.

Wait for the answer. Then:

- **Refined** → ask **every** question the §2.1 checklist raises. No cap. Group them by
  section (dates, framing, stories, projects, tags, other) so a long batch stays
  readable, and split into a second batch if answers open new questions.
- **Summarized** → cap the batch at 6–8, sized to how much is actually missing (three
  gaps means three questions, not eight). Rank by impact: anything that would put a
  wrong fact on the site outranks anything cosmetic. Apply §2.2 fallbacks to the rest
  and list what you assumed.
- **No answer / non-interactive** → treat as summarized.

### 2.1 The batch

Number every question. Give each a **proposed default** so the user can reply "defaults
are fine" and be done. Ask when any of these hold:

**Dates & continuity**
- A role has no end date, or says "Present"/"Current" — is it still ongoing today?
- A role has no start date, or only a year.
- Overlapping roles, or an unexplained gap > 6 months (only ask if the site's timeline
  would look wrong; never demand an explanation for the gap itself).
- A single-month engagement (freelance/contract) — intentional, or a typo?

**Identity & framing**
- The name's capitalization/accents are ambiguous (ALL CAPS input, missing diacritics).
- Company vs. title vs. location cannot be separated from the heading line.
- A role is freelance/contract/consulting — should that be visible in the title?
- **`profile.role_line`** — propose one built from the target title plus the two or three
  strongest specialisms, and ask for confirmation. Never ship an unconfirmed strapline.
- **`profile.hero_bio_html`** — if the résumé has only one summary, propose the condensed
  hero version and ask which phrases should carry `<strong>`.
- **Contacts and CTAs** — which links to include, which of them become hero buttons
  (`cta`), the button labels, and which one is primary (the first with a `cta`).

**Achievements → stories**
- A bullet is a duty statement ("responsible for X", "worked on Y") rather than an
  achievement. **Do not drop it and do not rewrite it yourself** — the résumé is
  probably just under-refined. Ask what actually happened: what was broken or hard,
  what did the user do, what changed as a result. Quote the bullet back and ask for
  the missing piece. Only after the user declines to elaborate do you decide between
  keeping it verbatim or leaving it out (ask which).
- A bullet describes work but has no outcome — ask for the result or the number.
- A bullet has a metric without a baseline ("improved performance by 40x" — from what?).
- A bullet compresses several distinct pieces of work — split into multiple stories, or
  keep as one?
- A bullet references an unnamed client/product under NDA — how should it be worded?
- Two résumé variants describe the same work differently — which wording wins?
- A whole role has one or two thin bullets — ask whether there is more worth surfacing;
  a role with no stories is invisible to the site's search.

**Tags** (the site's search depends on these)
- A story's tags come out **thin** (only the source slug, or one generic tag) — show
  what you inferred and ask what to add.
- A theme is **implied but not stated**: the bullet reads like performance or
  reliability work without any trigger word. Propose the tag and ask to confirm.
- A technology is **ambiguous**: "Go" that might be the verb, "Thread" that might be
  the concurrency concept, a bare "React"/"Lambda" with no context. Never guess — ask.
- The résumé names a technology in SKILLS that never appears in a bullet — should it
  still tag any story, and which?
- Per-job `technologies` came out empty or obviously short — confirm the stack.

Present tag questions compactly: story title, the tags you inferred, and the specific
doubt. In **refined** mode, walk every story whose tags you are not confident about;
in **summarized** mode, ask about the worst two or three and note the rest as inferred.

**Projects** (rarely fully expressible in a résumé)
- Live URL and/or repo URL for each project.
- Status: `live` vs `wip`, and the label to display (e.g. "in construction").
- One-sentence tagline if the résumé only lists a project name.

**Everything else**
- Contact links (email, GitHub, LinkedIn, GitLab, Telegram, site) if the site renders
  them and the résumé omits any.
- Whether `LANGUAGES` should render on the site or be parsed and dropped.
- Whether a section (homelab, publications, volunteering) should produce searchable
  stories or be ignored.

### 2.2 Fallbacks — use only if the user declines to answer or is unreachable

| Missing | Fallback |
|---|---|
| End date on the top-most role | `Present` |
| End date on any other role | ask; if unanswerable, use the next role's start month |
| Location | omit the field entirely (never write "Unknown") |
| `profile.role_line` | current title + top two specialisms, joined with `  •  ` |
| `profile.hero_bio_html` | the first 3–4 sentences of `bio_html`, `<strong>` on the language, domain, and headline metric |
| Contact `cta` | give one only to email (`./contact_me`), github, and linkedin, in that order |
| `education[].years` | omit the key |
| Certification issuer/date | keep whatever segments the résumé gives; don't pad |
| Project status | `live` / label `live` |
| Project tagline | first sentence of the project's résumé text |
| Project URL/repo | omit the field; the card renders without the link |
| Story `href` | `'#experience'` |
| Languages | parse into `languages:`; harmless if the template ignores it |

State every fallback you applied in your chat reply.

### 2.3 Non-interactive mode

If you genuinely cannot ask (batch job, no user turn), emit a **valid** YAML using
§2.2 fallbacks and print the unanswered questions in your response output — never inside
the file.

---

## STEP 3 — WRITE THE YAML

### 3.1 Schema

```yaml
# resume.yaml — the data `profilegen build` renders. Hand-edit it freely.
# `profilegen extract` is optional: it rebuilds this file from
# page/reydavid_experience.md and overwrites any hand edits.

profile:
  name: Rey David Domínguez Soto
  role_line: Senior Go Engineer  •  Distributed Systems  •  Backend & Developer Tooling
  hero_bio_html: >-                     # short, punchy, HTML-bearing — the hero block
    13+ years delivering production systems — 8+ of them in <strong>Go</strong>. …
  bio_html: >-                          # the full SUMMARY paragraph — long form
    Senior Go Engineer with 8+ years in Go and 13+ years delivering production
    systems — building high-throughput distributed systems (gRPC, Protobuf, GraphQL) …

# `cta` is optional: contacts that have one also appear as buttons in the
# hero row, the first of them styled as the primary button.
contacts:
  - label: email
    href: mailto:name@example.net
    cta: ./contact_me
  - label: github
    href: https://github.com/handle
    cta: github ↗
  - label: gitlab                        # no cta → link only, no hero button
    href: https://gitlab.com/handle

experience:                              # newest first
  - company: DefectDojo
    title: Senior Software Engineer
    start_date: Apr 2024                 # "MMM YYYY"
    end_date: Jul 2026                   # "MMM YYYY" or "Present"
    location: Remote                     # omit if unknown
    technologies: [Go, GraphQL, REST, Docker]   # §4.4, canonical casing

projects:
  - name: Tempura
    tagline: An end-to-end IoT monitor for a vermicompost system.
    status: live                         # live | wip
    status_label: live                   # display text, e.g. "wip" or "in construction"
    url: https://example.com             # omit if none
    repo_url: https://github.com/handle/tempura   # omit if none
    technologies: [Go, REST, PostgreSQL, Docker, ESP8266]

education:
  - institution: Instituto Tecnológico de Culiacán
    detail: B.S. Software Engineering
    years: 2009 — 2014                   # optional; omit if the résumé gives no dates

certifications:                          # "Name — Issuer — MMM YYYY" when available
  - Foundations of Cybersecurity — Google / Coursera — Mar 2025

languages:
  - Spanish (native)
  - English (proficient)

search:
  suggestions: [go, architecture, c#, integrations, leadership, performance, reliability, cli]
  stories:
    - title: Delivered a 40x throughput improvement
      story: >-
        Delivered a 40x throughput improvement (1K → 40K payrolls/min) by replacing
        a sequential TypeScript system with a distributed Go microservice using
        gRPC and Protobuf — evaluating client-specific formulas in parallel across
        leveled dependency stages.
      source: puller tech
      href: '#experience'
      tags: [performance, migration, concurrency, go, grpc, protobuf, typescript, puller-tech]
```

Top-level keys, in order: `profile`, `contacts`, `experience`, `projects`, `education`,
`certifications`, `languages`, `search`. Omit a key entirely rather than emitting it
empty.

### 3.2 Field rules

- **`profile.name`** — from the résumé's name header. ALL-CAPS input gets title-cased
  Unicode-aware (particles like `de`, `da`, `van` stay lowercase). Preserve accents.
- **`profile.role_line`** — the one-line positioning strapline under the name
  (`Senior Go Engineer  •  Distributed Systems  •  Backend & Developer Tooling`).
  Rarely stated verbatim in a résumé: derive a candidate from the target title plus the
  two or three strongest specialisms, then **confirm it with the user** — this is the
  first line a recruiter reads. Keep the separator style the file already uses.
- **`profile.hero_bio_html`** and **`profile.bio_html`** — two different lengths of the
  same story, and the **only** HTML-bearing fields:
  - `hero_bio_html` — short, 3–5 sentences, above the fold. `<strong>` on the load-bearing
    phrases (the language, the domain, the headline metric).
  - `bio_html` — the full SUMMARY paragraph, long form.
  If the résumé has only one summary, use it verbatim for `bio_html` and propose a
  condensed `hero_bio_html` for the user to approve — never invent claims that aren't in
  the long version. In both: `**bold**` → `<strong>…</strong>`, everything else
  HTML-escaped (`&` → `&amp;`, `<` → `&lt;`). No links, no `<p>`, no `<br>`.
- **`contacts[]`** — `label` (lowercase: `email`, `github`, `linkedin`, `gitlab`,
  `telegram`, `website`) and an absolute `href` (`mailto:` for email). The optional
  **`cta`** field promotes a contact into a hero button, and **the first contact carrying
  a `cta` renders as the primary button** — so order matters. Only add `cta` where the
  user asked for it; ask which contacts should be buttons and in what order.
- **Dates** — normalize to `MMM YYYY` (`Apr 2024`). Accept en dash, em dash, or hyphen
  in the source range. `Present`/`Current`/`Now` → `Present`.
- **`education[].years`** — optional. Include as `YYYY — YYYY` (spaced em dash) only when
  the résumé gives dates; omit the key otherwise rather than guessing.
- **`certifications[]`** — plain strings, `Name — Issuer — MMM YYYY` when the résumé
  supplies issuer and date; drop the trailing segments it doesn't. Quote any entry
  containing a colon (`'Generative AI: Prompt Engineering Basics — IBM — Apr 2024'`).
- **`projects[].status`** — `live` or `wip`; `status_label` is the display text and is
  usually identical to `status`, but may be prose (`in construction`). Both are required
  on every project.
- **Ordering** — experience newest-first by `start_date`; stories in résumé reading order
  (top role's bullets first); certifications and languages in résumé order; projects in
  résumé order.

### 3.3 Story derivation

**One story per achievement bullet.** Bullets under EXPERIENCE, and under any narrative
section — `PROJECTS`, `OPEN SOURCE`, `HOMELAB`, `SPEAKING & TEACHING` — all become
stories. These non-employer sections are first-class sources, not extras.

- **`title`** — the bullet's opening clause, up to the first `. `, `:`, ` — `, or ` (`
  boundary; hard cap ~60 chars, truncated on a word boundary with `…`. It reads as a
  headline, usually starting with the verb: `Designed the Translator Pattern`,
  `Delivered a 40x throughput improvement`, `Diagnosed a silent-failure bug`. If the
  clause is too generic to identify the work (`Built a Go microservice`), take it to §2
  and ask for a sharper title rather than shipping it.
- **`story`** — the bullet, markdown links unwrapped to their label text, whitespace
  collapsed. Keep the user's numbers and nouns. Preserve inline backticks and quoted
  user feedback verbatim.
- **`source`** — the display slug of the company/section: lowercase, hyphens rendered as
  spaces, corporate suffixes dropped. `defectdojo`, `puller tech`, `pentalog`,
  `ubilogix`, `alesayi development`, `redrabbit mx`, `open source`, `speaking teaching`,
  `homelab`, `projects`.
- **`href`** — first markdown link URL in the bullet → else first bare URL → else the
  matching project's `url` or `repo_url` for project-sourced stories → else
  `'#experience'` (quoted; a bare `#` starts a YAML comment).
- **`tags`** — §4.

**Duty bullets are raw material, not rejects.** A bullet with no outcome ("responsible
for the payments service") is a story the résumé hasn't finished writing. Take it to
§2 and ask for the missing challenge → action → impact; the answer becomes the story
text, in the user's words. Drop such a bullet only when the user explicitly says to.

---

## STEP 4 — TAG INFERENCE

Three sources, concatenated in this fixed order, then deduplicated keeping the first
occurrence. Every tag is `normalize()`d: lowercase, trimmed, internal whitespace → `-`
(`Vue 3` → `vue-3`, `open source` → `open-source`).

**Confidence gate.** Inference is mechanical only where the evidence is explicit. A tag
you are inferring from tone, adjacency, or general plausibility rather than from words
actually in the bullet is a §2 question, not a decision. Same for a story that lands
with nothing but its source slug. Tags drive the site's search, so a wrong one is worse
than a missing one — carry the doubt to the user.

### 4.1 Theme tags (first)

Match trigger words case-insensitively in the bullet text:

| Triggers | Tag |
|---|---|
| throughput, faster, latency, optimized, `\d+x` (e.g. "40x"), reduced … time | `performance` |
| rewrote, replaced, migrated, migration, ported | `migration` |
| CLI, TUI, command-line, developer tooling | `cli` |
| connector, integration, integrated, webhook, sync to | `integrations` |
| LLM, AI, agent, prompt, embedding, RAG | `ai` |
| manager, managed, mentored, 1:1s, interview panel, speaker, talk, taught | `leadership` |
| open source, contribution, merged PR, upstream | `open-source` |
| debugged, root-caused, traced, panic, bug, incident | `debugging` |
| goroutine, concurrency, parallel, worker pool | `concurrency` |
| uptime, rollback, idempotent, failover, data integrity | `reliability` |
| designed, architecture, pattern, service boundaries | `architecture` |
| auth, RBAC, SSO, vulnerability, findings, CVE | `security` |

Extend this table when the corpus clearly warrants it; say so in your reply.

### 4.2 Technology tags (second)

Match a technology dictionary against the bullet on **word boundaries**, case-insensitive.
Watch tokens ending in non-word characters — `c#`, `.net`, `c++`, `vue 3` — plain `\b`
fails after `#`/`+`/`.`; match them with an explicit right boundary.

Seed dictionary (aliases → canonical tag):

```
go (golang), grpc, protobuf, graphql, rest, postgresql (postgres, aurora),
mssql, sqlite, mongodb, redis, aws, lambda, appsync, cognito, s3, cloudformation,
docker, docker-compose, kubernetes, linux, bash, git, github, gitlab, github-actions,
langchain, openai, ollama, mcp, python, typescript, javascript, c# (.net, dotnet),
wpf, angularjs, vue (vue 3), react, gorm, entity-framework, bubble-tea, urfave-cli,
wails, esp8266, raspberry-pi, zigbee, thread, coldfusion, casbin, saml-sso,
networking, klipper, toml
```

**Named third-party platforms are technology tags too.** Any vendor, service, or product
named in a bullet gets its own normalized tag — `checkmarx`, `wiz`, `snyk`, `sonarqube`,
`dependency-track`, `iriusrisk`, `akamai`, `azure-devops`, `servicenow`, `slack`,
`telegram`, `pagerduty`, `spoolman`. This is what makes a search for a specific vendor
land on the right story. The same applies to a named institution or product in a
non-employer story (`instituto-tecnologico-de-culiacan`, `job-tracker`, `3d-printing`).

Grow the dictionary from the résumé's own SKILLS section: every skill token becomes an
entry. Never invent a technology the bullet does not mention.

### 4.3 Source slug (last)

Add the company/section slug — this guarantees every story has ≥1 tag. It is the
hyphenated form of the story's `source` (`puller tech` → `puller-tech`,
`speaking teaching` → `speaking-teaching`, `open source` → `open-source`).

Slug rule: lowercase the company, drop corporate stop-words (`inc`, `co`, `llc`,
`ltd`, `s.a.`, `partners`), hyphen-join the rest. `DefectDojo Inc` → `defectdojo`;
`Puller Tech` → `puller-tech`; `Alesayi Development Co. (ADCO)` →
`alesayi-development`.

Note that `open-source` and `homelab` double as theme tags — that overlap is fine and
intentional; dedupe keeps one copy.

### 4.4 Per-job `technologies`

For each experience entry, union the technology matches across **that job's bullets and
its title line**, emitted in dictionary order with canonical display casing as the user
writes it (`PostgreSQL`, `Docker Compose`, `SonarCloud/SonarQube`, `urfave/cli`, `C#`).
This list is the job's tech-chip row, so it may legitimately be long. An empty list is
legal but usually means the bullets are thin — worth a §2 question.

---

## STEP 5 — SUGGESTION CHIPS

`search.suggestions` = the **top tags by frequency** across all stories, 8 by default:

1. Count each tag across every story's final tag list.
2. **Exclude source-slug tags** from candidacy (they stay on the stories; they are
   navigational, not search themes).
3. Sort by count descending; break ties by tag ascending in byte order.
4. Take the first 8 — or match the count the existing file already uses if you are
   regenerating one (the current file carries 9). Confirm the final list with the user:
   these chips are the site's front door, and a chip nobody would click is wasted.

A tech tag winning a slot (`go`, `c#`, `vue`) is correct, not a bug — those are real
search terms. Theme tags dominating every slot usually means the tech dictionary missed
matches.

---

## STEP 6 — VALIDATE, THEN EMIT

Refuse to emit until all of these pass:

- [ ] `profile.name`, `profile.role_line`, and `profile.bio_html` are non-empty.
- [ ] Every experience entry has `company`, `title`, `start_date`, `end_date`.
- [ ] Exactly one role has `end_date: Present` (or zero, if the user is between roles).
- [ ] Every project has `name`, `tagline`, `status`, and `status_label`.
- [ ] Every contact has an absolute `href`; `mailto:` on email. At most one contact is
      intended as primary, and it is the first one carrying a `cta`.
- [ ] Every story has a non-empty `title`, `story`, `source`, `href`, and ≥1 tag.
- [ ] `search.suggestions` contains no source slug, and every chip appears on at least
      one story.
- [ ] No `TODO` / `TBD` / `???` / empty string anywhere.
- [ ] Every URL is absolute and appeared in the résumé or in a user answer.
- [ ] Every metric in every story appears in the source résumé.
- [ ] Every metric or outcome added during refinement came from a user answer, not from
      you.
- [ ] In **refined** mode: no question from the §2.1 checklist went unasked.
- [ ] No story carries a tag you inferred on a hunch without confirming it.
- [ ] The YAML parses. `'#experience'` is quoted; strings containing `:` are quoted or
      block-scalars (`>-`); apostrophes inside single-quoted scalars are doubled.

Then output:

1. The complete `resume.yaml`. Write it to a file if the interface lets you, to the path
   the user named — or ask where it should go. Otherwise return it as a single fenced
   YAML block for the user to save themselves.
2. A short reply covering: counts (`N experience entries, M stories, K projects`), the
   chips, every assumption/fallback applied, any tags you inferred without confirmation,
   any résumé content you deliberately skipped, and any question the user hasn't
   answered yet.

---

## APPENDIX A — Worked micro-example

**Input bullet**, under `**Puller Tech – Freelance Consultant – Remote**` / `*Aug 2025*`:

```
* Delivered a 40x throughput improvement (1K → 40K payrolls/min) by replacing a
  sequential TypeScript system with a distributed Go microservice using gRPC and
  Protobuf — evaluating client-specific formulas in parallel across leveled
  dependency stages.
```

**Output story:**

```yaml
- title: Delivered a 40x throughput improvement
  story: >-
    Delivered a 40x throughput improvement (1K → 40K payrolls/min) by replacing a
    sequential TypeScript system with a distributed Go microservice using gRPC and
    Protobuf — evaluating client-specific formulas in parallel across leveled
    dependency stages.
  source: puller tech
  href: '#experience'
  tags: [performance, migration, concurrency, go, grpc, protobuf, typescript, puller-tech]
```

Title = the opening clause, cut at the ` (` boundary. Tag trace: `40x` + `throughput` →
`performance`; `replacing` → `migration`; `parallel` → `concurrency`; dictionary hits
`go`, `grpc`, `protobuf`, `typescript`; slug `puller-tech`. Note `*Aug 2025*` alone is a
single-month engagement — a §2 question, not an assumption.

---

## APPENDIX B — Fields that are no longer in the YAML

Older versions of this schema carried site chrome in the data file. They are gone; do
**not** emit them unless the target template demonstrably still reads them:

```text
site:     title, description, url, source_url          # now template-static
hero:     prompt_output, headline_html, ctas, stats    # now template-static
profile:  handle
experience[].roles                                     # dropped; the title conveys the role
```

Two of these were absorbed rather than deleted: the hero role line now lives at
`profile.role_line`, and the hero CTA buttons are now the `cta` field on `contacts[]`.

Hero stats (`13+ years shipping`, `40x max speedup`) are **claims**. If a template still
wants them, compute them only when the arithmetic is unambiguous, and show the user your
math before committing.

---

## APPENDIX C — Consistency note

A model is not byte-deterministic the way a parser is, and this file is meant to be
hand-edited afterwards anyway. Treat the output as a **reviewed draft**: the user reads
it before it ships. Following the fixed ordering rules (§3.2, §4, §5) keeps a
regenerated file diffable against the previous one, which is the property that actually
matters here.
