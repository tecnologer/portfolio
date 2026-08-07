package portfolio

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"
)

// mdFileName is referenced by the palette command on the site
// (`$ cat reydavid_experience.md | grep -i "..."`), so the extracted file
// must keep this exact name.
const mdFileName = "reydavid_experience.md"

// Build renders data/resume.yaml into <out>/index.html, the landing page.
func Build(dataPath, templatePath, outDir string) error {
	resume, err := LoadResume(dataPath)
	if err != nil {
		return err
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(templateFuncs()).ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir %s: %w", outDir, err)
	}

	// Version stamp: major version + build timestamp, e.g. v2 (build 20260807.0120).
	// Built with Sprintf, not Format — a literal "2" in the layout is the day-of-month.
	version := fmt.Sprintf("v2 (build %s)", time.Now().Format("20060102.1504"))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, PageData{resume, version, mdFileName}); err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	htmlPath := filepath.Join(outDir, "index.html")
	if err := os.WriteFile(htmlPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", htmlPath, err)
	}

	fmt.Printf("profilegen: wrote %s (%s)\n", htmlPath, version)

	return nil
}

// templateFuncs is the func map template/index.html.tmpl is parsed with.
// Shared with the tests so they render against the same helpers as Build.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"sub":  func(a, b int) int { return a - b },
		"html": func(s string) template.HTML { return template.HTML(s) }, //nolint:gosec // trusted, self-authored content
	}
}

// PageData is the data handed to template/index.html.tmpl.
type PageData struct {
	*Resume
	Version string
	MDName  string
}
