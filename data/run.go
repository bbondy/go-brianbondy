package data

import "sort"

type Run struct {
	Date        string   `json:"date"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	StravaURLs  []string `json:"strava_urls,omitempty"`
	ImagePath   string   `json:"image_path,omitempty"`
	ImagePaths  []string `json:"image_paths,omitempty"`
	BlogPostID  int      `json:"blog_post_id,omitempty"`
	Time        string   `json:"time,omitempty"`
	Distance    string   `json:"distance,omitempty"`
	Elevation   string   `json:"elevation,omitempty"`
}

type Runs []Run

func (r Runs) Len() int      { return len(r) }
func (r Runs) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r Runs) Less(i, j int) bool {
	return r[i].Date > r[j].Date
}

var cachedRuns Runs
var runsLoaded bool

// GetRuns loads the runs from the JSON manifest in the order they appear, but only once (cached in memory)
func GetRuns() (Runs, error) {
	if runsLoaded {
		return cachedRuns, nil
	}
	runs := make(Runs, 0)
	err := ReadJsonFile("data/runManifest.json", &runs)
	if err != nil {
		return runs, err
	}
	sort.Sort(runs)
	cachedRuns = runs
	runsLoaded = true
	return cachedRuns, nil
}

// ClearRunsCache clears the cached runs (for testing or reload)
func ClearRunsCache() {
	runsLoaded = false
	cachedRuns = nil
}
