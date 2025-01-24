package main

import (
	"encoding/json"
	"fmt"
	"github.com/bbondy/go-brianbondy/data"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
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
	return v.FieldByName(name).IsValid()
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
	blogPostIndex := 0 //page
	vars := mux.Vars(r)

	if page, ok := vars["page"]; ok {
		blogPostIndex, _ = strconv.Atoi(page)
		blogPostIndex -= 1
	}

	filteredBlogPosts := blogPosts
	year := 0

	if yearStr, ok := vars["year"]; ok {
		year, _ = strconv.Atoi(yearStr)
		filteredBlogPosts = blogPostYearMap[year]
	}

	tag, tagOk := vars["tag"]
	if tagOk {
		filteredBlogPosts = blogPostTagMap[tag]
	}

	if idStr, ok := vars["id"]; ok {
		id, _ := strconv.Atoi(idStr)
		if foundPost, ok := blogPostIdMap[id]; ok {
			filteredBlogPosts = []data.BlogPost{foundPost}
		} else {
			errorPage(w, "No blog posts for this query", "blog")
		}
	}

	if blogPostIndex >= len(filteredBlogPosts) || blogPostIndex < 0 {
		errorPage(w, "No blog posts for this query", "blog")
		return
	}

	parsedDate, _ := time.Parse(layoutISO, filteredBlogPosts[blogPostIndex].Created)

	var fbImagePath string
	if filteredBlogPosts[blogPostIndex].FBImagePath != nil {
		fbImagePath = *filteredBlogPosts[blogPostIndex].FBImagePath
	}

	var fbDescription string
	if filteredBlogPosts[blogPostIndex].FBDescription != nil {
		fbDescription = *filteredBlogPosts[blogPostIndex].FBDescription
	}

	blogPostUri :=  "/blog/" + strconv.Itoa(filteredBlogPosts[blogPostIndex].Id) + "/" + SlugifyTitle(filteredBlogPosts[blogPostIndex].Title)

	p := &data.BlogPostPage{
		Title:        GetTitle("Blog posts"),
		BlogPost:     filteredBlogPosts[blogPostIndex],
		BlogPostBody: getMarkdownData("blog/" + strconv.Itoa(filteredBlogPosts[blogPostIndex].Id) + ".markdown"),
		BlogPostUri:  blogPostUri,
		BlogPostDate: parsedDate.Format(layoutUS),
		NextPage:     blogPostIndex + 2,
		PrevPage:     blogPostIndex,
		MaxPage:      len(filteredBlogPosts),
		Tag:          tag,
		Year:         year,
		FBImagePath:  fbImagePath,
		FBDescription: fbDescription,
		FBShareUrl: blogPostUri,
		MarkdownSlug: "blog",
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/blogPost.html"))
	t.Execute(w, p)
}

func generateRSSHandler(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host
	rssXML, err := data.ConvertToRSS(blogPosts, GetTitle("Blog posts"), target)
	if err != nil {
		http.Error(w, "Error generating RSS feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(rssXML)
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

func getMarkdownTemplateHandler(titleSlug string, markdownSlug string) *negroni.Negroni {
	handler := func(w http.ResponseWriter, r *http.Request) {
		p := &data.SimpleMarkdownPage{
			Title:        GetTitle(titleSlug),
			Content:      getMarkdownData(markdownSlug),
			MarkdownSlug: markdownSlug,
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

	router.Handle("/", handleBlogPost)
	router.Handle("/rss", handleRSS)
	router.Handle("/blog/{id:[0-9]+}", handleBlogPost)
	router.Handle("/blog/{id:[0-9]+}/{slug}", handleBlogPost)
	router.Handle("/page/{page:[0-9]+}", handleBlogPost)
	router.Handle("/tagged/{tag}", handleBlogPost)
	router.Handle("/tagged/{tag}/page/{page:[0-9]+}", handleBlogPost)
	router.Handle("/posted/{year:[0-9]+}", handleBlogPost)
	router.Handle("/posted/{year:[0-9]+}/page/{page:[0-9]+}", handleBlogPost)
	router.Handle("/blog/page/{page}", handleRedirect)
	router.Handle("/blog/tagged/{tag}", handleRedirect)
	router.Handle("/blog/tagged/{tag}/page/{page}", handleRedirect)
	router.Handle("/blog/posted/{year:[0-9]+}", handleRedirect)
	router.Handle("/blog/posted/{year:[0-9]+}/page/{page:[0-9]+}", handleRedirect)
	router.Handle("/blog/filters", handleFilterPage)
	router.Handle("/about", getMarkdownTemplateHandler("About", "about.markdown"))
	router.Handle("/other", getMarkdownTemplateHandler("Other", "other.markdown"))
	router.Handle("/contact", getMarkdownTemplateHandler("Contact", "contact.markdown"))
	router.Handle("/projects", getMarkdownTemplateHandler("Projects", "projects.markdown"))
	router.Handle("/advice", getMarkdownTemplateHandler("Advice", "advice.markdown"))
	router.Handle("/books", getMarkdownTemplateHandler("Books", "books.markdown"))
	router.Handle("/braille", getMarkdownTemplateHandler("Braille", "braille.markdown"))
	router.Handle("/compression", getMarkdownTemplateHandler("Compression", "compression.markdown"))
	router.Handle("/compression/huffman", getMarkdownTemplateHandler("Huffman Compression", "compression/huffman.markdown"))
	router.Handle("/compression/BWT", getMarkdownTemplateHandler("Burrows-Wheeler", "compression/BWT.markdown"))
	router.Handle("/compression/PPM", getMarkdownTemplateHandler("Burrows-Wheeler", "compression/PPM.markdown"))
	router.Handle("/math", getMarkdownTemplateHandler("Mathematics", "math.markdown"))
	router.Handle("/math/main", getMarkdownTemplateHandler("Main", "math/main.markdown"))
	router.Handle("/math/pi", getMarkdownTemplateHandler("Pi", "math/pi.markdown"))
	router.Handle("/math/primes", getMarkdownTemplateHandler("Primes", "math/primes.markdown"))
	router.Handle("/math/numberTheory", getMarkdownTemplateHandler("Mathematics", "math/numberTheory.markdown"))
	router.Handle("/math/graphTheory", getMarkdownTemplateHandler("Mathematics", "math/graphTheory.markdown"))
	router.Handle("/math/mathTricks", getMarkdownTemplateHandler("Mathematics", "math/mathTricks.markdown"))
	router.Handle("/morseCode", getMarkdownTemplateHandler("Morse Code", "morseCode.markdown"))
	router.Handle("/resume", getMarkdownTemplateHandler("Resume", "resume.markdown"))
	router.Handle("/running", getMarkdownTemplateHandler("Running", "running.markdown"))
}

func main() {
	initializeBlogPosts()
	router := mux.NewRouter()
	initializeRoutes(router)
	http.ListenAndServe(":8080", router)
}
