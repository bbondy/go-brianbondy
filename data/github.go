package data

// GitHubStats represents GitHub statistics for a repository
type GitHubStats struct {
	CommitCount int `json:"commitCount,omitempty"`
	PrCount     int `json:"prCount,omitempty"`
	Languages   []Language
}

// Language represents a programming language with its usage percentage and color
type Language struct {
	Name    string
	Percent float64
	Color   string
}
