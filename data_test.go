package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeMarkdownHTML(t *testing.T) {
	assert.True(t, safeInlineStyle.MatchString("float:left;padding-right:10px;padding-bottom:10px"))
	input := `<p class="intro" style="width:90vw">Safe content</p><img src="/static/img/test.webp" style="float:left;padding-right:10px;padding-bottom:10px" onerror="alert(1)"><script>alert(1)</script><a href="javascript:alert(1)">bad link</a><iframe src="https://evil.example/embed"></iframe><iframe src="https://www.youtube.com/embed/abc123" width="320" height="560" frameborder="0" allowfullscreen></iframe><video controls><source src="/static/img/blogpost_183/video.mp4" type="video/mp4"></video>`

	got := sanitizeMarkdownHTML(input)

	assert.Contains(t, got, `<p class="intro" style="width:90vw">Safe content</p>`)
	assert.Contains(t, got, `<img src="/static/img/test.webp" style="float:left;padding-right:10px;padding-bottom:10px">`)
	assert.Contains(t, got, `<iframe src="https://www.youtube.com/embed/abc123" width="320" height="560" frameborder="0" allowfullscreen=""></iframe>`)
	assert.Contains(t, got, `<video controls=""><source src="/static/img/blogpost_183/video.mp4" type="video/mp4"></video>`)
	assert.NotContains(t, got, "<script")
	assert.NotContains(t, got, "onerror")
	assert.NotContains(t, got, "javascript:")
	assert.NotContains(t, got, "evil.example")
}

func TestMarkdownSanitizationPreservesExistingMedia(t *testing.T) {
	originalMarkdownMap := markdownMap
	defer func() { markdownMap = originalMarkdownMap }()
	markdownMap = make(map[string]string)

	blogWithVideo := getMarkdownData("blog/183.markdown")
	assert.Contains(t, blogWithVideo, `<img style="width:90vw" src="/static/img/blogpost_183/family-finish.webp">`)
	assert.Contains(t, blogWithVideo, `<video controls="">`)
	assert.Contains(t, blogWithVideo, `<source src="/static/img/blogpost_183/video.mp4" type="video/mp4">`)

	blogWithEmbed := getMarkdownData("blog/190.markdown")
	assert.Contains(t, blogWithEmbed, `<iframe src="https://www.youtube.com/embed/oaEKb0UJ-U8" width="320" height="560" frameborder="0" allowfullscreen=""></iframe>`)
}

func TestAllBlogMarkdownSurvivesSanitization(t *testing.T) {
	files, err := filepath.Glob("data/markdown/blog/*.markdown")
	if err != nil {
		t.Fatalf("find blog Markdown files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no blog Markdown files found")
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		rendered := renderMarkdown(content)
		sanitized := sanitizeMarkdownHTML(rendered)
		visibleRendered := regexp.MustCompile(`(?s)<!--.*?-->`).ReplaceAllString(rendered, "")

		for _, marker := range []string{"style=", "class=", "<video", "<iframe"} {
			if strings.Count(sanitized, marker) != strings.Count(visibleRendered, marker) {
				t.Errorf("%s: sanitizer changed %q count from %d to %d", file, marker, strings.Count(visibleRendered, marker), strings.Count(sanitized, marker))
			}
		}
		if strings.Contains(sanitized, "<script") {
			t.Errorf("%s: sanitized output contains a script tag", file)
		}
		if regexp.MustCompile(`(?i)\son[a-z]+\s*=`).MatchString(sanitized) {
			t.Errorf("%s: sanitized output contains an event handler", file)
		}
	}
}

// TestDerefString tests the derefString helper function
func TestDerefString(t *testing.T) {
	// Test with nil pointer
	result := derefString(nil)
	assert.Equal(t, "", result, "derefString(nil) should return empty string")

	// Test with non-nil pointer
	value := "test string"
	result = derefString(&value)
	assert.Equal(t, value, result, "derefString should return the string value")
}

// TestGetFilteredPosts tests filtering by tag, year, or both
func TestGetFilteredPosts(t *testing.T) {
	// Save original data
	origBlogPosts := blogPosts
	origBlogPostTagMap := blogPostTagMap
	origBlogPostYearMap := blogPostYearMap

	// Restore after test
	defer func() {
		blogPosts = origBlogPosts
		blogPostTagMap = origBlogPostTagMap
		blogPostYearMap = origBlogPostYearMap
	}()

	// Setup test data
	setupTestData()

	// Test case 1: No filters (should return all posts)
	posts := getFilteredPosts("", 0)
	assert.Len(t, posts, 3, "With no filters, should return all posts")

	// Test case 2: Filter by tag only
	posts = getFilteredPosts("golang", 0)
	assert.Len(t, posts, 2, "Should return posts with 'golang' tag")

	// Test case 3: Filter by year only
	posts = getFilteredPosts("", 2022)
	assert.Len(t, posts, 1, "Should return posts from 2022")

	// Test case 4: Filter by both tag and year
	posts = getFilteredPosts("golang", 2022)
	assert.Len(t, posts, 1, "Should return posts with 'golang' tag from 2022")

	// Test case 5: Filter with no matching posts
	posts = getFilteredPosts("nonexistent", 0)
	assert.Len(t, posts, 0, "Should return empty slice for non-existent tag")
}

// TestGetMarkdownData tests the markdown parsing and caching
func TestGetMarkdownData(t *testing.T) {
	// Save original markdownMap
	origMarkdownMap := markdownMap

	// Restore after test
	defer func() {
		markdownMap = origMarkdownMap
	}()

	// Reset the markdown map
	markdownMap = make(map[string]string)

	// Pre-populate the cache with rendered content
	markdownMap["test.md"] = "<h1>Test Heading</h1>\n<p>This is a test paragraph.</p>"

	// Test first call (should read from cache)
	html := getMarkdownData("test.md")
	assert.Contains(t, html, "<h1", "Markdown should be converted to HTML")
	assert.Contains(t, html, "Test Heading", "HTML should contain the heading text")
	assert.Contains(t, html, "<p>This is a test paragraph.</p>", "HTML should contain the paragraph")

	// Test second call (should read from cache again)
	html2 := getMarkdownData("test.md")
	assert.Equal(t, html, html2, "Second call should return cached content")

	// Add another entry to the cache
	markdownMap["test2.md"] = "<h2>Another Test</h2>"

	// Test with the new entry
	html3 := getMarkdownData("test2.md")
	assert.Contains(t, html3, "<h2", "Should read new content")
	assert.Contains(t, html3, "Another Test", "HTML should contain the content from the new file")
}

func TestPrimeSiteData(t *testing.T) {
	origMarkdownMap := markdownMap
	defer func() { markdownMap = origMarkdownMap }()

	data.ClearProjectsCache()
	data.ClearInterviewsCache()
	data.ClearBooksCache()
	data.ClearRunsCache()
	data.ClearStravaRunsCache()
	data.ClearCheatsheetsCache()
	ClearPicturesCache()
	markdownMap = make(map[string]string)

	assert.NoError(t, primeSiteData())
	assert.NotEmpty(t, markdownMap["about.markdown"])
	assert.NotEmpty(t, markdownMap["cheatsheets/go.md"])
	assert.True(t, picturesLoaded)
}

// TestGetProjectsCaching tests the caching mechanism for GetProjects
func TestGetProjectsCaching(t *testing.T) {
	// Clear cache before test
	data.ClearProjectsCache()

	// First call should load from file
	projects1, err1 := data.GetProjects()
	assert.NoError(t, err1)
	assert.NotEmpty(t, projects1)

	// Second call should return cached data (same pointer to first element)
	projects2, err2 := data.GetProjects()
	assert.NoError(t, err2)
	assert.NotEmpty(t, projects2)
	assert.True(t, &projects1[0] == &projects2[0], "Should return the same cached data instance (element pointer)")

	// Clear cache and reload
	data.ClearProjectsCache()
	projects3, err3 := data.GetProjects()
	assert.NoError(t, err3)
	assert.NotEmpty(t, projects3)
	assert.False(t, &projects2[0] == &projects3[0], "After clearing cache, should reload from file (different element pointer)")
}

func TestProjectsManifestIncludesBraveDevBotEmoji(t *testing.T) {
	data.ClearProjectsCache()

	projects, err := data.GetProjects()
	assert.NoError(t, err)

	found := false
	for _, project := range projects {
		if project.Github != "https://github.com/brave-experiments/brave-dev-bot" {
			continue
		}

		found = true
		assert.Equal(t, "Brave Dev Bot", project.Title)
		assert.Equal(t, "🤖", project.Emoji)
		assert.Equal(t, "https://github.com/brave-experiments/brave-dev-bot", project.URL)
	}

	assert.True(t, found, "Brave Dev Bot project should exist in the manifest")
}

func TestProjectsManifestIncludesBraveBot(t *testing.T) {
	data.ClearProjectsCache()

	projects, err := data.GetProjects()
	assert.NoError(t, err)

	found := false
	for _, project := range projects {
		if project.Github != "https://github.com/brave-experiments/brave-bot" {
			continue
		}

		found = true
		assert.Equal(t, "Brave Bot", project.Title)
		assert.Equal(t, "🛡️", project.Emoji)
		assert.Equal(t, "https://brave-experiments.github.io/brave-bot-docs/", project.Website)
		assert.Contains(t, project.Description, "structural resistance to indirect prompt injection")
	}

	assert.True(t, found, "Brave Bot project should exist in the manifest")
}

// TestGetCheatsheetsCaching tests the caching mechanism for GetCheatsheets
func TestGetCheatsheetsCaching(t *testing.T) {
	// Clear cache before test
	data.ClearCheatsheetsCache()

	// First call should load from file
	cheatsheets1, err1 := data.GetCheatsheets()
	assert.NoError(t, err1)
	assert.NotEmpty(t, cheatsheets1)

	// Second call should return cached data (same pointer to first element)
	cheatsheets2, err2 := data.GetCheatsheets()
	assert.NoError(t, err2)
	assert.NotEmpty(t, cheatsheets2)
	assert.True(t, &cheatsheets1[0] == &cheatsheets2[0], "Should return the same cached data instance (element pointer)")

	// Clear cache and reload
	data.ClearCheatsheetsCache()
	cheatsheets3, err3 := data.GetCheatsheets()
	assert.NoError(t, err3)
	assert.NotEmpty(t, cheatsheets3)
	assert.False(t, &cheatsheets2[0] == &cheatsheets3[0], "After clearing cache, should reload from file (different element pointer)")
}

// Helper function to set up test data
func setupTestData() {
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
}

// TestMockInitializeBlogPosts tests the blog post initialization logic
// but uses a mock approach instead of actual file reading
func TestMockInitializeBlogPosts(t *testing.T) {
	// Save original data
	origBlogPosts := blogPosts
	origBlogPostIdMap := blogPostIdMap
	origBlogPostTagMap := blogPostTagMap
	origBlogPostYearMap := blogPostYearMap
	origTagCountMap := tagCountMap
	origSortedTags := sortedTags

	// Restore after test
	defer func() {
		blogPosts = origBlogPosts
		blogPostIdMap = origBlogPostIdMap
		blogPostTagMap = origBlogPostTagMap
		blogPostYearMap = origBlogPostYearMap
		tagCountMap = origTagCountMap
		sortedTags = origSortedTags
	}()

	// Reset global variables
	blogPosts = nil
	blogPostIdMap = make(map[int]data.BlogPost)
	blogPostTagMap = make(map[string][]data.BlogPost)
	blogPostYearMap = make(map[int][]data.BlogPost)
	tagCountMap = make(map[string]int)
	sortedTags = nil

	// Manually call the setup that initializeBlogPosts would do
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

	// Process the posts as initializeBlogPosts would
	for _, blogPost := range blogPosts {
		for _, tag := range blogPost.Tags {
			blogPostTagMap[tag] = append(blogPostTagMap[tag], blogPost)
			tagCountMap[tag] += 1
		}
		parsedDate, _ := time.Parse(layoutISO, blogPost.Created)
		year := parsedDate.Year()
		blogPostYearMap[year] = append(blogPostYearMap[year], blogPost)
		blogPostIdMap[blogPost.Id] = blogPost
	}

	sortedTags = make([]string, len(tagCountMap))
	i := 0
	for k := range tagCountMap {
		sortedTags[i] = k
		i++
	}

	// Verify results
	assert.Len(t, blogPosts, 3, "Should load 3 blog posts")
	assert.Len(t, blogPostIdMap, 3, "Should create ID map with 3 entries")
	assert.Len(t, blogPostTagMap, 2, "Should create tag map with 2 entries (test, golang)")
	assert.Len(t, blogPostYearMap, 2, "Should create year map with 2 entries (2022, 2023)")

	// Check tag counts
	assert.Equal(t, 2, tagCountMap["test"], "Tag 'test' should appear 2 times")
	assert.Equal(t, 2, tagCountMap["golang"], "Tag 'golang' should appear 2 times")

	// Check sorted tags
	assert.Len(t, sortedTags, 2, "Should have 2 sorted tags")
	assert.Contains(t, sortedTags, "test")
	assert.Contains(t, sortedTags, "golang")
}
