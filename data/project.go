package data

// Project represents a single project entry
type Project struct {
	Title          string       `json:"title"`
	Image          string       `json:"image"`
	URL            string       `json:"url"`
	Github         string       `json:"github,omitempty"`
	Website        string       `json:"website,omitempty"`
	BlogPostID     int          `json:"blogPostID,omitempty"`
	Description    string       `json:"description"`
	GitHubStats    *GitHubStats `json:"githubStats,omitempty"`    // GitHub statistics from Python script
	SearchKeywords []string     `json:"searchKeywords,omitempty"` // Keywords for GitHub search to avoid double counting
}

type Projects []Project

var cachedProjects Projects
var projectsLoaded bool

// GetProjects loads the projects from the JSON manifest in the order they appear, but only once (cached in memory)
func GetProjects() (Projects, error) {
	if projectsLoaded {
		return cachedProjects, nil
	}
	projects := make(Projects, 0)
	err := ReadJsonFile("data/projectManifest.json", &projects)
	if err != nil {
		return projects, err
	}
	cachedProjects = projects
	projectsLoaded = true
	return cachedProjects, nil
}

// ClearProjectsCache clears the cached projects (for testing or reload)
func ClearProjectsCache() {
	projectsLoaded = false
	cachedProjects = nil
}
