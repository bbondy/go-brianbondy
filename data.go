package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

var markdownMap = make(map[string]string)
var blogPostTagMap = make(map[string][]data.BlogPost)
var blogPostYearMap = make(map[int][]data.BlogPost)
var blogPosts []data.BlogPost
var blogPostIdMap = make(map[int]data.BlogPost)
var tagCountMap = make(map[string]int)
var sortedTags []string

func initializeBlogPosts() {
	blogPostManifest, _ := ioutil.ReadFile("data/blogPostManifest.json")
	err := json.Unmarshal([]byte(blogPostManifest), &blogPosts)
	if err != nil {
		panic(fmt.Errorf("error parsing JSON"))
	}

	for i := range blogPosts {
		blogPost := &blogPosts[i]

		// Calculate reading time for each blog post
		content := getMarkdownData("blog/" + fmt.Sprintf("%d", blogPost.Id) + ".markdown")
		blogPost.ReadingTime = calculateReadingTime(content)

		for _, tag := range blogPost.Tags {
			blogPostTagMap[tag] = append(blogPostTagMap[tag], *blogPost)
			tagCountMap[tag] += 1
		}
		parsedDate, _ := time.Parse(layoutISO, blogPost.Created)
		year := parsedDate.Year()
		blogPostYearMap[year] = append(blogPostYearMap[year], *blogPost)
		blogPostIdMap[blogPost.Id] = *blogPost
	}
	sortedTags = make([]string, len(tagCountMap))
	i := 0
	for k := range tagCountMap {
		sortedTags[i] = k
		i++
	}
	sort.SliceStable(sortedTags, func(i, j int) bool {
		tag1 := sortedTags[i]
		tag2 := sortedTags[j]
		return tagCountMap[tag1] > tagCountMap[tag2]
	})
}

func getFilteredPosts(tag string, year int) []data.BlogPost {
	var filtered []data.BlogPost

	if tag != "" && year != 0 {
		// Filter by both tag and year
		for _, post := range blogPosts {
			parsedDate, _ := time.Parse(layoutISO, post.Created)
			if parsedDate.Year() == year {
				// Check if post has the specified tag
				for _, postTag := range post.Tags {
					if postTag == tag {
						filtered = append(filtered, post)
						break
					}
				}
			}
		}
		return filtered
	}

	// Single filter cases
	if tag != "" {
		return blogPostTagMap[tag]
	}
	if year != 0 {
		return blogPostYearMap[year]
	}
	return blogPosts
}

func getMarkdownData(slug string) string {
	_, ok := markdownMap[slug]
	if !ok {
		data, _ := ioutil.ReadFile("data/markdown/" + slug)
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		parser := parser.NewWithExtensions(extensions)
		html := markdown.ToHTML([]byte(data), parser, nil)
		markdownMap[slug] = string(html)
	}
	return markdownMap[slug]
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// calculateReadingTime calculates the estimated reading time in minutes
// Uses average reading speed of 300 words per minute
// Adds 30 seconds per video and 5 seconds per image
func calculateReadingTime(content string) int {
	// Remove HTML tags to get plain text
	plainText := stripHTMLTags(content)

	// Count words (split by whitespace)
	words := len(strings.Fields(plainText))

	// Calculate reading time (300 WPM average)
	readingTime := words / 300

	// Count videos (looking for video tags or video files)
	videoCount := countVideos(content)
	videoTime := (videoCount * 30) / 60 // 30 seconds per video, convert to minutes
	readingTime += videoTime

	// Count images (looking for img tags)
	imageCount := countImages(content)
	imageTime := (imageCount * 5) / 60 // 5 seconds per image, convert to minutes
	readingTime += imageTime

	// Ensure minimum reading time of 1 minute
	if readingTime < 1 {
		readingTime = 1
	}

	return readingTime
}

// stripHTMLTags removes HTML tags from content to get plain text
func stripHTMLTags(html string) string {
	// Simple regex to remove HTML tags
	// This is a basic implementation - for production, consider using a proper HTML parser
	var result strings.Builder
	var inTag bool

	for _, char := range html {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// countVideos counts the number of videos in HTML content
func countVideos(content string) int {
	count := 0

	// Count video tags
	count += strings.Count(content, "<video")

	// Count iframe tags (for embedded videos)
	count += strings.Count(content, "<iframe")

	// Count video file extensions in src attributes
	videoExtensions := []string{".mp4", ".webm", ".ogg", ".mov", ".avi"}
	for _, ext := range videoExtensions {
		count += strings.Count(content, ext)
	}

	return count
}

// countImages counts the number of images in HTML content
func countImages(content string) int {
	// Count img tags
	return strings.Count(content, "<img")
}
