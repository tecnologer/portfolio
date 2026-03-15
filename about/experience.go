package about

// Experience represents a single work experience entry.
type Experience struct {
	Company      string   `json:"company"`      // Employer name.
	Title        string   `json:"title"`        // Job title held.
	StartDate    string   `json:"start_date"`   // Month and year the role began.
	EndDate      string   `json:"end_date"`     // Month and year the role ended, or "Present".
	Location     string   `json:"location"`     // City, country, or "Remote".
	Roles        []string `json:"roles"`        // Responsibilities performed (e.g. Back-End, Architect).
	Technologies []string `json:"technologies"` // Tools, languages, and platforms used.
}

// ListExperience returns the work history in reverse chronological order.
func ListExperience() []*Experience {
	return []*Experience{
		{
			Company:   "DefectDojo",
			Title:     "Senior Software Engineer, Security Tooling",
			StartDate: "April 2024",
			EndDate:   "Present",
			Location:  "Remote, Mexico",
			Roles: []string{
				"Back-End",
				"Security Tooling",
				"Open Source",
			},
			Technologies: []string{
				"Go",
				"PostgreSQL",
				"Docker",
				"Docker Compose",
				"Git",
				"GitHub Actions",
				"Linux",
				"Bash",
				"REST API",
				"Checkmarx One",
				"Wiz",
				"Snyk",
				"SonarCloud",
				"SonarQube",
				"Dependency-Track",
				"IriusRisk",
				"Akamai",
				"Azure DevOps",
				"GitHub Boards",
				"GitLab Boards",
				"ServiceNow",
				"Slack",
				"Telegram",
				"PagerDuty",
			},
		},
		{
			Company:   "Pentalog",
			Title:     "Senior Software Engineer",
			StartDate: "December 2019",
			EndDate:   "February 2024",
			Location:  "Zapopan, Mexico",
			Roles: []string{
				"Back-End",
				"Software Architect",
				"Speaker",
				"People Manager",
			},
			Technologies: []string{
				"Go",
				"Python",
				"C# .NET",
				"Dotnet Core",
				"Entity Framework",
				"MSSQL",
				"PostgreSQL",
				"Aurora",
				"GraphQL",
				"AWS Lambda",
				"AWS Cognito",
				"AWS S3",
				"AWS CloudFormation",
				"AWS SAM",
				"AWS AppSync",
				"AWS Secrets Manager",
				"GitHub Actions",
				"LangChain",
				"OpenAI API",
				"Casbin",
				"Docker",
				"Linux",
			},
		},
		{
			Company:   "Ubilogix",
			Title:     "Software Engineer",
			StartDate: "March 2017",
			EndDate:   "December 2019",
			Location:  "Ensenada, Baja California, Mexico",
			Roles: []string{
				"Full-Stack",
				"Software Architect",
			},
			Technologies: []string{
				"Go",
				"C# .NET",
				"WPF",
				"SQLite",
				"Zigbee",
				"Git",
				"GitLab",
				"Windows",
				"Ubuntu",
			},
		},
		{
			Company:   "Alesayi Development Company (ADCO)",
			Title:     "System Architect",
			StartDate: "August 2015",
			EndDate:   "January 2017",
			Location:  "Jeddah, Saudi Arabia",
			Roles: []string{
				"Full-Stack",
				"Software Architect",
			},
			Technologies: []string{
				"C# .NET MVC 5",
				"HTML 5",
				"CSS",
				"JavaScript",
				"MSSQL",
				"JQuery",
				"AngularJS",
				"TFS",
				"Git",
				"Windows",
			},
		},
		{
			Company:   "RedRabbit MX",
			Title:     "Senior Full Stack web Developer",
			StartDate: "June 2012",
			EndDate:   "June 2015",
			Location:  "Culiacan, Sinaloa, Mexico",
			Roles: []string{
				"Full-Stack",
			},
			Technologies: []string{
				"Coldfusion",
				"Python",
				"HTML 5",
				"CSS",
				"JavaScript",
				"JQuery",
				"AngularJS",
				"MSSQL",
				"MongoDB",
				"Git",
				"Windows",
			},
		},
	}
}
