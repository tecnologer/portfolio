# Portfolio

A self-referential static portfolio generator built in Go. Professional data is written as Go source files in the `about/` package, which the generator reads, syntax-highlights, and renders as HTML pages with a Dracula dark theme.

Deployed to [tecnologer.net](https://www.tecnologer.net) via GitHub Pages.

## How it works

The generator (`main.go`) reads `.go` files from the `about/` package, applies syntax highlighting via [chroma](https://github.com/alecthomas/chroma), linkifies URLs/emails/cross-references, and renders each file into `page/` using the HTML template.

The Go source files serve a dual purpose — they define the data **and** are the displayed content:

- **me.go** — Personal profile, bio, and links to other sections
- **experience.go** — Work history (DefectDojo, Pentalog, Ubilogix, ADCO, RedRabbit MX)
- **education.go** — Academic background
- **certification.go** — Professional certifications
- **contact.go** — Contact channels (Email, LinkedIn, GitHub, GitLab, Telegram, website)

## Usage

```bash
make run    # Generate HTML pages: go run main.go
make build  # Build binary: go build -o portfolio main.go
```

Output is written to `page/` with auto-generated version timestamps.
