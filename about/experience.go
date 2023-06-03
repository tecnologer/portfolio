package about

// Experience represents a work experience.
type Experience struct {
	Company      string   `json:"company"`      // The name of the company.
	Title        string   `json:"title"`        // The job title.
	StartDate    string   `json:"start_date"`   // The start date of the contract.
	EndDate      string   `json:"end_date"`     // The end date of the contract.
	Location     string   `json:"location"`     // The location of the company.
	Roles        []string `json:"role"`         // The list of roles perfomed
	Technologies []string `json:"technologies"` // The technologies used in the experience.
}

// ListExperience returns a list of experiences.
func ListExperience() []*Experience {
	return []*Experience{
		{
			Company:   "Pentalog",
			Title:     "Senior Software Engineer",
			StartDate: "December 2019",
			EndDate:   "current",
			Location:  "Guadalajara, Mexico",
			Roles: []string{
				"Back-End",
				"Software Architect",
				"Speaker",
				"People Manager",
			},
			Technologies: []string{
				"C# dotnet",
				"MSSQL",
				"Go",
				"PostgreSQL",
				"GraphQL",
				"AWS",
				"Git",
				"Github Actions",
			},
		},
		{
			Company:   "Ubilogix",
			Title:     "Software Engineer",
			StartDate: "March 2017",
			EndDate:   "December 2019",
			Location:  "Ensenada, Mexico",
			Roles: []string{
				"Full-stack",
				"Software Architect",
			},
			Technologies: []string{
				"C# .NET",
				"WPF",
				"Go",
				"SQLite",
				"Zigbee",
				"Git/GitLab",
			},
		},
		{
			Company:   "Alesayi Development Company (ADCO)",
			Title:     "System Architect",
			StartDate: "August 2015",
			EndDate:   "January 2017",
			Location:  "Jeddah, Saudi Arabia",
			Roles: []string{
				"Full-stack",
				"Software Architect",
			},
			Technologies: []string{
				"C# .NET MVC 5",
				"HTML 5",
				"CSS",
				"Javascript",
				"MSSQL",
				"TFS",
				"Git/Github",
			},
		},
		{
			Company:   "RedRabbit MX",
			Title:     "Full Stack web Developer",
			StartDate: "June 2012",
			EndDate:   "June 2015",
			Location:  "Culiacan, Mexico",
			Roles: []string{
				"Full-stack",
			},
			Technologies: []string{
				"Coldfusion",
				"Python",
				"HTML 5",
				"CSS",
				"Javascript",
				"AngularJS",
				"JQuery",
				"MSSQL",
			},
		},
	}
}
