package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/codegangsta/negroni"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

const (
	layoutISO = "2006-01-02"
	layoutUS  = "January 2, 2006"
)

func directToHttps(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Host == "localhost:8080" ||
		r.URL.Scheme == "https" ||
		strings.HasPrefix(r.Proto, "HTTPS") ||
		r.Header.Get("X-Forwarded-Proto") == "https" {
		next(w, r)
	} else {
		target := "https://" + r.Host + r.URL.Path
		http.Redirect(w, r, target,
			http.StatusTemporaryRedirect)
	}
}

func avail(name string, data interface{}) bool {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	field := v.FieldByName(name)
	if !field.IsValid() {
		return false
	}

	// Check if the field is a string and not empty
	if field.Kind() == reflect.String {
		return field.String() != ""
	}
	// Return true if the field is not a string but is valid
	return true
}

var markdownMap = make(map[string]string)
var blogPostTagMap = make(map[string][]data.BlogPost)
var blogPostYearMap = make(map[int][]data.BlogPost)
var blogPosts []data.BlogPost
var blogPostIdMap = make(map[int]data.BlogPost)
var tagCountMap = make(map[string]int)
var sortedTags []string

var funcMap = template.FuncMap{
	"avail": avail,
	"htmlSafe": func(html string) template.HTML {
		return template.HTML(html)
	},
	"getTagCount": func(tag string) int {
		count, ok := tagCountMap[tag]
		if !ok {
			return 0
		}
		return count
	},
	"slugifyTitle": slugifyTitle,
	"truncateTitle": func(title string) string {
		words := strings.Fields(title)
		if len(words) <= 4 {
			return title
		}
		// Take first 4 words
		truncated := strings.Join(words[:4], " ")
		// Remove any trailing punctuation
		truncated = strings.TrimRight(truncated, ".,!?:;")
		return truncated + "..."
	},
	"tagUrl": func(tag string) string {
		return fmt.Sprintf("/all?tag=%s", url.QueryEscape(tag))
	},
	"yearUrl": func(year int) string {
		return fmt.Sprintf("/all?year=%d", year)
	},
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
	http.Redirect(w, r, path, 302)
}

func errorPage(w http.ResponseWriter, message string, slug string) {
	w.WriteHeader(http.StatusNotFound)
	p := &data.SimpleMarkdownPage{
		Title:        "Error",
		Content:      message,
		MarkdownSlug: slug,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/simpleMarkdown.html"))
	t.Execute(w, p)
	// fmt.Fprint(w, message)
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
		year, _ = strconv.Atoi(yearStr)
	} else if yearStr, ok := vars["year"]; ok {
		year, _ = strconv.Atoi(yearStr)
	}

	filteredBlogPosts := getFilteredPosts(tag, year)

	// Handle individual blog post view
	if idStr, ok := vars["id"]; ok {
		id, _ := strconv.Atoi(idStr)
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
			t.Execute(w, p)
			return
		}
		errorPage(w, "No blog posts for this query", "blog")
		return
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
			NextPost:     nextPost,
			Tag:          tag,
			Year:         year,
			ImagePath:    derefString(post.ImagePath),
			Description:  derefString(post.Description),
			MarkdownSlug: "blog",
		}
		t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/blogPost.html"))
		t.Execute(w, p)
		return
	}

	errorPage(w, "No blog posts found", "blog")
}

func getImageMimeType(imagePath string) string {
	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	default:
		return ""
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
	w.Write([]byte(rss))
}

func filtersPageHandler(w http.ResponseWriter, r *http.Request) {
	current_year := time.Now().Year()
	year_range := make([]int, 10)
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
	t.Execute(w, p)
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
		t.Execute(w, p)
	}
	return negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(handler)))
}

func initializeBlogPosts() {
	blogPostManifest, _ := ioutil.ReadFile("data/blogPostManifest.json")
	err := json.Unmarshal([]byte(blogPostManifest), &blogPosts)
	if err != nil {
		panic(fmt.Errorf("Error parsing JSON"))
	}

	for _, blogPost := range blogPosts {
		for _, tag := range blogPost.Tags {
			blogPostTagMap[tag] = append(blogPostTagMap[tag], blogPost)
			if _, ok := tagCountMap[tag]; ok {
				tagCountMap[tag] += 1
			} else {
				tagCountMap[tag] = 1
			}
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
	sort.SliceStable(sortedTags, func(i, j int) bool {
		tag1 := sortedTags[i]
		tag2 := sortedTags[j]
		return tagCountMap[tag1] > tagCountMap[tag2]
	})
}

func initializeRoutes(router *mux.Router) {
	fs := http.FileServer(http.Dir("static/"))
	s := http.StripPrefix("/static/", fs)
	router.PathPrefix("/static/").Handler(s)

	handleBlogPost := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(blogPostPageHandler)))
	handleRedirect := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(redirectHandler)))
	handleFilterPage := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(filtersPageHandler)))
	handleRSS := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(generateRSSHandler)))
	handleTagRedirect := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(tagRedirectHandler)))
	handlePaginationRedirect := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(paginationRedirectHandler)))
	handleBlogIdRedirect := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(blogIdRedirectHandler)))
	handleYearRedirect := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(yearRedirectHandler)))
	handleHome := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(homePageHandler)))
	handleAllPosts := negroni.New(
		negroni.HandlerFunc(directToHttps),
		negroni.Wrap(http.HandlerFunc(allPostsHandler)))

	router.Handle("/", handleHome)
	router.Handle("/rss", handleRSS)
	router.Handle("/blog/{id:[0-9]+}", handleBlogIdRedirect)
	router.Handle("/blog/{id:[0-9]+}/{slug}", handleBlogPost)
	router.Handle("/page/{page:[0-9]+}", handlePaginationRedirect)
	router.Handle("/tagged/{tag}", handleTagRedirect)
	router.Handle("/posted/{year:[0-9]+}", handleYearRedirect)
	router.Handle("/posted/{year:[0-9]+}/page/{page:[0-9]+}", handlePaginationRedirect)
	router.Handle("/blog/page/{page}", handleRedirect)
	router.Handle("/blog/tagged/{tag}", handleRedirect)
	router.Handle("/blog/tagged/{tag}/page/{page}", handleRedirect)
	router.Handle("/blog/posted/{year:[0-9]+}", handleRedirect)
	router.Handle("/blog/posted/{year:[0-9]+}/page/{page:[0-9]+}", handleRedirect)
	router.Handle("/blog/filters", handleFilterPage)
	router.Handle("/about", getMarkdownTemplateHandler("About", "about.markdown", "/about"))
	router.Handle("/other", getMarkdownTemplateHandler("Other", "other.markdown", "/other"))
	router.Handle("/contact", getMarkdownTemplateHandler("Contact", "contact.markdown", "/contact"))
	router.Handle("/projects", getMarkdownTemplateHandler("Projects", "projects.markdown", "/projects"))
	router.Handle("/advice", getMarkdownTemplateHandler("Advice", "advice.markdown", "/advice"))
	router.Handle("/books", getMarkdownTemplateHandler("Books", "books.markdown", "/books"))
	router.Handle("/resume", getMarkdownTemplateHandler("Resume", "resume.markdown", "/resume"))
	router.Handle("/running", getMarkdownTemplateHandler("Running", "running.markdown", "/running"))
	router.Handle("/all", handleAllPosts)
}

func main() {
	initializeBlogPosts()
	router := mux.NewRouter()
	initializeRoutes(router)
	http.ListenAndServe(":8080", router)
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func slugifyTitle(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace any non-alphanumeric characters (except hyphens) with a hyphen
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
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

func paginationRedirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, _ := strconv.Atoi(vars["page"])

	// Get tag and year from query parameters
	tag := r.URL.Query().Get("tag")
	year := 0
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		year, _ = strconv.Atoi(yearStr)
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
	id, _ := strconv.Atoi(vars["id"])

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
	year, _ := strconv.Atoi(vars["year"])

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
			BlogPost: post,
			Preview:  template.HTML(preview),
			PostDate: parsedDate.Format(layoutUS),
			PostUrl:  fmt.Sprintf("/blog/%d/%s", post.Id, slugifyTitle(post.Title)),
		})
	}

	// Get all posts for the list
	allPosts := make([]data.BlogPostPreview, 0, len(blogPosts))
	for _, post := range blogPosts {
		parsedDate, _ := time.Parse(layoutISO, post.Created)
		allPosts = append(allPosts, data.BlogPostPreview{
			BlogPost: post,
			PostDate: parsedDate.Format(layoutUS),
			PostUrl:  fmt.Sprintf("/blog/%d/%s", post.Id, slugifyTitle(post.Title)),
		})
	}

	p := &data.HomePage{
		Title:        "Brian R. Bondy",
		Posts:        previewPosts,
		AllPosts:     allPosts,
		MarkdownSlug: "home",
	}

	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/home.html"))
	t.Execute(w, p)
}

func extractFirstParagraph(content string) string {
	// First try to find content between first <p> tags
	re := regexp.MustCompile(`<p>(.*?)</p>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		// Get the content inside the first <p> tags
		preview := matches[1]

		// Remove any remaining HTML tags
		tagRegex := regexp.MustCompile(`<[^>]*>`)
		preview = tagRegex.ReplaceAllString(preview, "")

		if preview != "" {
			// If preview is too long, truncate it
			if len(preview) > 300 {
				// Try to cut at a word boundary
				lastSpace := strings.LastIndex(preview[:300], " ")
				if lastSpace > 0 {
					preview = preview[:lastSpace] + "..."
				} else {
					preview = preview[:300] + "..."
				}
			}
			return preview
		}
	}

	// If no paragraph found or it was empty, remove all HTML tags from content
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	cleanContent := tagRegex.ReplaceAllString(content, " ")

	// Remove extra whitespace
	cleanContent = strings.Join(strings.Fields(cleanContent), " ")

	// Take first 300 characters or less
	if len(cleanContent) > 300 {
		lastSpace := strings.LastIndex(cleanContent[:300], " ")
		if lastSpace > 0 {
			cleanContent = cleanContent[:lastSpace] + "..."
		} else {
			cleanContent = cleanContent[:300] + "..."
		}
	}

	return cleanContent
}

func allPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Get tag and year from query parameters
	tag := r.URL.Query().Get("tag")
	year := 0
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		year, _ = strconv.Atoi(yearStr)
	}

	// Get filtered posts based on tag and/or year
	filteredPosts := getFilteredPosts(tag, year)

	allPosts := make([]data.BlogPostPreview, 0, len(filteredPosts))
	for _, post := range filteredPosts {
		parsedDate, _ := time.Parse(layoutISO, post.Created)
		allPosts = append(allPosts, data.BlogPostPreview{
			BlogPost: post,
			PostDate: parsedDate.Format(layoutUS),
			PostUrl:  fmt.Sprintf("/blog/%d/%s", post.Id, slugifyTitle(post.Title)),
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
	t.Execute(w, p)
}
