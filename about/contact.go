package about

// Contact represents a channel through which the portfolio owner can be reached.
type Contact struct {
	Platform string `json:"platform"` // Service or medium (e.g. Email, LinkedIn, GitHub).
	Address  string `json:"address"`  // URL or address on that platform.
}

// ListContactOptions returns the available contact channels.
func ListContactOptions() []*Contact {
	return []*Contact{
		{
			Platform: "Email",
			Address:  "rdominguez@tecnologer.net",
		},
		{
			Platform: "LinkedIn",
			Address:  "https://www.linkedin.com/in/tecnologer",
		},
		{
			Platform: "GitHub",
			Address:  "https://github.com/Tecnologer",
		},
		{
			Platform: "GitLab",
			Address:  "https://gitlab.com/tecnologer",
		},
		{
			Platform: "Telegram",
			Address:  "https://t.me/tecnologer",
		},
		{
			Platform: "Personal Website",
			Address:  "https://www.tecnologer.net",
		},
	}
}
