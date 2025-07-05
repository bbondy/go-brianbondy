package data

// Interview represents a single interview entry
type Interview struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	URL         string `json:"url"`
	Platform    string `json:"platform"`
	Channel     string `json:"channel"`
	Description string `json:"description"`
}

type Interviews []Interview

// GetInterviews loads the interviews from the JSON manifest in the order they appear
func GetInterviews() (Interviews, error) {
	interviews := make(Interviews, 0)
	err := ReadJsonFile("data/interviewManifest.json", &interviews)
	if err != nil {
		return interviews, err
	}
	return interviews, nil
}
