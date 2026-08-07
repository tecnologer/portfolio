//nolint:testpackage // white-box: this file exercises the unexported fixtures, which is unexported.
package portfolio

// Fixture literals shared across the test files. Company names must match
// the real page/reydavid_experience.md entries, since golden_test.go
// asserts against the extracted resume.
const (
	companyDefectDojo = "DefectDojo Inc"
	companyPullerTech = "Puller Tech"
	companyAlesayi    = "Alesayi Development Co. (ADCO)"
)

// Synthetic source names used by the tag and story fixtures.
const (
	sourceAcme      = "acme"
	sourceAcmeTitle = "Acme"
	sourceAlpha     = "alpha"
	sourceOne       = "source1"
	locationRemote  = "Remote"
)

// schoolITC is the education-entry institution used across the parser
// fixtures; it mirrors page/reydavid_experience.md verbatim.
//
//nolint:misspell // "Instituto" is Spanish, not a misspelling of "institution".
const schoolITC = "Instituto Tecnológico de Culiacán"
