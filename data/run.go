package data

import (
	"math"
	"sort"
	"strconv"
	"sync"
	"time"
)

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
	err := ReadJsonFile("data/memorableRuns.json", &runs)
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

// StravaRun represents a run from the Strava API manifest
type StravaRun struct {
	Date       string  `json:"date"`
	Title      string  `json:"title"`
	Type       string  `json:"type,omitempty"`
	DistanceKm float64 `json:"distance_km"`
	Time       string  `json:"time"`
	Pace       string  `json:"pace"`
	ActivityID string  `json:"activity_id,omitempty"`
	Elevation  string  `json:"elevation,omitempty"`
}

type StravaRuns []StravaRun

func (r StravaRuns) Len() int      { return len(r) }
func (r StravaRuns) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r StravaRuns) Less(i, j int) bool {
	return r[i].Date > r[j].Date
}

// ContributionGraph represents a GitHub-style contribution graph
type ContributionGraph struct {
	Days     []ContributionDay
	MaxCount int
	Years    []int
}

type ContributionDay struct {
	Date          string
	Count         int
	Level         int         // 0-4 for color intensity
	Runs          []StravaRun // Actual runs for this date
	DistanceRatio float64     // 0.0 to 1.0, for gradient coloring
	Lightness     float64     // 15 + 45*ratio, rounded to 1 decimal
}

var cachedStravaRuns StravaRuns
var stravaRunsLoaded bool

// GetStravaRuns loads the Strava runs from the JSON manifest, but only once (cached in memory)
func GetStravaRuns() (StravaRuns, error) {
	if stravaRunsLoaded {
		return cachedStravaRuns, nil
	}
	runs := make(StravaRuns, 0)
	err := ReadJsonFile("data/stravaRunManifest.json", &runs)
	if err != nil {
		return runs, err
	}
	for i := range runs {
		runs[i].Type = normalizeStravaActivityType(runs[i])
	}
	sort.Sort(runs)
	cachedStravaRuns = runs
	stravaRunsLoaded = true
	return cachedStravaRuns, nil
}

// normalizeStravaActivityType groups equivalent Strava categories for display and filtering.
func normalizeStravaActivityType(run StravaRun) string {
	if run.Type == "Workout" {
		return "StairStepper"
	}
	return run.Type
}

// ClearStravaRunsCache clears the cached Strava runs (for testing or reload)
func ClearStravaRunsCache() {
	stravaRunsLoaded = false
	cachedStravaRuns = nil
	contributionGraphCache.Lock()
	contributionGraphCache.graphs = make(map[string]*ContributionGraph2D)
	contributionGraphCache.Unlock()
}

// isActiveRun returns true if the entry represents a real activity (non-zero distance or time)
func isActiveRun(r StravaRun) bool {
	if r.DistanceKm > 0 {
		return true
	}
	if parseTimeStringToMinutes(r.Time) > 0 {
		return true
	}
	return false
}

// GenerateContributionGraph creates a GitHub-style contribution graph from Strava runs
func GenerateContributionGraph(yearFilter string) (*ContributionGraph, error) {
	runs, err := GetStravaRuns()
	if err != nil {
		return nil, err
	}

	// Filter runs by year if specified
	var filteredRuns StravaRuns
	if yearFilter != "" && yearFilter != "365" {
		year, err := strconv.Atoi(yearFilter)
		if err == nil {
			for _, run := range runs {
				if len(run.Date) >= 4 {
					if runYear, err := strconv.Atoi(run.Date[:4]); err == nil && runYear == year {
						filteredRuns = append(filteredRuns, run)
					}
				}
			}
		}
	} else {
		// Use all runs for "Last 365 Days" view
		filteredRuns = runs
	}

	// Create a map of date to runs
	dateRuns := make(map[string][]StravaRun)
	for _, run := range filteredRuns {
		dateRuns[run.Date] = append(dateRuns[run.Date], run)
	}

	// Get all available years
	years := getAvailableYears(runs)

	// Generate the last 365 days
	days := make([]ContributionDay, 0, 365)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -364) // 365 days total

	maxCount := 0
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		runsForDate := dateRuns[dateStr]
		count := len(runsForDate)
		if count > maxCount {
			maxCount = count
		}
		days = append(days, ContributionDay{
			Date:  dateStr,
			Count: count,
			Runs:  runsForDate,
		})
	}

	// Calculate levels (0-4) based on count relative to max
	for i := range days {
		if maxCount == 0 {
			days[i].Level = 0
		} else {
			level := int(float64(days[i].Count) / float64(maxCount) * 4)
			if level > 4 {
				level = 4
			}
			days[i].Level = level
		}
	}

	return &ContributionGraph{
		Days:     days,
		MaxCount: maxCount,
		Years:    years,
	}, nil
}

// getAvailableYears returns all years that have runs, sorted descending
func getAvailableYears(runs StravaRuns) []int {
	yearMap := make(map[int]bool)
	for _, run := range runs {
		if len(run.Date) >= 4 {
			if year, err := strconv.Atoi(run.Date[:4]); err == nil {
				yearMap[year] = true
			}
		}
	}

	years := make([]int, 0, len(yearMap))
	for year := range yearMap {
		years = append(years, year)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	return years
}

// GetRunsForDate returns all runs for a specific date
func GetRunsForDate(date string) ([]StravaRun, error) {
	runs, err := GetStravaRuns()
	if err != nil {
		return nil, err
	}

	var runsForDate []StravaRun
	for _, run := range runs {
		if run.Date == date {
			runsForDate = append(runsForDate, run)
		}
	}

	return runsForDate, nil
}

// GetStravaRunTotals calculates totals from Strava runs
func GetStravaRunTotals() (int, float64, int, int, error) {
	runs, err := GetStravaRuns()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	totalRuns := len(runs)
	totalDistanceKm := 0.0
	totalTimeMinutes := 0
	totalElevationM := 0
	for _, run := range runs {
		totalDistanceKm += run.DistanceKm
		totalTimeMinutes += parseTimeStringToMinutes(run.Time)
		totalElevationM += parseElevationStringToMeters(run.Elevation)
	}
	return totalRuns, totalDistanceKm, totalElevationM, totalTimeMinutes, nil
}

// GetStravaRunTotalsFor calculates totals for a given view ("all", "365", or specific year)
func GetStravaRunTotalsFor(yearFilter string) (int, float64, int, int, error) {
	runs, err := GetStravaRuns()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	var filtered StravaRuns
	switch yearFilter {
	case "all", "":
		filtered = runs
	case "365":
		now := time.Now()
		endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		startDate := endDate.AddDate(0, 0, -364)
		for _, run := range runs {
			if t, err := time.Parse("2006-01-02", run.Date); err == nil {
				if (t.Equal(startDate) || t.After(startDate)) && (t.Equal(endDate) || t.Before(endDate)) {
					filtered = append(filtered, run)
				}
			}
		}
	default:
		if year, err := strconv.Atoi(yearFilter); err == nil {
			for _, run := range runs {
				if len(run.Date) >= 4 {
					if runYear, err := strconv.Atoi(run.Date[:4]); err == nil && runYear == year {
						filtered = append(filtered, run)
					}
				}
			}
		}
	}

	totalRuns := len(filtered)
	totalDistanceKm := 0.0
	totalTimeMinutes := 0
	totalElevationM := 0
	for _, run := range filtered {
		totalDistanceKm += run.DistanceKm
		totalTimeMinutes += parseTimeStringToMinutes(run.Time)
		totalElevationM += parseElevationStringToMeters(run.Elevation)
	}
	return totalRuns, totalDistanceKm, totalElevationM, totalTimeMinutes, nil
}

// GetStravaRunTotalsForTypes returns view totals restricted to the selected activity types.
func GetStravaRunTotalsForTypes(yearFilter string, types map[string]bool) (int, float64, int, int, error) {
	runs, err := GetStravaRuns()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	var totalRuns int
	var totalDistanceKm float64
	var totalTimeMinutes, totalElevationM int
	now := time.Now()
	startDate := now.AddDate(0, 0, -364)
	for _, run := range runs {
		inView := yearFilter == "" || yearFilter == "all"
		if yearFilter == "365" {
			t, parseErr := time.Parse("2006-01-02", run.Date)
			inView = parseErr == nil && !t.Before(startDate) && !t.After(now)
		} else if year, parseErr := strconv.Atoi(yearFilter); parseErr == nil {
			inView = len(run.Date) >= 4 && run.Date[:4] == strconv.Itoa(year)
		}
		typ := run.Type
		if typ == "" {
			typ = "Unknown"
		}
		if !inView || !types[typ] {
			continue
		}
		totalRuns++
		totalDistanceKm += run.DistanceKm
		totalTimeMinutes += parseTimeStringToMinutes(run.Time)
		totalElevationM += parseElevationStringToMeters(run.Elevation)
	}
	return totalRuns, totalDistanceKm, totalElevationM, totalTimeMinutes, nil
}

// GetLastUpdatedDate returns the most recent date from Strava runs
func GetLastUpdatedDate() (string, error) {
	runs, err := GetStravaRuns()
	if err != nil {
		return "", err
	}
	if len(runs) == 0 {
		return "", nil
	}
	// runs are already sorted by date descending, so first one is most recent
	mostRecentDate := runs[0].Date

	// Parse and format the date nicely
	if t, err := time.Parse("2006-01-02", mostRecentDate); err == nil {
		return t.Format("January 2, 2006"), nil
	}
	return mostRecentDate, nil
}

// Helper to get days, hours, minutes from total minutes
func SplitMinutesToDaysHoursMinutes(totalMinutes int) (days, hours, minutes int) {
	days = totalMinutes / (24 * 60)
	hours = (totalMinutes % (24 * 60)) / 60
	minutes = totalMinutes % 60
	return
}

// parseTimeStringToMinutes converts "Xh Ym" or "Xm" to total minutes
func parseTimeStringToMinutes(timeStr string) int {
	if len(timeStr) == 0 {
		return 0
	}

	totalMinutes := 0

	// Use regex or string parsing to handle multi-digit numbers
	i := 0
	for i < len(timeStr) {
		// Parse number
		numStr := ""
		for i < len(timeStr) && timeStr[i] >= '0' && timeStr[i] <= '9' {
			numStr += string(timeStr[i])
			i++
		}

		if numStr == "" {
			i++
			continue
		}

		num := 0
		for _, digit := range numStr {
			num = num*10 + int(digit-'0')
		}

		// Check what unit follows
		if i < len(timeStr) {
			switch timeStr[i] {
			case 'h':
				totalMinutes += num * 60
			case 'm':
				totalMinutes += num
			}
			i++
		}
	}

	return totalMinutes
}

// parseTimeStringToHours converts "Xh Ym" or "Xm" to total hours as float64
func parseTimeStringToHours(timeStr string) float64 {
	totalMinutes := parseTimeStringToMinutes(timeStr)
	return float64(totalMinutes) / 60.0
}

// parseElevationStringToMeters converts "Xm" or "X.Xm" to meters
func parseElevationStringToMeters(elevationStr string) int {
	if len(elevationStr) == 0 {
		return 0
	}

	elevationM := 0

	// Use regex or string parsing to handle multi-digit numbers
	i := 0
	for i < len(elevationStr) {
		// Parse number
		numStr := ""
		for i < len(elevationStr) && elevationStr[i] >= '0' && elevationStr[i] <= '9' {
			numStr += string(elevationStr[i])
			i++
		}

		if numStr == "" {
			i++
			continue
		}

		num := 0
		for _, digit := range numStr {
			num = num*10 + int(digit-'0')
		}

		// Check what unit follows
		if i < len(elevationStr) {
			if elevationStr[i] == 'm' {
				elevationM = num
			}
			i++
		}
	}

	return elevationM
}

// ContributionGraph2D represents the grid for the contribution graph
// Weeks are columns, days are rows (0=Sunday, 1=Monday, ... 6=Saturday)
type ContributionGraph2D struct {
	Grid           [][]ContributionDay // [week][day]
	MonthLabels    []string            // 3-letter month for each week
	MonthLabelShow []bool              // true if label should be shown for this week
	UniqueMonths   []string            // unique months for evenly spaced labels
	DayLabels      []string            // 3-letter day for each row
	Weeks          int
	Days           int
	Years          []int // Available years for dropdown
}

// ContributionGraphCanvas is the small client payload used to draw the heatmap.
// It intentionally omits activity details; those are fetched only after a day is clicked.
type ContributionGraphCanvas struct {
	StartDate  string                           `json:"startDate"`
	Weeks      int                              `json:"weeks"`
	ActiveDays map[string]ContributionCanvasDay `json:"activeDays"`
}

type ContributionCanvasDay struct {
	Count     int      `json:"count"`
	Level     int      `json:"level"`
	Lightness float64  `json:"lightness"`
	Types     []string `json:"types"`
}

var contributionGraphCache = struct {
	sync.RWMutex
	graphs map[string]*ContributionGraph2D
}{graphs: make(map[string]*ContributionGraph2D)}

// GenerateContributionGraph2D returns a 2D grid for the contribution graph, with month and day labels
func GenerateContributionGraph2D(yearFilter string) (*ContributionGraph2D, error) {
	// The Strava manifest is static for the lifetime of this process.
	cacheKey := yearFilter
	contributionGraphCache.RLock()
	graph := contributionGraphCache.graphs[cacheKey]
	contributionGraphCache.RUnlock()
	if graph != nil {
		return graph, nil
	}

	graph, err := generateContributionGraph2D(yearFilter)
	if err != nil {
		return nil, err
	}
	contributionGraphCache.Lock()
	contributionGraphCache.graphs[cacheKey] = graph
	contributionGraphCache.Unlock()
	return graph, nil
}

func generateContributionGraph2D(yearFilter string) (*ContributionGraph2D, error) {
	runs, err := GetStravaRuns()
	if err != nil {
		return nil, err
	}
	var filteredRuns StravaRuns
	var startDate, endDate time.Time
	if yearFilter == "all" {
		// All time: include every activity; compute bounds from data
		filteredRuns = runs
		var minDate, maxDate *time.Time
		for _, run := range runs {
			if t, err := time.Parse("2006-01-02", run.Date); err == nil {
				if minDate == nil || t.Before(*minDate) {
					tt := t
					minDate = &tt
				}
				if maxDate == nil || t.After(*maxDate) {
					tt := t
					maxDate = &tt
				}
			}
		}
		if minDate == nil || maxDate == nil {
			endDate = time.Now()
			startDate = endDate.AddDate(0, 0, -364)
		} else {
			startDate = time.Date(minDate.Year(), minDate.Month(), minDate.Day(), 0, 0, 0, 0, time.UTC)
			endDate = time.Date(maxDate.Year(), maxDate.Month(), maxDate.Day(), 0, 0, 0, 0, time.UTC)
		}
	} else if yearFilter != "" && yearFilter != "365" {
		year, err := strconv.Atoi(yearFilter)
		if err == nil {
			startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			endDate = time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
			// Backtrack to previous Monday if Jan 1 is not Monday
			deltaToMonday := (int(startDate.Weekday()) - int(time.Monday) + 7) % 7
			if deltaToMonday != 0 {
				startDate = startDate.AddDate(0, 0, -deltaToMonday)
			}
			for _, run := range runs {
				if len(run.Date) >= 4 {
					if runYear, err := strconv.Atoi(run.Date[:4]); err == nil && runYear == year {
						filteredRuns = append(filteredRuns, run)
					}
				}
			}
		}
	} else {
		filteredRuns = runs
		endDate = time.Now()
		// Start with exact last 365 days
		desiredStart := endDate.AddDate(0, 0, -364)
		// Backtrack to previous Monday to align weeks
		deltaToMonday := (int(desiredStart.Weekday()) - int(time.Monday) + 7) % 7
		startDate = desiredStart.AddDate(0, 0, -deltaToMonday)
	}

	// For non-"all" views, align start date to Monday of that week
	if yearFilter != "all" {
		deltaToMonday := (int(startDate.Weekday()) - int(time.Monday) + 7) % 7
		if deltaToMonday != 0 {
			startDate = startDate.AddDate(0, 0, -deltaToMonday)
		}
	}

	// Map date to runs (ignore zero-distance AND zero-time entries)
	dateRuns := make(map[string][]StravaRun)
	for _, run := range filteredRuns {
		if !isActiveRun(run) {
			continue
		}
		dateRuns[run.Date] = append(dateRuns[run.Date], run)
	}

	// For all-time view, optionally start at the first Monday that has activity
	if yearFilter == "all" {
		// Find earliest activity date from filteredRuns and align to Monday before that activity
		var minActive time.Time
		minSet := false
		for _, run := range filteredRuns {
			if !isActiveRun(run) {
				continue
			}
			if t, err := time.Parse("2006-01-02", run.Date); err == nil {
				if !minSet || t.Before(minActive) {
					minActive = t
					minSet = true
				}
			}
		}
		if minSet {
			weekdayA := int(minActive.Weekday())
			// If already Monday, keep; else backtrack to previous Monday
			if weekdayA == int(time.Monday) {
				startDate = time.Date(minActive.Year(), minActive.Month(), minActive.Day(), 0, 0, 0, 0, time.UTC)
			} else {
				// Compute days to subtract to reach Monday
				delta := (7 + weekdayA - int(time.Monday)) % 7
				if delta == 0 {
					delta = 7
				}
				startDate = time.Date(minActive.Year(), minActive.Month(), minActive.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -delta)
			}
		}
	}

	rows := 7
	// Determine columns based on actual date span for all views
	daysSpan := int(endDate.Sub(startDate).Hours()/24) + 1
	cols := 1
	if daysSpan > 7 {
		cols = int(math.Ceil(float64(daysSpan) / 7.0))
	}
	grid := make([][]ContributionDay, rows)
	for r := 0; r < rows; r++ {
		grid[r] = make([]ContributionDay, cols)
	}
	monthLabels := make([]string, cols)
	monthLabelShow := make([]bool, cols)
	var prevLabel string
	var prevYear int
	maxDistance := 0.0
	for c := 0; c < cols; c++ {
		for r := 0; r < rows; r++ {
			// Monday-first row order; startDate is aligned to Monday
			curDate := startDate.AddDate(0, 0, c*rows+r)
			if curDate.After(endDate) {
				grid[r][c] = ContributionDay{
					Date:          curDate.Format("2006-01-02"),
					Count:         0,
					Level:         0,
					Runs:          nil,
					DistanceRatio: 0.0,
					Lightness:     15.0,
				}
				continue
			}
			dateStr := curDate.Format("2006-01-02")
			runsForDate := dateRuns[dateStr]
			distance := 0.0
			for _, run := range runsForDate {
				distance += run.DistanceKm
			}
			if distance > maxDistance {
				maxDistance = distance
			}
			grid[r][c] = ContributionDay{
				Date:  dateStr,
				Count: len(runsForDate),
				Level: 0, // will set below
				Runs:  runsForDate,
			}
		}
		// Set month label for the first row of each column
		colDate := startDate.AddDate(0, 0, c*rows)
		// Show Jan'YY at the first week of each year, else just the month
		curYear := colDate.Year()
		label := ""
		if c == 0 || curYear != prevYear {
			label = colDate.Format("Jan'06")
		} else {
			label = colDate.Format("Jan")
		}
		monthLabels[c] = label
		if c == 0 || label != prevLabel {
			monthLabelShow[c] = true
		} else {
			monthLabelShow[c] = false
		}
		prevLabel = label
		prevYear = curYear
	}

	// Trim leading all-empty columns (weeks) so the graph doesn't start with blank weeks
	firstNonEmptyCol := 0
	for c := 0; c < cols; c++ {
		empty := true
		for r := 0; r < rows; r++ {
			if grid[r][c].Count > 0 {
				empty = false
				break
			}
		}
		if !empty {
			break
		}
		firstNonEmptyCol++
	}
	if firstNonEmptyCol > 0 && firstNonEmptyCol < cols {
		newCols := cols - firstNonEmptyCol
		newGrid := make([][]ContributionDay, rows)
		for r := 0; r < rows; r++ {
			newGrid[r] = make([]ContributionDay, newCols)
			copy(newGrid[r], grid[r][firstNonEmptyCol:])
		}
		grid = newGrid
		monthLabels = monthLabels[firstNonEmptyCol:]
		monthLabelShow = monthLabelShow[firstNonEmptyCol:]
		cols = newCols
	}

	// Calculate total hours and find percentiles for better scaling
	var totalHours []float64
	for c := 0; c < cols; c++ {
		for r := 0; r < rows; r++ {
			hours := 0.0
			for _, run := range grid[r][c].Runs {
				hours += parseTimeStringToHours(run.Time)
			}
			if hours > 0 {
				totalHours = append(totalHours, hours)
			}
		}
	}

	// Calculate percentiles for smart scaling
	var p75, p90, maxHours float64
	if len(totalHours) > 0 {
		sort.Float64s(totalHours)
		p75Index := int(float64(len(totalHours)) * 0.75)
		p90Index := int(float64(len(totalHours)) * 0.90)
		if p75Index < len(totalHours) {
			p75 = totalHours[p75Index]
		}
		if p90Index < len(totalHours) {
			p90 = totalHours[p90Index]
		}
		maxHours = totalHours[len(totalHours)-1]
	}

	// Assign levels and dynamic colors using hours-based hybrid scaling
	for c := 0; c < cols; c++ {
		for r := 0; r < rows; r++ {
			hours := 0.0
			totalDistance := 0.0
			for _, run := range grid[r][c].Runs {
				hours += parseTimeStringToHours(run.Time)
				totalDistance += run.DistanceKm
			}
			level := 0
			ratio := 0.0
			isOutlier := false
			isUltraDistance := totalDistance >= 80.47 // 50 miles = 80.47 km

			if hours > 0 {
				// Check for ultra-distance activities first (50+ miles)
				if isUltraDistance {
					level = 6   // Special ultra-distance level
					ratio = 1.0 // Maximum intensity
					// Check if this is an outlier (above 90th percentile)
				} else if hours > p90 {
					isOutlier = true
					level = 5 // Special outlier level
					// For outliers, use log scale based on max hours for smoother transition
					if maxHours > p90 {
						ratio = 0.8 + 0.2*(math.Log1p(hours-p90)/math.Log1p(maxHours-p90))
					} else {
						ratio = 1.0
					}
				} else if p75 > 0 {
					// Use logarithmic scale for normal range (0 to 75th percentile)
					logMax := math.Log1p(p75)
					logHours := math.Log1p(math.Min(hours, p75))
					ratio = logHours / logMax

					// Distribute levels 1-4 across the normal range for legend compatibility
					switch {
					case ratio > 0.75:
						level = 4
					case ratio > 0.5:
						level = 3
					case ratio > 0.25:
						level = 2
					default:
						level = 1
					}
				} else {
					level = 1
					ratio = 0.5
				}
			}

			// Dynamic color calculation - correct scale: less activity = darker, more activity = lighter
			var lightness float64
			if hours == 0 {
				lightness = 15.0 // Dark background for no activity
			} else if isUltraDistance {
				// Ultra-distance activities: maximum brightness (80%)
				lightness = 80.0
			} else if isOutlier {
				// Dynamic lightness for outliers: 60% to 75% (brightest)
				lightness = 60.0 + 15.0*(ratio-0.8)/0.2
			} else {
				// Dynamic lightness for normal activities: 25% to 55% (dark to medium)
				lightness = 25.0 + 30.0*ratio
			}
			lightness = math.Round(lightness*10) / 10

			grid[r][c].Level = level
			grid[r][c].DistanceRatio = ratio
			grid[r][c].Lightness = lightness
		}
	}
	// Generate unique months list - handle special case for "Last 365 Days" and "All time"
	uniqueMonths := []string{}
	if yearFilter == "" || yearFilter == "365" || yearFilter == "all" {
		// For "Last 365 Days", we might need to show the same month twice
		// if the range spans from one year to the next (e.g., Jul 2024 -> Jul 2025)
		seenMonthsWithYear := make(map[string]bool)
		for i, label := range monthLabels {
			if monthLabelShow[i] {
				// Create a unique key combining month and position (start vs end)
				colDate := startDate.AddDate(0, 0, i*rows)
				monthYear := colDate.Format("Jan2006")

				if !seenMonthsWithYear[monthYear] {
					uniqueMonths = append(uniqueMonths, label)
					seenMonthsWithYear[monthYear] = true
				}
			}
		}
	} else {
		// For specific years, use the original logic (no duplicates needed)
		seenMonths := make(map[string]bool)
		for i, label := range monthLabels {
			if monthLabelShow[i] && !seenMonths[label] {
				uniqueMonths = append(uniqueMonths, label)
				seenMonths[label] = true
			}
		}
	}

	dayLabels := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	years := getAvailableYears(runs)
	return &ContributionGraph2D{
		Grid:           grid,
		MonthLabels:    monthLabels,
		MonthLabelShow: monthLabelShow,
		UniqueMonths:   uniqueMonths,
		DayLabels:      dayLabels,
		Weeks:          cols,
		Days:           rows,
		Years:          years,
	}, nil
}

// CanvasData returns the minimum data required to draw and filter the heatmap.
func (g *ContributionGraph2D) CanvasData() ContributionGraphCanvas {
	payload := ContributionGraphCanvas{Weeks: g.Weeks, ActiveDays: make(map[string]ContributionCanvasDay)}
	if len(g.Grid) == 0 || len(g.Grid[0]) == 0 {
		return payload
	}
	payload.StartDate = g.Grid[0][0].Date
	for _, row := range g.Grid {
		for _, day := range row {
			if day.Count == 0 {
				continue
			}
			types := make([]string, 0, len(day.Runs))
			seen := make(map[string]bool)
			for _, run := range day.Runs {
				typ := run.Type
				if typ == "" {
					typ = "Unknown"
				}
				if !seen[typ] {
					seen[typ] = true
					types = append(types, typ)
				}
			}
			payload.ActiveDays[day.Date] = ContributionCanvasDay{Count: day.Count, Level: day.Level, Lightness: day.Lightness, Types: types}
		}
	}
	return payload
}

// GetActivityTypeBreakdown calculates activity type percentages from Strava data
func GetActivityTypeBreakdown(yearFilter string) ([]ActivityTypeBreakdown, error) {
	runs, err := GetStravaRuns()
	if err != nil {
		return nil, err
	}

	// Filter runs by year if specified (same logic as GenerateContributionGraph2D)
	var filteredRuns StravaRuns
	if yearFilter == "all" {
		filteredRuns = runs
	} else if yearFilter != "" && yearFilter != "365" {
		year, err := strconv.Atoi(yearFilter)
		if err == nil {
			for _, run := range runs {
				if len(run.Date) >= 4 {
					if runYear, err := strconv.Atoi(run.Date[:4]); err == nil && runYear == year {
						filteredRuns = append(filteredRuns, run)
					}
				}
			}
		}
	} else {
		// Filter to last 365 days
		now := time.Now()
		endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		startDate := endDate.AddDate(0, 0, -364) // 365 days total

		for _, run := range runs {
			if runDate, err := time.Parse("2006-01-02", run.Date); err == nil {
				if (runDate.Equal(startDate) || runDate.After(startDate)) &&
					(runDate.Equal(endDate) || runDate.Before(endDate)) {
					filteredRuns = append(filteredRuns, run)
				}
			}
		}
	}

	if len(filteredRuns) == 0 {
		return []ActivityTypeBreakdown{}, nil
	}

	// Count activities by type
	typeCounts := make(map[string]int)
	for _, run := range filteredRuns {
		activityType := run.Type
		if activityType == "" {
			activityType = "Unknown"
		}
		typeCounts[activityType]++
	}

	// Convert to breakdown with percentages
	totalActivities := len(filteredRuns)
	breakdown := make([]ActivityTypeBreakdown, 0, len(typeCounts))

	for activityType, count := range typeCounts {
		percentage := (float64(count) / float64(totalActivities)) * 100
		// Only include activity types with 0.4% or more
		if percentage >= 0.4 {
			breakdown = append(breakdown, ActivityTypeBreakdown{
				Type:       activityType,
				Count:      count,
				Percentage: percentage,
			})
		}
	}

	// Sort by count descending (highest first)
	sort.Slice(breakdown, func(i, j int) bool {
		return breakdown[i].Count > breakdown[j].Count
	})

	return breakdown, nil
}
