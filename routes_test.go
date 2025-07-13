package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Helper function to set up test blog data for the routes tests
func setupRoutesTestData() {
	// Reset all the global variables
	blogPosts = []data.BlogPost{
		{
			Id:      1,
			Title:   "Test Post 1",
			Created: "2023-01-01",
			Tags:    []string{"test", "golang"},
		},
		{
			Id:      2,
			Title:   "Test Post 2",
			Created: "2023-02-01",
			Tags:    []string{"test"},
		},
		{
			Id:      3,
			Title:   "Test Post 3",
			Created: "2022-01-01",
			Tags:    []string{"golang"},
		},
	}

	// Setup blog post ID map
	blogPostIdMap = make(map[int]data.BlogPost)
	for _, post := range blogPosts {
		blogPostIdMap[post.Id] = post
	}

	// Setup tag maps
	blogPostTagMap = make(map[string][]data.BlogPost)
	tagCountMap = make(map[string]int)
	for _, post := range blogPosts {
		for _, tag := range post.Tags {
			blogPostTagMap[tag] = append(blogPostTagMap[tag], post)
			tagCountMap[tag]++
		}
	}

	// Setup year map
	blogPostYearMap = make(map[int][]data.BlogPost)
	for _, post := range blogPosts {
		parsedDate, _ := time.Parse(layoutISO, post.Created)
		year := parsedDate.Year()
		blogPostYearMap[year] = append(blogPostYearMap[year], post)
	}

	// Setup sorted tags
	sortedTags = []string{"test", "golang"}

	// Setup markdown map with some test content
	markdownMap = make(map[string]string)
	markdownMap["about.markdown"] = "<h1>About</h1>"
	markdownMap["other.markdown"] = "<h1>Other</h1>"
	markdownMap["contact.markdown"] = "<h1>Contact</h1>"

	markdownMap["advice.markdown"] = "<h1>Advice</h1>"
	markdownMap["books.markdown"] = "<h1>Books</h1>"
	markdownMap["resume.markdown"] = "<h1>Resume</h1>"
	markdownMap["running.markdown"] = "<h1>Running</h1>"

	// Add blog post markdown content
	for _, post := range blogPosts {
		markdownMap["blog/"+string(rune(post.Id))+".markdown"] =
			"<p>This is content for blog post " + post.Title + "</p>"
	}
}

// TestRouteRegistration verifies that all expected routes are registered
func TestRouteRegistration(t *testing.T) {
	// Save original data
	origBlogPosts := blogPosts
	origBlogPostIdMap := blogPostIdMap
	origBlogPostTagMap := blogPostTagMap
	origBlogPostYearMap := blogPostYearMap
	origTagCountMap := tagCountMap
	origSortedTags := sortedTags
	origMarkdownMap := markdownMap

	// Restore after test
	defer func() {
		blogPosts = origBlogPosts
		blogPostIdMap = origBlogPostIdMap
		blogPostTagMap = origBlogPostTagMap
		blogPostYearMap = origBlogPostYearMap
		tagCountMap = origTagCountMap
		sortedTags = origSortedTags
		markdownMap = origMarkdownMap
	}()

	// Create router and initialize routes
	router := mux.NewRouter()
	initializeRoutes(router)

	// Setup mock for blog data
	setupRoutesTestData()

	// Define routes to test
	testCases := []struct {
		name          string
		method        string
		path          string
		expectedMatch bool
	}{
		{"Home route", "GET", "/", true},
		{"RSS route", "GET", "/rss", true},
		{"Blog post with ID and slug", "GET", "/blog/1/test-post-1", true},
		{"Blog ID redirect", "GET", "/blog/1", true},
		{"Pagination redirect", "GET", "/page/2", true},
		{"Tag redirect", "GET", "/tagged/golang", true},
		{"Year redirect", "GET", "/posted/2022", true},
		{"Year with page redirect", "GET", "/posted/2022/page/1", true},
		{"Legacy blog pagination redirect", "GET", "/blog/page/2", true},
		{"Legacy blog tag redirect", "GET", "/blog/tagged/golang", true},
		{"Filters page", "GET", "/blog/filters", true},
		{"About page", "GET", "/about", true},
		{"Projects page", "GET", "/projects", true},
		{"Resume page", "GET", "/resume", true},
		{"All posts page", "GET", "/all", true},
		{"Non-existent route", "GET", "/this-does-not-exist", false},
	}

	// Test each route
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest(tc.method, tc.path, nil)
			assert.NoError(t, err)

			// Match route
			var match mux.RouteMatch
			matched := router.Match(req, &match)
			if matched && match.Route == nil {
				matched = false
			}
			assert.Equal(t, tc.expectedMatch, matched, "Route match status unexpected")
		})
	}
}

// TestURLParameterExtraction tests that URL parameters are correctly extracted
func TestURLParameterExtraction(t *testing.T) {
	// Create router and initialize routes
	router := mux.NewRouter()
	initializeRoutes(router)

	testCases := []struct {
		name       string
		path       string
		paramName  string
		paramValue string
	}{
		{"Blog post ID", "/blog/123/some-slug", "id", "123"},
		{"Blog post slug", "/blog/123/some-slug", "slug", "some-slug"},
		{"Pagination page number", "/page/42", "page", "42"},
		{"Tag name", "/tagged/golang", "tag", "golang"},
		{"Year", "/posted/2022", "year", "2022"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tc.path, nil)
			var match mux.RouteMatch
			matched := router.Match(req, &match)

			assert.True(t, matched, "Route should match")
			assert.Contains(t, match.Vars, tc.paramName, "URL param should be extracted")
			assert.Equal(t, tc.paramValue, match.Vars[tc.paramName], "URL param value should match")
		})
	}
}

// TestStaticFileServer verifies the static file server is set up correctly
func TestStaticFileServer(t *testing.T) {
	// Create router and initialize routes
	router := mux.NewRouter()
	initializeRoutes(router)

	// Create request for static file
	req, _ := http.NewRequest("GET", "/static/test.css", nil)

	// Check if route matches
	var match mux.RouteMatch
	matched := router.Match(req, &match)

	assert.True(t, matched, "Static file route should match")
}

// TestMiddlewareChain verifies that middleware is correctly applied
func TestMiddlewareChain(t *testing.T) {
	router := mux.NewRouter()
	initializeRoutes(router)

	// Setup test environment
	setupRoutesTestData()

	// Test https redirect middleware
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Send request without HTTPS header
	router.ServeHTTP(w, req)

	// Should redirect to HTTPS
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location, err := w.Result().Location()
	assert.NoError(t, err)
	assert.Equal(t, "https://"+req.Host+"/", location.String())
}
