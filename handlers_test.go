package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Setup test data and mocks
func setupTestEnvironment(t *testing.T) {
	// Setup test blog posts
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

	// Setup markdown map with mock content
	markdownMap = make(map[string]string)
	markdownMap["blog"] = "<p>This is test blog content</p>"
	markdownMap["home"] = "<p>This is test home content</p>"
	markdownMap["about"] = "<p>This is test about content</p>"
	markdownMap["filters"] = "<p>This is test filters content</p>"

	// Add blog post markdown content
	for _, post := range blogPosts {
		markdownMap["blog/"+fmt.Sprintf("%d", post.Id)+".markdown"] =
			"<p>This is content for blog post " + post.Title + "</p>"
	}
}

// Utility function to create a request with context vars
func newRequestWithVars(method, path string, vars map[string]string) *http.Request {
	r, _ := http.NewRequest(method, path, nil)
	if vars != nil {
		r = mux.SetURLVars(r, vars)
	}
	return r
}

// Helper for testing redirects
func assertRedirect(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedLocation string) {
	assert.Equal(t, expectedStatus, w.Code, "Expected redirect status code")
	location, err := w.Result().Location()
	assert.NoError(t, err, "Expected Location header")
	assert.Equal(t, expectedLocation, location.String(), "Unexpected redirect location")
}

// Test for errorPage handler
func TestErrorPage(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	errorPage(w, "Test error message", "blog")

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	assert.Contains(t, bodyString, "Error")
	assert.Contains(t, bodyString, "Test error message")
}

// Test for generateRSSHandler
func TestGenerateRSSHandler(t *testing.T) {
	setupTestEnvironment(t)
	imagePathWithoutLeadingSlash := "static/img/first.webp"
	imagePathWithLeadingSlash := "/static/img/second.png"
	blogPosts[0].ImagePath = &imagePathWithoutLeadingSlash
	blogPosts[1].ImagePath = &imagePathWithLeadingSlash

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/rss", nil)
	r.Host = "example.com"

	generateRSSHandler(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/xml", resp.Header.Get("Content-Type"))

	// Parse the XML to verify it's well-formed
	var rssDoc struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Title string `xml:"title"`
			Link  string `xml:"link"`
			Items []struct {
				Title     string `xml:"title"`
				Link      string `xml:"link"`
				Enclosure struct {
					URL string `xml:"url,attr"`
				} `xml:"enclosure"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	err := xml.Unmarshal(bodyBytes, &rssDoc)
	assert.NoError(t, err, "Expected valid XML")
	assert.Equal(t, "Brian R. Bondy's Blog", rssDoc.Channel.Title)
	assert.Equal(t, 3, len(rssDoc.Channel.Items), "Expected 3 items in RSS feed")
	assert.Equal(t, "https://brianbondy.com/static/img/first.webp", rssDoc.Channel.Items[0].Enclosure.URL)
	assert.Equal(t, "https://brianbondy.com/static/img/second.png", rssDoc.Channel.Items[1].Enclosure.URL)
}

func TestGenerateSitemapHandler(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/sitemap.xml", nil)

	generateSitemapHandler(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/xml; charset=utf-8", resp.Header.Get("Content-Type"))

	var sitemap sitemapURLSet
	assert.NoError(t, xml.NewDecoder(resp.Body).Decode(&sitemap))
	assert.Equal(t, "http://www.sitemaps.org/schemas/sitemap/0.9", sitemap.XMLNS)
	assert.Contains(t, sitemap.URLs, sitemapURL{Location: "https://brianbondy.com/about"})
	assert.Contains(t, sitemap.URLs, sitemapURL{Location: "https://brianbondy.com/career"})
	assert.Contains(t, sitemap.URLs, sitemapURL{
		Location: "https://brianbondy.com/blog/1/test-post-1",
		LastMod:  "2023-01-01",
	})
}

func TestRobotsHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/robots.txt", nil)

	robotsHandler(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "User-agent: *\nAllow: /\n\nSitemap: https://brianbondy.com/sitemap.xml\n", string(body))
}

// Test for filtersPageHandler
func TestFiltersPageHandler(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/filters", nil)

	filtersPageHandler(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test for cheatsheetsHandler
func TestCheatsheetsHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/cheatsheets", nil)

	cheatsheetsHandler(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	assert.Contains(t, bodyString, "Cheatsheets")
	assert.Contains(t, bodyString, "Docker")
}

// Test for cheatsheetHandler
func TestCheatsheetHandler(t *testing.T) {
	// Existing cheatsheet
	w := httptest.NewRecorder()
	r := newRequestWithVars("GET", "/cheatsheets/go", map[string]string{"slug": "go"})

	cheatsheetHandler(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	assert.Contains(t, bodyString, ">Go<")

	// Missing cheatsheet
	w = httptest.NewRecorder()
	r = newRequestWithVars("GET", "/cheatsheets/missing", map[string]string{"slug": "missing"})

	cheatsheetHandler(w, r)

	resp = w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// Test for tagRedirectHandler
func TestTagRedirectHandler(t *testing.T) {
	setupTestEnvironment(t)

	// Test with existing tag
	w := httptest.NewRecorder()
	r := newRequestWithVars("GET", "/tagged/golang", map[string]string{"tag": "golang"})

	tagRedirectHandler(w, r)

	// It should redirect to the first post with the golang tag
	assertRedirect(t, w, http.StatusFound, "/blog/1/test-post-1?tag=golang")

	// Test with non-existent tag
	w = httptest.NewRecorder()
	r = newRequestWithVars("GET", "/tagged/nonexistent", map[string]string{"tag": "nonexistent"})

	tagRedirectHandler(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Test for tagRedirectHandler with year
func TestTagRedirectHandlerWithYear(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	r := newRequestWithVars("GET", "/tagged/golang?year=2022", map[string]string{"tag": "golang"})
	r.URL.RawQuery = "year=2022"

	tagRedirectHandler(w, r)

	// It should redirect to the first post from 2022 with the golang tag
	assertRedirect(t, w, http.StatusFound, "/blog/3/test-post-3?tag=golang&year=2022")
}

// Test for redirectHandler
func TestRedirectHandler(t *testing.T) {
	setupTestEnvironment(t)

	testCases := []struct {
		path           string
		expectedTarget string
	}{
		{"/blog/page/2", "/page/2"},
		{"/blog/tagged/test", "/tagged/test"},
		{"/blog/posted/2022", "/posted/2022"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Path: %s", tc.path), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", tc.path, nil)

			redirectHandler(w, r)

			assertRedirect(t, w, http.StatusFound, tc.expectedTarget)
		})
	}
}

// Test for paginationRedirectHandler
func TestPaginationRedirectHandler(t *testing.T) {
	setupTestEnvironment(t)

	// Test with valid page
	w := httptest.NewRecorder()
	r := newRequestWithVars("GET", "/page/2", map[string]string{"page": "2"})

	paginationRedirectHandler(w, r)

	// It should redirect to the second post
	assertRedirect(t, w, http.StatusFound, "/blog/2/test-post-2?page=2")

	// Test with out-of-range page
	w = httptest.NewRecorder()
	r = newRequestWithVars("GET", "/page/10", map[string]string{"page": "10"})

	paginationRedirectHandler(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPaginationRedirectHandlerInvalidPage(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	r := newRequestWithVars("GET", "/page/abc", map[string]string{"page": "abc"})

	paginationRedirectHandler(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test for paginationRedirectHandler with tag filter
func TestPaginationRedirectHandlerWithFilters(t *testing.T) {
	setupTestEnvironment(t)

	// Test with tag filter
	w := httptest.NewRecorder()
	r := newRequestWithVars("GET", "/page/2?tag=test", map[string]string{"page": "2"})
	r.URL.RawQuery = "tag=test"

	paginationRedirectHandler(w, r)

	assertRedirect(t, w, http.StatusFound, "/blog/2/test-post-2?page=2&tag=test")

	// Test with year filter
	w = httptest.NewRecorder()
	r = newRequestWithVars("GET", "/page/1?year=2022", map[string]string{"page": "1"})
	r.URL.RawQuery = "year=2022"

	paginationRedirectHandler(w, r)

	assertRedirect(t, w, http.StatusFound, "/blog/3/test-post-3?page=1&year=2022")

	// Test with both tag and year
	w = httptest.NewRecorder()
	r = newRequestWithVars("GET", "/page/1?tag=golang&year=2022", map[string]string{"page": "1"})
	r.URL.RawQuery = "tag=golang&year=2022"

	paginationRedirectHandler(w, r)

	assertRedirect(t, w, http.StatusFound, "/blog/3/test-post-3?page=1&tag=golang&year=2022")
}

// Test for blogIdRedirectHandler
func TestBlogIdRedirectHandler(t *testing.T) {
	setupTestEnvironment(t)

	// Test with existing ID
	w := httptest.NewRecorder()
	r := newRequestWithVars("GET", "/blog/1", map[string]string{"id": "1"})

	blogIdRedirectHandler(w, r)

	assertRedirect(t, w, http.StatusMovedPermanently, "/blog/1/test-post-1")

	// Test with query parameters
	w = httptest.NewRecorder()
	r = newRequestWithVars("GET", "/blog/1?tag=test", map[string]string{"id": "1"})
	r.URL.RawQuery = "tag=test"

	blogIdRedirectHandler(w, r)

	assertRedirect(t, w, http.StatusMovedPermanently, "/blog/1/test-post-1?tag=test")

	// Test with non-existent ID
	w = httptest.NewRecorder()
	r = newRequestWithVars("GET", "/blog/999", map[string]string{"id": "999"})

	blogIdRedirectHandler(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBlogIdRedirectHandlerInvalidID(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	r := newRequestWithVars("GET", "/blog/abc", map[string]string{"id": "abc"})

	blogIdRedirectHandler(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test for yearRedirectHandler
func TestYearRedirectHandler(t *testing.T) {
	setupTestEnvironment(t)

	// Test with valid year
	w := httptest.NewRecorder()
	r := newRequestWithVars("GET", "/posted/2022", map[string]string{"year": "2022"})

	yearRedirectHandler(w, r)

	assertRedirect(t, w, http.StatusFound, "/blog/3/test-post-3?year=2022")

	// Test with year and tag
	w = httptest.NewRecorder()
	r = newRequestWithVars("GET", "/posted/2022?tag=golang", map[string]string{"year": "2022"})
	r.URL.RawQuery = "tag=golang"

	yearRedirectHandler(w, r)

	assertRedirect(t, w, http.StatusFound, "/blog/3/test-post-3?year=2022&tag=golang")

	// Test with year that has no posts
	w = httptest.NewRecorder()
	r = newRequestWithVars("GET", "/posted/2020", map[string]string{"year": "2020"})

	yearRedirectHandler(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestYearRedirectHandlerInvalidYear(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	r := newRequestWithVars("GET", "/posted/abcd", map[string]string{"year": "abcd"})

	yearRedirectHandler(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test for homePageHandler
func TestHomePageHandler(t *testing.T) {
	setupTestEnvironment(t)
	imagePath := "/static/img/test.webp"
	blogPosts[0].ImagePath = &imagePath

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	homePageHandler(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.NotContains(t, string(body), `src="//static/`)
	assert.Contains(t, string(body), `src="/static/img/test.webp"`)
}

// Test for allPostsHandler
func TestAllPostsHandler(t *testing.T) {
	setupTestEnvironment(t)

	testCases := []struct {
		name         string
		url          string
		query        string
		expectStatus int
	}{
		{"All posts", "/blog/all", "", http.StatusOK},
		{"Tag filtered", "/blog/all", "tag=test", http.StatusOK},
		{"Year filtered", "/blog/all", "year=2022", http.StatusOK},
		{"Tag and year filtered", "/blog/all", "tag=golang&year=2022", http.StatusOK},
		{"No matching posts", "/blog/all", "tag=nonexistent", http.StatusOK},
		{"Invalid year filter", "/blog/all", "year=abc", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", tc.url, nil)
			if tc.query != "" {
				r.URL.RawQuery = tc.query
			}

			allPostsHandler(w, r)

			resp := w.Result()
			assert.Equal(t, tc.expectStatus, resp.StatusCode)
		})
	}
}

// Test for blogPostPageHandler
func TestBlogPostPageHandler(t *testing.T) {
	setupTestEnvironment(t)

	testCases := []struct {
		name         string
		path         string
		vars         map[string]string
		query        string
		expectStatus int
	}{
		{
			"Valid post with ID",
			"/blog/1/test-post-1",
			map[string]string{"id": "1"},
			"",
			http.StatusOK,
		},
		{
			"Invalid post ID",
			"/blog/999/not-found",
			map[string]string{"id": "999"},
			"",
			http.StatusNotFound,
		},
		{
			"Post with tag filter",
			"/blog/1/test-post-1",
			map[string]string{"id": "1"},
			"tag=golang",
			http.StatusOK,
		},
		{
			"Post with year filter",
			"/blog/1/test-post-1",
			map[string]string{"id": "1"},
			"year=2023",
			http.StatusOK,
		},
		{
			"No ID provided, should show first post",
			"/blog",
			map[string]string{},
			"",
			http.StatusOK,
		},
		{
			"No ID with tag filter",
			"/blog",
			map[string]string{},
			"tag=golang",
			http.StatusOK,
		},
		{
			"No ID with year filter",
			"/blog",
			map[string]string{},
			"year=2022",
			http.StatusOK,
		},
		{
			"No matching posts",
			"/blog",
			map[string]string{},
			"tag=nonexistent",
			http.StatusNotFound,
		},
		{
			"Invalid year filter",
			"/blog",
			map[string]string{},
			"year=abc",
			http.StatusBadRequest,
		},
		{
			"Invalid post ID format",
			"/blog/abc/not-a-number",
			map[string]string{"id": "abc"},
			"",
			http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := newRequestWithVars("GET", tc.path, tc.vars)
			if tc.query != "" {
				r.URL.RawQuery = tc.query
			}

			blogPostPageHandler(w, r)

			assert.Equal(t, tc.expectStatus, w.Code)
		})
	}
}

// Test for getMarkdownTemplateHandler
func TestGetMarkdownTemplateHandler(t *testing.T) {
	setupTestEnvironment(t)

	// Create a handler for the about page
	handler := getMarkdownTemplateHandler("About", "about", "/about")

	// Create a test request
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/about", nil)
	r.Host = "brianbondy.com"

	// Set header to simulate HTTPS, so directToHttps won't redirect
	r.Header.Set("X-Forwarded-Proto", "https")

	// Call the handler
	handler.ServeHTTP(w, r)

	// Check response
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
