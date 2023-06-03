// The "about" package is a collection of definitions and structures that aim to expose my professional experience and the technologies I have worked with. It provides a convenient way to present my background and skills to others who are interested in learning more about my expertise.
package about

// Me represents personal information and experience.
type Me struct {
	Name           string        `json:"name"`            // The name of the person.
	Bio            string        `json:"bio"`             // A brief biography.
	Experience     []*Experience `json:"experience"`      // List of experiences.
	ContactOptions []*Contact    `json:"contact_options"` // List of contact platform
}

// NewMe creates a new instance of the Me struct with predefined values.
// It initializes the Name, Bio, and Experience fields with default values or data obtained from external sources.
// Returns a pointer to the newly created Me struct.
func NewMe() *Me {
	return &Me{
		Name: "Rey David",
		Bio: `
			I am a Senior Software Engineer with experience in developing scalable and distributed systems. 
			I am passionate about leveraging technology to solve complex problems and enjoy collaborating with teams to deliver innovative solutions.
		`,
		Experience:     ListExperience(),
		ContactOptions: ListContactOptions(),
	}
}
