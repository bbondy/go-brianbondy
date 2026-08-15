package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

const (
	layoutISO = "2006-01-02"
	layoutUS  = "January 2, 2006"
	siteURL   = "https://brianbondy.com"
)

// getErrorMessageForCode returns a user-friendly message for a given HTTP error code.
func getErrorMessageForCode(code int) string {
	switch code {
	case http.StatusNotFound:
		return "Move along. Move along."
	case http.StatusInternalServerError:
		return "An unexpected server error occurred. Please try again later."
	case http.StatusBadRequest:
		return "The request could not be understood by the server."
	case http.StatusForbidden:
		return "You do not have permission to access this page."
	case http.StatusUnauthorized:
		return "You are not authorized to view this page."
	case http.StatusMethodNotAllowed:
		return "The method is not allowed for the requested URL."
	case http.StatusRequestTimeout:
		return "The request timed out. Please try again."
	case http.StatusTooManyRequests:
		return "You have made too many requests. Please slow down."
	case http.StatusServiceUnavailable:
		return "The service is temporarily unavailable. Please try again later."
	case http.StatusGatewayTimeout:
		return "The server did not receive a timely response."
	default:
		return "An error occurred."
	}
}

func errorPage(w http.ResponseWriter, message string, slug string) {
	errorPageWithStatus(w, message, slug, http.StatusNotFound)
}

func errorPageWithStatus(w http.ResponseWriter, message string, slug string, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	var title string
	switch statusCode {
	case http.StatusNotFound:
		title = "These aren't the pages you're looking for."
	case http.StatusInternalServerError:
		title = "Server Error"
	case http.StatusBadRequest:
		title = "Bad Request"
	case http.StatusForbidden:
		title = "Access Forbidden"
	case http.StatusUnauthorized:
		title = "Unauthorized"
	case http.StatusMethodNotAllowed:
		title = "Method Not Allowed"
	case http.StatusRequestTimeout:
		title = "Request Timeout"
	case http.StatusTooManyRequests:
		title = "Too Many Requests"
	case http.StatusServiceUnavailable:
		title = "Service Unavailable"
	case http.StatusGatewayTimeout:
		title = "Gateway Timeout"
	default:
		title = "Error"
	}

	if message == "" {
		message = getErrorMessageForCode(statusCode)
	}

	p := &data.SimpleMarkdownPage{
		Title:        title,
		Content:      message,
		MarkdownSlug: slug,
		ErrorCode:    statusCode,
	}

	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/error.html"))
	err := t.Execute(w, p)
	if err != nil {
		log.Printf("Error executing error page template: %v", err)
		return
	}
}

func parseIntParam(w http.ResponseWriter, value string, name string, slug string) (int, bool) {
	if value == "" {
		return 0, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		errorPageWithStatus(w, fmt.Sprintf("Invalid %s value", name), slug, http.StatusBadRequest)
		return 0, false
	}
	return parsed, true
}

func testErrorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	errorCodeStr := vars["errorcode"]

	errorCode, err := strconv.Atoi(errorCodeStr)
	if err != nil {
		http.Error(w, "Invalid error code", http.StatusBadRequest)
		return
	}

	// Validate that it's a 4xx or 5xx error code
	if errorCode < 400 || errorCode >= 600 {
		http.Error(w, "Error code must be between 400-599", http.StatusBadRequest)
		return
	}

	message := getErrorMessageForCode(errorCode)
	errorPageWithStatus(w, message, "", errorCode)
}

func getMarkdownTemplateHandler(titleSlug string, markdownSlug string, fbShareUrl string) *negroni.Negroni {
	handler := func(w http.ResponseWriter, r *http.Request) {
		p := &data.SimpleMarkdownPage{
			Title:        GetTitle(titleSlug),
			Content:      getMarkdownData(markdownSlug),
			MarkdownSlug: markdownSlug,
			ShareUrl:     fbShareUrl,
		}
		t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/simpleMarkdown.html"))
		err := t.Execute(w, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(handler)))
}

func runningHandler(w http.ResponseWriter, r *http.Request) {
	runs, err := data.GetRuns()
	if err != nil {
		errorPage(w, "Unable to load run data", "running")
		return
	}

	// Get year filter from query parameter, default to "all"
	yearFilter := r.URL.Query().Get("year")
	if yearFilter == "" {
		yearFilter = "all"
	}

	// Get Strava totals for selected view
	totalRuns, totalDistanceKm, totalElevationM, totalTimeMinutes, err := data.GetStravaRunTotalsFor(yearFilter)
	if err != nil {
		totalRuns = 0
		totalDistanceKm = 0
		totalElevationM = 0
		totalTimeMinutes = 0
	}
	timeDays, timeHours, timeMinutes := data.SplitMinutesToDaysHoursMinutes(totalTimeMinutes)

	// Generate 2D contribution graph with month and day labels
	contributionGraph2D, err := data.GenerateContributionGraph2D(yearFilter)
	if err != nil {
		contributionGraph2D = nil
	}
	// Remove debug print
	// if contributionGraph2D != nil {
	// 	log.Printf("DEBUG: runs=%d, weeks=%d, days=%d", len(runs), contributionGraph2D.Weeks, contributionGraph2D.Days)
	// } else {
	// 	log.Printf("DEBUG: contributionGraph2D is nil")
	// }

	stravaTotals := data.StravaRunTotals{
		TotalRuns:        totalRuns,
		TotalDistanceKm:  totalDistanceKm,
		TotalElevationM:  totalElevationM,
		TotalTimeDays:    timeDays,
		TotalTimeHours:   timeHours,
		TotalTimeMinutes: timeMinutes,
	}

	// Get activity type breakdown for the selected year filter
	activityBreakdown, err := data.GetActivityTypeBreakdown(yearFilter)
	if err != nil {
		activityBreakdown = []data.ActivityTypeBreakdown{}
	}

	p := &data.RunningPage{
		Title:                 GetTitle("Running"),
		MarkdownSlug:          "running",
		Runs:                  runs,
		ContributionGraph:     nil, // not used in new template
		StravaRunTotals:       stravaTotals,
		ActivityTypeBreakdown: activityBreakdown,
	}

	// Get last updated date
	lastUpdated, err := data.GetLastUpdatedDate()
	if err != nil {
		lastUpdated = "Unknown"
	}

	// Pass the 2D graph as a separate variable
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/running.html"))
	err = t.Execute(w, map[string]interface{}{
		"Page":                p,
		"ContributionGraph2D": contributionGraph2D,
		"Years": func() []int {
			if contributionGraph2D != nil {
				return contributionGraph2D.Years
			}
			return nil
		}(),
		"SelectedYear":    yearFilter,
		"LastUpdatedDate": lastUpdated,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// runsForDateHandler handles AJAX requests for runs on a specific date
func runsForDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "Date parameter required", http.StatusBadRequest)
		return
	}

	runs, err := data.GetRunsForDate(date)
	if err != nil {
		http.Error(w, "Unable to load runs for date", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(runs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
	projects, err := data.GetProjects()
	if err != nil {
		errorPage(w, "Unable to load project data", "projects")
		return
	}

	// Build a map of blog post IDs to blog post data for use in the template
	blogPostMap := make(map[int]data.BlogPost)
	for _, post := range blogPosts {
		blogPostMap[post.Id] = post
	}

	p := struct {
		Title        string
		MarkdownSlug string
		Projects     data.Projects
		BlogPostMap  map[int]data.BlogPost
	}{
		Title:        GetTitle("Projects"),
		MarkdownSlug: "projects",
		Projects:     projects,
		BlogPostMap:  blogPostMap,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/projects.html"))
	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func cheatsheetsHandler(w http.ResponseWriter, r *http.Request) {
	cheatsheets, err := data.GetCheatsheets()
	if err != nil {
		errorPage(w, "Unable to load cheatsheets", "cheatsheets")
		return
	}

	p := struct {
		Title        string
		MarkdownSlug string
		Cheatsheets  data.Cheatsheets
	}{
		Title:        GetTitle("Cheatsheets"),
		MarkdownSlug: "cheatsheets",
		Cheatsheets:  cheatsheets,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/cheatsheets.html"))
	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func cheatsheetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	if slug == "" {
		errorPageWithStatus(w, "Cheatsheet not found", "cheatsheets", http.StatusNotFound)
		return
	}

	cheatsheets, err := data.GetCheatsheets()
	if err != nil {
		errorPage(w, "Unable to load cheatsheets", "cheatsheets")
		return
	}

	var found *data.Cheatsheet
	for i := range cheatsheets {
		if cheatsheets[i].Slug == slug {
			found = &cheatsheets[i]
			break
		}
	}
	if found == nil {
		errorPageWithStatus(w, "Cheatsheet not found", "cheatsheets", http.StatusNotFound)
		return
	}

	p := &data.SimpleMarkdownPage{
		Title:        GetTitle(found.Title),
		Content:      getMarkdownData("cheatsheets/" + slug + ".md"),
		MarkdownSlug: "cheatsheets",
		ShareUrl:     "/cheatsheets/" + slug,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/simpleMarkdown.html"))
	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func interviewsHandler(w http.ResponseWriter, r *http.Request) {
	interviews, err := data.GetInterviews()
	if err != nil {
		errorPage(w, "Unable to load interview data", "interviews")
		return
	}

	p := struct {
		Title        string
		MarkdownSlug string
		Interviews   data.Interviews
	}{
		Title:        GetTitle("Interviews"),
		MarkdownSlug: "interviews",
		Interviews:   interviews,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/interviews.html"))
	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func generateRSSHandler(w http.ResponseWriter, r *http.Request) {
	feed := &feeds.Feed{
		Title:       "Brian R. Bondy's Blog",
		Link:        &feeds.Link{Href: "https://" + r.Host},
		Description: "Brian R. Bondy's Blog - Coding, Running, and Life",
		Author:      &feeds.Author{Name: "Brian R. Bondy"},
		Created:     time.Now(),
		Image: &feeds.Image{
			Url:    fmt.Sprintf("https://%s/static/img/avatar.png", r.Host),
			Title:  "Brian R. Bondy's Blog",
			Link:   fmt.Sprintf("https://%s", r.Host),
			Width:  200,
			Height: 200,
		},
	}

	var items []*feeds.Item
	for _, post := range blogPosts {
		parsedDate, _ := time.Parse(layoutISO, post.Created)
		fullContent := getMarkdownData("blog/" + strconv.Itoa(post.Id) + ".markdown")

		// Try to get description from content first
		description := extractFirstParagraph(fullContent)

		// If description is empty, use post.Description as fallback
		if description == "" && post.Description != nil {
			description = *post.Description
		}

		// If still empty, use a default description
		if description == "" {
			description = "Read more about " + post.Title
		}

		// Create the full URL for both link and guid
		postURL := fmt.Sprintf("https://%s/blog/%d/%s", r.Host, post.Id, slugifyTitle(post.Title))
		guidURL := fmt.Sprintf("https://%s/blog/%d", r.Host, post.Id)

		item := &feeds.Item{
			Title:       post.Title,
			Link:        &feeds.Link{Href: postURL},
			Description: description,
			Author:      &feeds.Author{Name: "Brian R. Bondy"},
			Created:     parsedDate,
			Id:          guidURL, // This sets the GUID
		}

		// Add image enclosure if available
		if post.ImagePath != nil && *post.ImagePath != "" {
			imageURL := fmt.Sprintf("https://%s%s", r.Host, *post.ImagePath)
			mimeType := getImageMimeType(*post.ImagePath)
			if mimeType != "" {
				item.Enclosure = &feeds.Enclosure{
					Url:    imageURL,
					Type:   mimeType,
					Length: "0", // Setting length to 0 as we don't have the file size
				}
			}
		}

		items = append(items, item)
	}
	feed.Items = items

	rss, err := feed.ToRss()
	if err != nil {
		http.Error(w, "Error generating RSS feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	_, err = w.Write([]byte(rss))
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Location string `xml:"loc"`
	LastMod  string `xml:"lastmod,omitempty"`
}

func generateSitemapHandler(w http.ResponseWriter, _ *http.Request) {
	paths := []string{
		"/",
		"/about",
		"/advice",
		"/all",
		"/books",
		"/cheatsheets",
		"/contact",
		"/interviews",
		"/pictures",
		"/projects",
		"/resume",
		"/running",
	}

	urls := make([]sitemapURL, 0, len(paths)+len(blogPosts))
	for _, path := range paths {
		urls = append(urls, sitemapURL{Location: siteURL + path})
	}
	for _, post := range blogPosts {
		urls = append(urls, sitemapURL{
			Location: siteURL + "/blog/" + strconv.Itoa(post.Id) + "/" + slugifyTitle(post.Title),
			LastMod:  post.Created,
		})
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		log.Printf("Error writing sitemap XML header: %v", err)
		return
	}
	if err := xml.NewEncoder(w).Encode(sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}); err != nil {
		log.Printf("Error generating sitemap: %v", err)
	}
}

func filtersPageHandler(w http.ResponseWriter, r *http.Request) {
	current_year := time.Now().Year()
	start_year := 2005
	year_range := make([]int, current_year-start_year+1)
	for i := range year_range {
		year_range[i] = current_year - i
	}

	p := &data.FiltersPage{
		Title:        GetTitle("Filters"),
		Content:      "Test content - filters",
		TagCountMap:  tagCountMap,
		SortedTags:   sortedTags,
		MarkdownSlug: "filters",
		Years:        year_range,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/filters.html"))
	err := t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func tagRedirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tag := vars["tag"]
	year := 0
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		year, _ = strconv.Atoi(yearStr)
	}

	// Get filtered posts
	filteredPosts := getFilteredPosts(tag, year)

	if len(filteredPosts) > 0 {
		// Redirect to the first post with the tag/year filters as query params
		firstPost := filteredPosts[0]
		target := fmt.Sprintf("/blog/%d/%s?tag=%s",
			firstPost.Id,
			slugifyTitle(firstPost.Title),
			tag)
		if year != 0 {
			target += fmt.Sprintf("&year=%d", year)
		}
		http.Redirect(w, r, target, http.StatusFound)
		return
	}

	// If no posts found, show error
	errorPage(w, "No blog posts found with that tag", "blog")
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	replacements := map[string]string{
		"/blog/page/":   "/page/",
		"/blog/tagged/": "/tagged/",
		"/blog/posted/": "/posted/",
	}
	for from, to := range replacements {
		path = strings.ReplaceAll(path, from, to)
	}
	http.Redirect(w, r, path, http.StatusFound)
}

func paginationRedirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, ok := parseIntParam(w, vars["page"], "page", "blog")
	if !ok {
		return
	}

	// Get tag and year from query parameters
	tag := r.URL.Query().Get("tag")
	year := 0
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		parsedYear, parseOk := parseIntParam(w, yearStr, "year", "blog")
		if !parseOk {
			return
		}
		year = parsedYear
	}

	// Get filtered posts
	filteredPosts := getFilteredPosts(tag, year)

	// Convert page number to post index (0-based)
	postIndex := page - 1
	if postIndex < 0 || postIndex >= len(filteredPosts) {
		errorPage(w, "Invalid page number", "blog")
		return
	}

	// Redirect to the post at that index
	post := filteredPosts[postIndex]
	target := fmt.Sprintf("/blog/%d/%s",
		post.Id,
		slugifyTitle(post.Title))

	// Add query parameters if present
	params := make([]string, 0)
	if page > 0 {
		params = append(params, fmt.Sprintf("page=%d", page))
	}
	if tag != "" {
		params = append(params, fmt.Sprintf("tag=%s", tag))
	}
	if year != 0 {
		params = append(params, fmt.Sprintf("year=%d", year))
	}
	if len(params) > 0 {
		target += "?" + strings.Join(params, "&")
	}

	http.Redirect(w, r, target, http.StatusFound)
}

func blogIdRedirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := parseIntParam(w, vars["id"], "id", "blog")
	if !ok {
		return
	}

	if post, ok := blogPostIdMap[id]; ok {
		// Build the canonical URL with the slug
		target := fmt.Sprintf("/blog/%d/%s", id, slugifyTitle(post.Title))

		// Preserve any query parameters
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}

		http.Redirect(w, r, target, http.StatusMovedPermanently)
		return
	}

	errorPage(w, "Blog post not found", "blog")
}

func yearRedirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	year, ok := parseIntParam(w, vars["year"], "year", "blog")
	if !ok {
		return
	}

	// Get tag from query parameters
	tag := r.URL.Query().Get("tag")

	// Get filtered posts
	filteredPosts := getFilteredPosts(tag, year)

	if len(filteredPosts) > 0 {
		// Redirect to the first post with the year filter as query param
		firstPost := filteredPosts[0]
		target := fmt.Sprintf("/blog/%d/%s?year=%d",
			firstPost.Id,
			slugifyTitle(firstPost.Title),
			year)

		// Add tag if present
		if tag != "" {
			target += fmt.Sprintf("&tag=%s", tag)
		}

		http.Redirect(w, r, target, http.StatusFound)
		return
	}

	// If no posts found, show error
	errorPage(w, "No blog posts found for that year", "blog")
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	const previewCount = 4

	// Get the most recent posts for preview cards
	previewPosts := make([]data.BlogPostPreview, 0, previewCount)
	for i := 0; i < previewCount && i < len(blogPosts); i++ {
		post := blogPosts[i]
		parsedDate, _ := time.Parse(layoutISO, post.Created)

		fullContent := getMarkdownData("blog/" + strconv.Itoa(post.Id) + ".markdown")
		preview := extractFirstParagraph(fullContent)

		previewPosts = append(previewPosts, data.BlogPostPreview{
			BlogPost:    post,
			Preview:     template.HTML(preview),
			PostDate:    parsedDate.Format(layoutUS),
			PostUrl:     fmt.Sprintf("/blog/%d/%s", post.Id, slugifyTitle(post.Title)),
			ReadingTime: post.ReadingTime,
		})
	}

	// Get all posts for the list
	allPosts := make([]data.BlogPostPreview, 0, len(blogPosts))
	for _, post := range blogPosts {
		parsedDate, _ := time.Parse(layoutISO, post.Created)
		allPosts = append(allPosts, data.BlogPostPreview{
			BlogPost:    post,
			PostDate:    parsedDate.Format(layoutUS),
			PostUrl:     fmt.Sprintf("/blog/%d/%s", post.Id, slugifyTitle(post.Title)),
			ReadingTime: post.ReadingTime,
		})
	}

	p := &data.HomePage{
		Title:        "Brian R. Bondy",
		Posts:        previewPosts,
		AllPosts:     allPosts,
		MarkdownSlug: "home",
	}

	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/home.html"))
	err := t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func allPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Get tag and year from query parameters
	tag := r.URL.Query().Get("tag")
	year := 0
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		parsedYear, parseOk := parseIntParam(w, yearStr, "year", "blog")
		if !parseOk {
			return
		}
		year = parsedYear
	}

	// Get filtered posts based on tag and/or year
	filteredPosts := getFilteredPosts(tag, year)

	allPosts := make([]data.BlogPostPreview, 0, len(filteredPosts))
	for _, post := range filteredPosts {
		parsedDate, _ := time.Parse(layoutISO, post.Created)
		allPosts = append(allPosts, data.BlogPostPreview{
			BlogPost:    post,
			PostDate:    parsedDate.Format(layoutUS),
			PostUrl:     fmt.Sprintf("/blog/%d/%s", post.Id, slugifyTitle(post.Title)),
			ReadingTime: post.ReadingTime,
		})
	}

	title := "All Blog Posts"
	if tag != "" {
		title = fmt.Sprintf("Blog Posts Tagged with \"%s\"", tag)
	}
	if year != 0 {
		if tag != "" {
			title = fmt.Sprintf("Blog Posts from %d Tagged with \"%s\"", year, tag)
		} else {
			title = fmt.Sprintf("Blog Posts from %d", year)
		}
	}

	p := &data.AllPostsPage{
		Title:        GetTitle(title),
		Posts:        allPosts,
		MarkdownSlug: "all",
		Tag:          tag,
		Year:         year,
	}

	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/allPosts.html"))
	err := t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Keeps it simple with 1 blog post per page
func blogPostPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Get tag and year from either URL vars or query parameters
	tag := vars["tag"]
	year := 0

	// If not in URL vars, check query parameters
	if tag == "" {
		tag = r.URL.Query().Get("tag")
	}
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		parsedYear, parseOk := parseIntParam(w, yearStr, "year", "blog")
		if !parseOk {
			return
		}
		year = parsedYear
	} else if yearStr, ok := vars["year"]; ok {
		parsedYear, parseOk := parseIntParam(w, yearStr, "year", "blog")
		if !parseOk {
			return
		}
		year = parsedYear
	}

	filteredBlogPosts := getFilteredPosts(tag, year)

	// Handle individual blog post view
	if idStr, ok := vars["id"]; ok {
		id, parseOk := parseIntParam(w, idStr, "id", "blog")
		if !parseOk {
			return
		}
		if foundPost, ok := blogPostIdMap[id]; ok {
			// Get the filtered posts based on tag/year
			filteredPosts := getFilteredPosts(tag, year)

			currentIndex := -1
			for i, post := range filteredPosts {
				if post.Id == id {
					currentIndex = i
					break
				}
			}

			var nextPost, prevPost *data.BlogPost
			if currentIndex > 0 {
				prevPost = &filteredPosts[currentIndex-1]
			}
			if currentIndex < len(filteredPosts)-1 {
				nextPost = &filteredPosts[currentIndex+1]
			}

			parsedDate, _ := time.Parse(layoutISO, foundPost.Created)

			p := &data.BlogPostPage{
				Title:        GetTitle(foundPost.Title),
				BlogPost:     foundPost,
				BlogPostBody: getMarkdownData("blog/" + strconv.Itoa(foundPost.Id) + ".markdown"),
				BlogPostDate: parsedDate.Format(layoutUS),
				ReadingTime:  foundPost.ReadingTime,
				NextPost:     nextPost,
				PrevPost:     prevPost,
				Tag:          tag,
				Year:         year,
				ImagePath:    derefString(foundPost.ImagePath),
				Description:  derefString(foundPost.Description),
				ShareUrl:     fmt.Sprintf("/blog/%d/%s", foundPost.Id, slugifyTitle(foundPost.Title)),
				MarkdownSlug: "blog",
			}
			t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/blogPost.html"))
			err := t.Execute(w, p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		errorPage(w, "No blog posts for this query", "blog")
	}

	// Handle root URL or other listing pages
	if len(filteredBlogPosts) > 0 {
		post := filteredBlogPosts[0]
		parsedDate, _ := time.Parse(layoutISO, post.Created)

		// Set up next post for the first post
		var nextPost *data.BlogPost
		if len(filteredBlogPosts) > 1 {
			nextPost = &filteredBlogPosts[1]
		}

		p := &data.BlogPostPage{
			Title:        GetTitle("Blog posts"),
			BlogPost:     post,
			BlogPostBody: getMarkdownData("blog/" + strconv.Itoa(post.Id) + ".markdown"),
			BlogPostDate: parsedDate.Format(layoutUS),
			ReadingTime:  post.ReadingTime,
			NextPost:     nextPost,
			Tag:          tag,
			Year:         year,
			ImagePath:    derefString(post.ImagePath),
			Description:  derefString(post.Description),
			MarkdownSlug: "blog",
		}
		t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/blogPost.html"))
		err := t.Execute(w, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	errorPage(w, "No blog posts found", "blog")
}

var cachedPictures []data.Picture
var picturesLoaded bool

func getCachedPictures() ([]data.Picture, error) {
	if picturesLoaded {
		return cachedPictures, nil
	}
	f, err := os.Open("data/picturesManifest.json")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()
	var pictures []data.Picture
	if err := json.NewDecoder(f).Decode(&pictures); err != nil {
		return nil, err
	}
	cachedPictures = pictures
	picturesLoaded = true
	return cachedPictures, nil
}

func ClearPicturesCache() {
	picturesLoaded = false
	cachedPictures = nil
}

func picturesHandler(w http.ResponseWriter, r *http.Request) {
	pictures, err := getCachedPictures()
	if err != nil {
		errorPage(w, "Could not open pictures manifest", "pictures")
		return
	}

	tag := r.URL.Query().Get("tag")
	blogID := r.URL.Query().Get("blog_id")

	var filterBlogID int
	var blogPostTitle string
	if blogID != "" {
		blogIDInt, err := strconv.Atoi(blogID)
		if err == nil {
			filterBlogID = blogIDInt
			if blogPost, exists := blogPostIdMap[blogIDInt]; exists {
				blogPostTitle = blogPost.Title
			}
		}
	}

	if tag != "" {
		filtered := make([]data.Picture, 0, len(pictures))
		tagLower := strings.ToLower(tag)
		for _, pic := range pictures {
			for _, t := range pic.Tags {
				if strings.ToLower(t) == tagLower {
					filtered = append(filtered, pic)
					break
				}
			}
		}
		pictures = filtered
	}

	if blogID != "" {
		blogIDInt, err := strconv.Atoi(blogID)
		if err == nil {
			filtered := make([]data.Picture, 0, len(pictures))
			for _, pic := range pictures {
				if pic.Id == blogIDInt {
					filtered = append(filtered, pic)
				}
			}
			pictures = filtered
		}
	}

	p := &data.PicturesPage{
		Title:         "Pictures Gallery",
		MarkdownSlug:  "pictures",
		Pictures:      pictures,
		FilterTag:     tag,
		FilterBlogID:  filterBlogID,
		BlogPostTitle: blogPostTitle,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/pictures.html"))
	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Books handler and route
func booksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := data.GetBooks()
	if err != nil {
		errorPage(w, "Could not load books", "books")
		return
	}
	p := struct {
		Title        string
		MarkdownSlug string
		Books        data.Books
		BookCount    int
	}{
		Title:        GetTitle("Books"),
		MarkdownSlug: "books",
		Books:        books,
		BookCount:    len(books),
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/books.html"))
	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
