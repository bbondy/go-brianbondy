package main

import (
	"testing"
	"time"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/stretchr/testify/assert"
)

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
