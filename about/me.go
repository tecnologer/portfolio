// The "about" package is a collection of definitions and structures that aim to expose my professional experience and the technologies I have worked with. It provides a convenient way to present my background and skills to others who are interested in learning more about my expertise.

// Within the "about" package, you will find various structures that represent different aspects of my experience. These structures may include details such as my work history, projects I have been involved in, technologies I have utilized, and any other relevant information that showcases my capabilities.

// By utilizing the "about" package, it becomes easier to present a comprehensive overview of my professional journey, making it simpler for potential collaborators or employers to understand my skills and areas of expertise.

// Overall, the "about" package serves as a concise and organized representation of my professional background, providing a convenient way to share my experience and technologies with others in a clear and structured manner.
package about

// Me represents personal information and experience.
type Me struct {
	Name       string        `json:"name"`
	Bio        string        `json:"bio"`
	Experience []*Experience `json:"experience"`
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
		Experience: ListExperience(),
	}
}
