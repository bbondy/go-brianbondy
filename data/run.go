package data

import (
	"math"
	"sort"
	"strconv"
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
	sort.Sort(runs)
	cachedStravaRuns = runs
	stravaRunsLoaded = true
	return cachedStravaRuns, nil
}

// ClearStravaRunsCache clears the cached Strava runs (for testing or reload)
func ClearStravaRunsCache() {
	stravaRunsLoaded = false
	cachedStravaRuns = nil
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

// GenerateContributionGraph2D returns a 2D grid for the contribution graph, with month and day labels
func GenerateContributionGraph2D(yearFilter string) (*ContributionGraph2D, error) {
	runs, err := GetStravaRuns()
	if err != nil {
		return nil, err
	}
	var filteredRuns StravaRuns
	var startDate, endDate time.Time
	if yearFilter != "" && yearFilter != "365" {
		year, err := strconv.Atoi(yearFilter)
		if err == nil {
			startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			endDate = time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
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
		startDate = endDate.AddDate(0, 0, -364)
	}

	// Map date to runs
	dateRuns := make(map[string][]StravaRun)
	for _, run := range filteredRuns {
		dateRuns[run.Date] = append(dateRuns[run.Date], run)
	}

	rows := 7
	cols := 53
	grid := make([][]ContributionDay, rows)
	for r := 0; r < rows; r++ {
		grid[r] = make([]ContributionDay, cols)
	}
	monthLabels := make([]string, cols)
	monthLabelShow := make([]bool, cols)
	var prevLabel string
	maxDistance := 0.0
	for c := 0; c < cols; c++ {
		for r := 0; r < rows; r++ {
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
		label := colDate.Format("Jan")
		monthLabels[c] = label
		if c == 0 || label != prevLabel {
			monthLabelShow[c] = true
		} else {
			monthLabelShow[c] = false
		}
		prevLabel = label
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
	// Generate unique months list - handle special case for "Last 365 Days"
	uniqueMonths := []string{}
	if yearFilter == "" || yearFilter == "365" {
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

	dayLabels := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
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

// GetActivityTypeBreakdown calculates activity type percentages from Strava data
func GetActivityTypeBreakdown(yearFilter string) ([]ActivityTypeBreakdown, error) {
	runs, err := GetStravaRuns()
	if err != nil {
		return nil, err
	}

	// Filter runs by year if specified (same logic as GenerateContributionGraph2D)
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
