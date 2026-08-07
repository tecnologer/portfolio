package portfolio

// Literals shared by the tag dictionaries in tags.go, the section keyword
// table in parse.go, and the date normalizers in extract.go/model.go. Kept
// here so one spelling is the single source of truth for each.

// datePresent is the normalized open-ended endpoint token (H-3 / D-31).
const datePresent = "Present"

// Theme and role tag slugs (§5.2, §5.3). These double as section slugs in
// parse.go's keyword table, which is why they live here rather than beside
// one dictionary.
const (
	tagPerformance   = "performance"
	tagMigration     = "migration"
	tagHomelab       = "homelab"
	tagOpenSource    = "open source"
	tagCommunity     = "community"
	tagSpeaker       = "speaker"
	tagArchitect     = "architect"
	tagFullStack     = "full-stack"
	tagBackEnd       = "back-end"
	tagPeopleManager = "people manager"
	tagDotnet        = "dotnet"
	tagPostgreSQL    = "postgresql"
)

// Technology display casing (techRule.Display), used verbatim in
// Experience.Technologies.
const (
	displayPostgreSQL = "PostgreSQL"
	displayDocker     = "Docker"
)
