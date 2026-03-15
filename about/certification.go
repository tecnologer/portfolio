package about

// Certification represents a professional certification.
type Certification struct {
	Name string `json:"name"` // Certification title.
}

// ListCertifications returns the list of professional certifications.
func ListCertifications() []*Certification {
	return []*Certification{
		{Name: "Generative AI for Software Developers Specialization"},
		{Name: "Generative AI: Prompt Engineering Basics"},
		{Name: "Generative AI: Introduction and Applications"},
		{Name: "Foundations of Cybersecurity"},
	}
}
