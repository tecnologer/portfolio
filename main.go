// Command profilegen renders data/resume.yaml into the tecnologer.net
// landing page.
//
// Usage:
//
//	profilegen build   [-data data/resume.yaml] [-template template/index.html.tmpl] [-out page]
//	profilegen extract [-in page/reydavid_experience.md] [-out data/resume.yaml]
//
// Typical flow:
//  1. edit page/reydavid_experience.md
//  2. `profilegen extract` (local only) regenerates data/resume.yaml
//  3. commit both
//  4. `profilegen build` renders page/index.html
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tecnologer/profilegen/internal/portfolio"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "build":
		fs := flag.NewFlagSet("build", flag.ExitOnError)
		data := fs.String("data", "data/resume.yaml", "path to the YAML source of truth")
		tmpl := fs.String("template", "template/index.html.tmpl", "path to the HTML template")
		out := fs.String("out", "page", "output directory")
		_ = fs.Parse(os.Args[2:])

		fail(portfolio.Build(*data, *tmpl, *out))

	case "extract":
		fs := flag.NewFlagSet("extract", flag.ExitOnError)
		in := fs.String("in", "page/reydavid_experience.md", "path to the resume .md to extract from")
		out := fs.String("out", "data/resume.yaml", "where to write the YAML")
		_ = fs.Parse(os.Args[2:])

		fail(portfolio.Extract(*in, *out))

	case "-h", "--help", "help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "profilegen: unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `profilegen — yaml-driven portfolio generator

commands:
  build     render data/resume.yaml → page/index.html
  extract   scaffold data/resume.yaml from a free-form resume .md

run "profilegen <command> -h" for flags.
`)
}

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "profilegen:", err)
		os.Exit(1)
	}
}
