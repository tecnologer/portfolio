//nolint:testpackage // white-box: this file exercises the unexported test path helpers, which are unexported.
package portfolio

// Repo-root inputs, relative to this package directory. go test runs with
// CWD = package dir; main.go's flag defaults resolve from the caller's CWD
// (the repo root). Hence the prefix here and not there.
const (
	repoRoot = "../../"
	srcMD    = repoRoot + "page/reydavid_experience.md"
	tmplPath = repoRoot + "template/index.html.tmpl"
	dataYAML = repoRoot + "data/resume.yaml"
)
