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

var cachedInterviews Interviews
var interviewsLoaded bool

// GetInterviews loads the interviews from the JSON manifest in the order they appear, but only once (cached in memory)
func GetInterviews() (Interviews, error) {
	if interviewsLoaded {
		return cachedInterviews, nil
	}
	interviews := make(Interviews, 0)
	err := ReadJsonFile("data/interviewManifest.json", &interviews)
	if err != nil {
		return interviews, err
	}
	cachedInterviews = interviews
	interviewsLoaded = true
	return cachedInterviews, nil
}

// ClearInterviewsCache clears the cached interviews (for testing or reload)
func ClearInterviewsCache() {
	interviewsLoaded = false
	cachedInterviews = nil
}
