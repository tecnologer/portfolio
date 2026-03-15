// Package about provides structures and data representing personal and
// professional information for the portfolio.
package about

// Me represents the portfolio owner's personal and professional profile.
type Me struct {
	Name           string           `json:"name"`            // Full display name.
	Bio            string           `json:"bio"`             // Short professional summary.
	Experience     []*Experience    `json:"experience"`      // Chronological work history.
	Education      []*Education     `json:"education"`       // Academic background.
	Certifications []*Certification `json:"certifications"`  // Professional certifications.
	ContactOptions []*Contact       `json:"contact_options"` // Available contact channels.
}

// NewMe returns a new Me populated with predefined personal and professional information.
func NewMe() *Me {
	return &Me{
		Name: "Rey David Dominguez Soto",
		Bio: `I am Rey David, a Senior Software Engineer with over 13 years of experience,
			specializing in Go backend engineering, security platforms, DevSecOps automation,
			and cloud integrations. I am passionate about solving complex problems and
			building innovative solutions.`,
		Experience:     ListExperience(),
		Education:      ListEducation(),
		Certifications: ListCertifications(),
		ContactOptions: ListContactOptions(),
	}
}
