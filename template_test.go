package main

import (
	"html/template"
	"testing"

	"github.com/bbondy/go-brianbondy/data" // Import the data package where BlogPost is defined
)

// Mock data structures for testing
type TestPost struct {
	Title   string
	Content string
	Empty   string
}

func TestAvail(t *testing.T) {
	// Test with struct
	post := TestPost{
		Title:   "Test Title",
		Content: "Test Content",
		Empty:   "",
	}

	// Test valid non-empty field
	if !avail("Title", post) {
		t.Error("avail should return true for Title field")
	}

	// Test valid non-empty field with pointer
	if !avail("Title", &post) {
		t.Error("avail should return true for Title field with pointer")
	}

	// Test valid empty string field
	if avail("Empty", post) {
		t.Error("avail should return false for empty string field")
	}

	// Test non-existent field
	if avail("NonExistent", post) {
		t.Error("avail should return false for non-existent field")
	}

	// Test with non-struct
	if avail("Title", "string") {
		t.Error("avail should return false with non-struct")
	}

	// Test with nil
	if avail("Title", nil) {
		t.Error("avail should return false with nil")
	}
}

func TestHTMLSafe(t *testing.T) {
	htmlString := "<p>Test</p>"
	htmlSafeFunc := funcMap["htmlSafe"].(func(string) template.HTML)

	result := htmlSafeFunc(htmlString)
	expected := template.HTML(htmlString)

	if result != expected {
		t.Errorf("htmlSafe should return template.HTML, got: %v, want: %v", result, expected)
	}
}

func TestGetTagCount(t *testing.T) {
	// Setup test data
	oldTagCountMap := tagCountMap
	tagCountMap = map[string]int{
		"golang": 5,
		"test":   2,
	}
	defer func() { tagCountMap = oldTagCountMap }() // Restore after test

	getTagCountFunc := funcMap["getTagCount"].(func(string) int)

	// Test existing tag
	if count := getTagCountFunc("golang"); count != 5 {
		t.Errorf("getTagCount for 'golang' should return 5, got: %d", count)
	}

	// Test another existing tag
	if count := getTagCountFunc("test"); count != 2 {
		t.Errorf("getTagCount for 'test' should return 2, got: %d", count)
	}

	// Test non-existent tag
	if count := getTagCountFunc("nonexistent"); count != 0 {
		t.Errorf("getTagCount for non-existent tag should return 0, got: %d", count)
	}
}

func TestTruncateTitle(t *testing.T) {
	truncateFunc := funcMap["truncateTitle"].(func(string) string)

	tests := []struct {
		input    string
		expected string
	}{
		{"Short title", "Short title"},
		{"This is exactly four words", "This is exactly four..."},
		{"This is a longer title with many words", "This is a longer..."},
		{"This, is. a! title? with: punctuation;", "This, is. a! title..."},
	}

	for _, test := range tests {
		result := truncateFunc(test.input)
		if result != test.expected {
			t.Errorf("truncateTitle(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestTagUrl(t *testing.T) {
	tagUrlFunc := funcMap["tagUrl"].(func(string) string)

	tests := []struct {
		input    string
		expected string
	}{
		{"golang", "/all?tag=golang"},
		{"web development", "/all?tag=web+development"},
		{"c++", "/all?tag=c%2B%2B"},
	}

	for _, test := range tests {
		result := tagUrlFunc(test.input)
		if result != test.expected {
			t.Errorf("tagUrl(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestYearUrl(t *testing.T) {
	yearUrlFunc := funcMap["yearUrl"].(func(int) string)

	tests := []struct {
		input    int
		expected string
	}{
		{2020, "/all?year=2020"},
		{2021, "/all?year=2021"},
		{1999, "/all?year=1999"},
	}

	for _, test := range tests {
		result := yearUrlFunc(test.input)
		if result != test.expected {
			t.Errorf("yearUrl(%d) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestGetYearCount(t *testing.T) {
	// Setup test data
	oldBlogPostYearMap := blogPostYearMap

	// Create a mock map with the correct type
	blogPostYearMap = make(map[int][]data.BlogPost)
	blogPostYearMap[2020] = make([]data.BlogPost, 2) // 2 posts
	blogPostYearMap[2021] = make([]data.BlogPost, 3) // 3 posts

	defer func() { blogPostYearMap = oldBlogPostYearMap }() // Restore after test

	getYearCountFunc := funcMap["getYearCount"].(func(int) int)

	// Test existing year
	if count := getYearCountFunc(2020); count != 2 {
		t.Errorf("getYearCount for 2020 should return 2, got: %d", count)
	}

	// Test another existing year
	if count := getYearCountFunc(2021); count != 3 {
		t.Errorf("getYearCount for 2021 should return 3, got: %d", count)
	}

	// Test non-existent year
	if count := getYearCountFunc(1999); count != 0 {
		t.Errorf("getYearCount for non-existent year should return 0, got: %d", count)
	}
}
