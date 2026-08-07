# profilegen

Generator for [tecnologer.net](https://tecnologer.net) — successor to the
Go-source-as-content engine. `data/resume.yaml` is what the site is built
from, and you can edit it by hand. `profilegen extract` is an optional
shortcut that (re)generates it from `page/reydavid_experience.md`, a
recruiter-format resume in markdown — it overwrites hand edits, so skip it
once you maintain the YAML directly:

```
page/reydavid_experience.md ──► profilegen extract ──┐  (optional)
                                                     ▼
                                                 data/resume.yaml ──► profilegen build ──► page/index.html
                                                          ▲
any resume (pdf/docx/…) ──► LLM + resume_yaml_prompt.md ──┘  (optional)
```

The markdown file also makes the site's search command literally true — it is
served as-is at `page/reydavid_experience.md`:

```console
$ curl -s https://tecnologer.net/reydavid_experience.md | grep -i performance
```

Note: story tags/chips shown on the site are **inferred** from the prose, not
copied verbatim. A chip can legitimately not appear as that literal word in
the markdown, so the rendered chip and the `grep` result above can diverge —
this is accepted, not a bug.

## Workflow

1. **Edit `data/resume.yaml`** — it is what gets rendered. `build` only
   complains when it is missing or has no `profile.name`.

   Or, to start from the markdown instead: edit
   `page/reydavid_experience.md` (`### SECTION` headers, bold entry lines,
   italic date lines, achievement bullets) and regenerate:

   ```console
   $ make extract        # or: go run . extract — overwrites data/resume.yaml
   ```

   `-in` defaults to `page/reydavid_experience.md`, `-out` to
   `data/resume.yaml`. This step is never run in CI — commit both files
   yourself.

   > **Tip — starting from a resume that isn't this markdown format?**
   > `extract` only parses `page/reydavid_experience.md`'s shape. For a PDF,
   > DOCX, LinkedIn export, or any other resume, hand
   > [`resume_yaml_prompt.md`](resume_yaml_prompt.md) to an LLM along with
   > your resume — it is a self-contained prompt that asks its clarifying
   > questions and emits a complete `data/resume.yaml`. Review the result,
   > then go to step 2.

2. **Build and preview:**

   ```console
   $ make site        # or: go run . build   (writes page/index.html)
   $ make serve       # preview at localhost:8081  (PORT=8080 make serve)
   ```

3. **Commit both `page/reydavid_experience.md` and `data/resume.yaml`**, then
   push. CI runs `go test ./...` and `go run . build`, then deploys `page/`.

## Layout

```
├── data/resume.yaml             # ← what the site is built from; hand-editable
├── page/reydavid_experience.md  # optional markdown source for `extract`
├── resume_yaml_prompt.md        # LLM prompt: any resume format → data/resume.yaml
├── template/index.html.tmpl     # page design (html/template)
├── page/                        # generated output (deploy this to GitHub Pages)
├── main.go                      # CLI entry: build | extract (flags only)
└── internal/portfolio/          # the generator
    ├── model.go                 # YAML schema
    ├── build.go                 # yaml → html
    ├── parse.go                 # markdown → Doc
    ├── tags.go                  # tag/technology/suggestion inference
    ├── extract.go               # Doc → resume.yaml
    └── testdata/                # golden fixtures (`make golden` regenerates)
```

## Deploy

CI (`.github/workflows/static.yml`) runs `go test ./...`, then
`go run . build`, then deploys the `page/` directory to GitHub Pages. `CNAME`
lives in `page/` and is tracked as-is.

One-time repo setting: in this repo's Settings → Pages, the source must be
set to "GitHub Actions".

## Notes

- Uses `github.com/goccy/go-yaml` (github-hosted). If you prefer
  `gopkg.in/yaml.v3`, the struct tags are compatible — swap the import in
  `internal/portfolio/model.go` and `internal/portfolio/extract.go` and run `go mod tidy`.
- Story `href` may be `#experience`/`#projects` (scrolls) or a full URL (opens
  in a new tab).
- The footer version stamp (`v2026.0806.1234`) is generated at build time,
  matching the old engine's scheme.
