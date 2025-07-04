package data

// Project represents a single project entry
type Project struct {
	Title       string `json:"title"`
	Image       string `json:"image"`
	URL         string `json:"url"`
	Github      string `json:"github,omitempty"`
	Website     string `json:"website,omitempty"`
	Description string `json:"description"`
}

type Projects []Project

// GetProjects loads the projects from the JSON manifest in the order they appear
func GetProjects() (Projects, error) {
	projects := make(Projects, 0)
	err := ReadJsonFile("data/projectManifest.json", &projects)
	if err != nil {
		return projects, err
	}
	return projects, nil
}
