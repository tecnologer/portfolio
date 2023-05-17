package about

// Contact represents a contact platform and address.
type Contact struct {
	Platform string `json:"platform"` // The contact platform (e.g., email, social media).
	Address  string `json:"address"`  // The contact address (e.g., email address, social media handle).
}

// ListContactOptions returns a list of contact platforms.
func ListContactOptions() []*Contact {
	return []*Contact{
		{
			Platform: "Email",
			Address:  "rdominguez@tecnologer.net",
		},
		{
			Platform: "LinkedIn",
			Address:  "https://www.linkedin.com/in/rey-david-dominguez-soto",
		},
		{
			Platform: "GitHub",
			Address:  "https://github.com/tecnologer",
		},
		{
			Platform: "Telegram",
			Address:  "https://t.me/tecnologer",
		},
	}
}
