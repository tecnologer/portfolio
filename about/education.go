package about

// Education represents an academic background entry.
type Education struct {
	Institution string `json:"institution"` // Name of the educational institution.
	Degree      string `json:"degree"`      // Degree or program completed.
	Field       string `json:"field"`       // Field of study.
	StartYear   string `json:"start_year"`  // Year the program began.
	EndYear     string `json:"end_year"`    // Year the program ended.
}

// ListEducation returns the academic background in reverse chronological order.
func ListEducation() []*Education {
	return []*Education{
		{
			Institution: "Instituto Tecnologico de Culiacan",
			Degree:      "Engineer's degree",
			Field:       "Computer Software Engineering",
			StartYear:   "2009",
			EndYear:     "2014",
		},
		{
			Institution: "Universidad de Guadalajara",
			Degree:      "Coursework",
			Field:       "Computer Science",
			StartYear:   "2007",
			EndYear:     "2008",
		},
	}
}
