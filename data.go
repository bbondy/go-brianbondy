package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

var markdownMap = make(map[string]string)
var markdownMapMutex sync.RWMutex
var blogPostTagMap = make(map[string][]data.BlogPost)
var blogPostYearMap = make(map[int][]data.BlogPost)
var blogPosts []data.BlogPost
var blogPostIdMap = make(map[int]data.BlogPost)
var tagCountMap = make(map[string]int)
var sortedTags []string

// tagGroupDefinitions controls the category order on the filters page. Tags
// not listed here remain visible in the final "Other" group.
var tagGroupDefinitions = []data.TagGroup{
	{Name: "Running", Tags: []string{"running", "byu", "cocodona", "destination-trails", "hurt100", "tahoe"}},
	{Name: "Programming", Tags: []string{"ai", "c", "c++", "c++0x", "c-sharp", "data-structure", "development", "django", "electron", "g++", "javascript", "node", "nodejs", "objective-c", "programming", "python", "stl"}},
	{Name: "Browsers", Tags: []string{"brave", "codefirefox", "firebug", "firefox", "firefox-ios", "firefox-os", "gecko", "mozilla", "snappy", "thunderbird", "xul", "xpcom"}},
	{Name: "Web Development", Tags: []string{"akismet", "analytics", "captcha", "css", "dataurl", "google-app-engine", "html", "html5", "http", "nofollow", "seo", "site", "svg", "urlzip", "web", "xhtml"}},
	{Name: "Platforms & Systems", Tags: []string{"android", "apple", "cocoa", "directx", "ia64", "imac", "internet-explorer", "linux", "mac", "mac-book-pro", "microsoft", "mobile", "uac", "ubuntu", "visual-studio", "win7", "winapi", "windows", "windows-8", "windows8", "wp7", "x64", "xaml", "xcode"}},
	{Name: "Networking & Infrastructure", Tags: []string{"backup", "compression", "ddns", "email", "encoding", "ftp", "imap", "ipfs", "nat", "networking", "pipelining", "pop3", "protocol", "silent-update", "smtp", "spam", "tcp", "udp"}},
	{Name: "Learning & Data", Tags: []string{"book", "data-analysis", "english", "khan-academy", "math", "oreilly", "puzzle", "review", "wikipedia"}},
	{Name: "Career & Industry", Tags: []string{"business", "career", "google", "hackathon", "mentor", "mvp", "stackoverflow", "stackexchange", "talks", "tech", "twitter", "yahoo"}},
	{Name: "Personal", Tags: []string{"family", "gaming", "life", "personal"}},
	{Name: "Site Updates", Tags: []string{"how-stuff-works", "link-bubble", "pref", "status-report"}},
}

var (
	// This permits declaration lists used by existing posts while excluding
	// functions, quoted values, and URL syntax, which can make CSS executable.
	safeInlineStyle = regexp.MustCompile(`(?i)^\s*[a-z-]+\s*:\s*[-#a-z0-9.%\s]+(?:;\s*[a-z-]+\s*:\s*[-#a-z0-9.%\s]+)*;?\s*$`)
	iframeSource    = regexp.MustCompile(`^https://www\.youtube\.com/embed/[A-Za-z0-9_-]+(?:\?[-A-Za-z0-9_=&%./]+)?$|^http://s\.vid\.ly/embeded\.html\?link=[A-Za-z0-9]+&autoplay=(?:true|false)$`)
	integerValue    = regexp.MustCompile(`^[0-9]+$`)
	iframeAllow     = regexp.MustCompile(`^[a-z-]+(?:;\s*[a-z-]+)*$`)
	markdownPolicy  = newMarkdownPolicy()
)

func newMarkdownPolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()

	// Existing content uses classes and a small set of presentational styles.
	// Keep those styles while restricting both property names and values.
	p.AllowAttrs("class").Globally()
	p.AllowAttrs("style").Matching(safeInlineStyle).Globally()

	// Preserve the existing, explicitly trusted media embeds without allowing
	// arbitrary third-party documents to be embedded in the site.
	p.AllowElements("iframe")
	p.AllowAttrs("src").Matching(iframeSource).OnElements("iframe")
	p.AllowAttrs("width", "height", "frameborder").Matching(integerValue).OnElements("iframe")
	p.AllowAttrs("title", "name", "scrolling").OnElements("iframe")
	p.AllowAttrs("allow").Matching(iframeAllow).OnElements("iframe")
	p.AllowAttrs("allowfullscreen").OnElements("iframe")

	// Preserve local HTML5 video without allowing remote media URLs.
	p.AllowElements("video", "source")
	p.AllowAttrs("controls").OnElements("video")
	p.AllowAttrs("src").Matching(regexp.MustCompile(`^/static/[A-Za-z0-9_./-]+$`)).OnElements("video", "source")
	p.AllowAttrs("type").Matching(regexp.MustCompile(`^video/[A-Za-z0-9.+-]+$`)).OnElements("source")

	return p
}

func sanitizeMarkdownHTML(html string) string {
	return markdownPolicy.Sanitize(html)
}

func renderMarkdown(content []byte) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	markdownParser := parser.NewWithExtensions(extensions)
	return string(markdown.ToHTML(content, markdownParser, nil))
}

func initializeBlogPosts() {
	if err := initializeBlogPostTranslations(); err != nil {
		panic(err)
	}
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

func buildTagGroups(counts map[string]int) []data.TagGroup {
	groups := make([]data.TagGroup, 0, len(tagGroupDefinitions)+1)
	assigned := make(map[string]bool)
	for _, definition := range tagGroupDefinitions {
		group := data.TagGroup{Name: definition.Name}
		for _, tag := range definition.Tags {
			if counts[tag] > 0 {
				group.Tags = append(group.Tags, tag)
				assigned[tag] = true
			}
		}
		if len(group.Tags) > 0 {
			groups = append(groups, group)
		}
	}

	other := data.TagGroup{Name: "Other"}
	for tag := range counts {
		if !assigned[tag] {
			other.Tags = append(other.Tags, tag)
		}
	}
	sort.Slice(other.Tags, func(i, j int) bool {
		if counts[other.Tags[i]] == counts[other.Tags[j]] {
			return other.Tags[i] < other.Tags[j]
		}
		return counts[other.Tags[i]] > counts[other.Tags[j]]
	})
	if len(other.Tags) > 0 {
		groups = append(groups, other)
	}
	return groups
}

// primeSiteData loads all handler-backed data before the HTTP server accepts
// requests. The caches are then read-only for the lifetime of the process.
func primeSiteData() error {
	if _, err := data.GetProjects(); err != nil {
		return fmt.Errorf("load projects: %w", err)
	}
	if _, err := data.GetInterviews(); err != nil {
		return fmt.Errorf("load interviews: %w", err)
	}
	if _, err := data.GetBooks(); err != nil {
		return fmt.Errorf("load books: %w", err)
	}
	if _, err := data.GetRuns(); err != nil {
		return fmt.Errorf("load runs: %w", err)
	}
	if _, err := data.GetStravaRuns(); err != nil {
		return fmt.Errorf("load Strava runs: %w", err)
	}
	if _, err := getCachedPictures(); err != nil {
		return fmt.Errorf("load pictures: %w", err)
	}

	markdownSlugs := []string{
		"about.markdown",
		"advice.markdown",
		"contact.markdown",
	}
	cheatsheets, err := data.GetCheatsheets()
	if err != nil {
		return fmt.Errorf("load cheatsheets: %w", err)
	}
	for _, cheatsheet := range cheatsheets {
		markdownSlugs = append(markdownSlugs, "cheatsheets/"+cheatsheet.Slug+".md")
	}

	for _, slug := range markdownSlugs {
		if err := primeMarkdownData(slug); err != nil {
			return err
		}
	}
	return nil
}

func primeMarkdownData(slug string) error {
	if _, err := os.Stat("data/markdown/" + slug); err != nil {
		return fmt.Errorf("load markdown %q: %w", slug, err)
	}
	getMarkdownData(slug)
	return nil
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
	markdownMapMutex.RLock()
	content, ok := markdownMap[slug]
	markdownMapMutex.RUnlock()
	if ok {
		return content
	}

	markdownData, _ := ioutil.ReadFile("data/markdown/" + slug)
	content = sanitizeMarkdownHTML(renderMarkdown(markdownData))
	markdownMapMutex.Lock()
	if cached, exists := markdownMap[slug]; exists {
		content = cached
	} else {
		markdownMap[slug] = content
	}
	markdownMapMutex.Unlock()
	return content
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
