package about

// Project represents a personal side project built outside of employment.
type Project struct {
	Name         string   `json:"name"`         // Project name.
	Tagline      string   `json:"tagline"`      // Short description of what the project does.
	Status       string   `json:"status"`       // Current state (e.g. Live, In construction).
	URL          string   `json:"url"`          // Public URL where the project can be seen, if any.
	RepoURL      string   `json:"repo_url"`     // Source code repository URL, if any.
	Technologies []string `json:"technologies"` // Tools, languages, and platforms used.
}

// ListProjects returns the personal side projects.
func ListProjects() []*Project {
	return []*Project{
		{
			Name:    "JobTracker",
			Tagline: "Open-source, local-first job application tracker. Go REST API + Vue 3 SPA.",
			Status:  "Live",
			URL:     "https://jobtracker.tecnologer.net",
			RepoURL: "https://github.com/Tecnologer/jobtracker",
			Technologies: []string{
				"Go", "Vue 3", "SQLite", "REST API", "Wails", "Docker",
			},
		},
		{
			Name:    "Cotiza3D",
			Tagline: "3D-printing quote and order manager. Single Go binary serving a REST API and an embedded Vue SPA, PostgreSQL-backed.",
			Status:  "In construction (landing page live)",
			URL:     "https://cotiza3d.com",
			Technologies: []string{
				"Go", "Vue 3", "PostgreSQL", "Docker",
			},
		},
	}
}
