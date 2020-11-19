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
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	layoutISO = "2006-01-02"
	layoutUS  = "January 2, 2006"
)

type SimpleMarkdownPage struct {
	Title, Content                        string
	MarkdownSlug                          string
	FBShareUrl, FBDescription, FBImageUrl string
}

type BlogPostPage struct {
	Title, Content, MarkdownSlug          string
	BlogPost                              data.BlogPost
	BlogPostBody                          string
	BlogPostUri                           string
	BlogPostDate                          string
	NextPage                              int
	PrevPage                              int
	MaxPage                               int
	Tag                                   string
	Year                                  int
	FBShareUrl, FBDescription, FBImageUrl string
}

type FiltersPage struct {
	Title, Content                        string
	MarkdownSlug                          string
	TagCountMap                           map[string]int
	SortedTags                            []string
	Years                                 []int
	FBShareUrl, FBDescription, FBImageUrl string
}

func slugifyTitle(title string) string {
	title = strings.ToLower(title)
	str := strings.ReplaceAll(title, " ", "-")
	reg, _ := regexp.Compile("[^a-zA-Z0-9-]+")
	str = reg.ReplaceAllString(str, "")
	return strings.ReplaceAll(str, "--", "-")
}

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

func getTitle(titleSlug string) string {
	return titleSlug + " - " + "Brian R. Bondy"
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
	p := &SimpleMarkdownPage{
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
	page, pageOK := vars["page"]
	if pageOK {
		blogPostIndex, _ = strconv.Atoi(page)
		blogPostIndex -= 1
	}

	filteredBlogPosts := blogPosts
	yearStr, yearOK := vars["year"]
	year := 0
	if yearOK {
		year, _ = strconv.Atoi(yearStr)
		filteredBlogPosts = blogPostYearMap[year]
	}

	tag, tagOK := vars["tag"]
	if tagOK {
		filteredBlogPosts = blogPostTagMap[tag]
	}

	idStr, idOK := vars["id"]
	if idOK {
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

	p := &BlogPostPage{
		Title:        getTitle("Blog posts"),
		BlogPost:     filteredBlogPosts[blogPostIndex],
		BlogPostBody: getMarkdownData("blog/" + strconv.Itoa(filteredBlogPosts[blogPostIndex].Id) + ".markdown"),
		BlogPostUri:  "/blog/" + strconv.Itoa(filteredBlogPosts[blogPostIndex].Id) + "/" + slugifyTitle(filteredBlogPosts[blogPostIndex].Title),
		BlogPostDate: parsedDate.Format(layoutUS),
		NextPage:     blogPostIndex + 2,
		PrevPage:     blogPostIndex,
		MaxPage:      len(filteredBlogPosts),
		Tag:          tag,
		Year:         year,
		MarkdownSlug: "blog",
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/blogPost.html"))
	t.Execute(w, p)
}

func filtersPageHandler(w http.ResponseWriter, r *http.Request) {
	current_year := time.Now().Year()
	year_range := make([]int, 10)
	for i := range year_range {
		year_range[i] = current_year - i
	}

	p := &FiltersPage{
		Title:        getTitle("Filters"),
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
		p := &SimpleMarkdownPage{
			Title:        getTitle(titleSlug),
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
		fmt.Println("Error parsing JSON")
		os.Exit(1)
	}

	for _, blogPost := range blogPosts {
		for _, tag := range blogPost.Tags {
			blogPostTagMap[tag] = append(blogPostTagMap[tag], blogPost)
			_, countOk := tagCountMap[tag]
			if !countOk {
				tagCountMap[tag] = 1
			} else {
				tagCountMap[tag] += 1
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

	router.Handle("/", handleBlogPost)
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
