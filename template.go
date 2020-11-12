package main

import (
	"html/template"
	"net/http"
	"fmt"
	"io/ioutil"
	"reflect"
	"github.com/bbondy/go-brianbondy/data"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type Page struct {
	Title, Content, MarkdownSlug string
	BlogPost data.BlogPost
	BlogPostBody string
	BlogPostUri string
	BlogPostDate string
	NextPage int
	PrevPage int
	MaxPage int
	Tag string
//	FBTitle, FBSiteName, FBShareUrl, FBDescription, FBImageUrl string
}

func slugifyTitle(title string) string {
	title = strings.ToLower(title)
	return strings.ReplaceAll(title, " ", "-");
}

func getTitle(titleSlug string) string {
	return titleSlug + " - " + "Brian R. Bondy";
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

var markdownMap = make(map[string]string);
var tagMap = make(map[string][]data.BlogPost);
var blogPosts []data.BlogPost

var funcMap = template.FuncMap{
		"avail": avail,
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
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

func redirect(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	replacements := map[string]string{
		"/blog/page/": "/page/",
		"/blog/tagged/": "/tagged/",
	}
	for from, to := range replacements {
		path = strings.ReplaceAll(path, from, to);
	}
	http.Redirect(w, r, path, 302)
}

func errorPage(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, message)
}

// Keeps it simple with 1 blog post per page
func blogPostPage(w http.ResponseWriter, r *http.Request) {

	blogPostIndex := 0;//page

	vars := mux.Vars(r)
  page, pageOK := vars["page"]
	if (pageOK) {
		blogPostIndex, _ = strconv.Atoi(page);
		blogPostIndex -= 1;
	}

	filteredBlogPosts := blogPosts;
	tag, tagOK := vars["tag"]
	if (tagOK) {
		filteredBlogPosts = tagMap[tag]
	}

	if blogPostIndex >= len(filteredBlogPosts) || blogPostIndex < 0 {
		errorPage(w, "No blog posts for this query")
		return
	}

	const (
    layoutISO = "2006-01-02"
    layoutUS  = "January 2, 2006"
	)
	parsedDate, _ := time.Parse(layoutISO, filteredBlogPosts[blogPostIndex].Created);

	p := &Page{
		Title: getTitle("Blog posts"),
		BlogPost: filteredBlogPosts[blogPostIndex],
		BlogPostBody: getMarkdownData("blog/" + strconv.Itoa(filteredBlogPosts[blogPostIndex].Id) + ".markdown"),
		BlogPostUri: "/" + strconv.Itoa(filteredBlogPosts[blogPostIndex].Id) + "/" + slugifyTitle(filteredBlogPosts[blogPostIndex].Title),
		BlogPostDate: parsedDate.Format(layoutUS),
		NextPage: blogPostIndex + 2,
		PrevPage: blogPostIndex,
		MaxPage: len(filteredBlogPosts),
		Tag: tag,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/blogPost.html"))
	t.Execute(w, p)
}

func filtersPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: getTitle("Filters"),
		Content: "Test content - filters",
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/filters.html"))
	t.Execute(w, p)
}

func getMarkdownTemplate(titleSlug string, markdownSlug string) func(w http.ResponseWriter, r *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		p := &Page{
			Title: getTitle(titleSlug),
			Content: getMarkdownData(markdownSlug),
			MarkdownSlug: markdownSlug,
		}
		t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/simpleMarkdown.html"))
		t.Execute(w, p)
	}
}

func initializeBlogPosts() {
	blogPostManifest, _ := ioutil.ReadFile("data/blogPostManifest.json");
	err := json.Unmarshal([]byte(blogPostManifest), &blogPosts)
	if (err != nil) {
		fmt.Println("Error parsing JSON")
		os.Exit(1);
	}

	for _, blogPost := range blogPosts {
		for _, tag := range blogPost.Tags {
			tagMap[tag] = append(tagMap[tag], blogPost)
		}
	}
}

func initializeRoutes(router *mux.Router) {
	fs := http.FileServer(http.Dir("static/"))
	s := http.StripPrefix("/static/", fs)
	router.PathPrefix("/static/").Handler(s)
	router.HandleFunc("/", blogPostPage)
	router.HandleFunc("/page/{page:[0-9]+}", blogPostPage)
	router.HandleFunc("/page/{page}", blogPostPage)
	router.HandleFunc("/tagged/{tag}", blogPostPage)
	router.HandleFunc("/tagged/{tag}/page/{page}", blogPostPage)
	router.HandleFunc("/blog/page/{page}", redirect)
	router.HandleFunc("/blog/tagged/{tag}", redirect)
	router.HandleFunc("/blog/tagged/{tag}/page/{page}", redirect)
	router.HandleFunc("/blog/filters", filtersPage)
	router.HandleFunc("/about", getMarkdownTemplate("About", "about.markdown"))
	router.HandleFunc("/other", getMarkdownTemplate("Other", "other.markdown"))
	router.HandleFunc("/contact", getMarkdownTemplate("Contact", "contact.markdown"))
	router.HandleFunc("/projects", getMarkdownTemplate("Projects", "projects.markdown"))
	router.HandleFunc("/advice", getMarkdownTemplate("Advice", "advice.markdown"))
	router.HandleFunc("/books", getMarkdownTemplate("Books", "books.markdown"))
	router.HandleFunc("/braille", getMarkdownTemplate("Braille", "braille.markdown"))
	router.HandleFunc("/compression", getMarkdownTemplate("Compression", "compression.markdown"))
	router.HandleFunc("/compression/huffman", getMarkdownTemplate("Huffman Compression", "compression/huffman.markdown"))
	router.HandleFunc("/compression/BWT", getMarkdownTemplate("Burrows-Wheeler", "compression/BWT.markdown"))
	router.HandleFunc("/compression/PPM", getMarkdownTemplate("Burrows-Wheeler", "compression/PPM.markdown"))
	router.HandleFunc("/math", getMarkdownTemplate("Mathematics", "math.markdown"))
	router.HandleFunc("/math/main", getMarkdownTemplate("Main", "math/main.markdown"))
	router.HandleFunc("/math/pi", getMarkdownTemplate("Pi", "math/pi.markdown"))
	router.HandleFunc("/math/primes", getMarkdownTemplate("Primes", "math/primes.markdown"))
	router.HandleFunc("/math/numberTheory", getMarkdownTemplate("Mathematics", "math/numberTheory.markdown"))
	router.HandleFunc("/math/graphTheory", getMarkdownTemplate("Mathematics", "math/graphTheory.markdown"))
	router.HandleFunc("/math/mathTricks", getMarkdownTemplate("Mathematics", "math/mathTricks.markdown"))
	router.HandleFunc("/morseCode", getMarkdownTemplate("Morse Code", "morseCode.markdown"))
	router.HandleFunc("/resume", getMarkdownTemplate("Resume", "resume.markdown"))
	router.HandleFunc("/running", getMarkdownTemplate("Running", "running.markdown"))
}

func main() {
	initializeBlogPosts();
	router := mux.NewRouter()
	initializeRoutes(router);
	http.ListenAndServe(":8080", router)
}
