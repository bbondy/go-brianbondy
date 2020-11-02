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
	"encoding/json"
)

type Page struct {
	Title, Content string
//	FBTitle, FBSiteName, FBShareUrl, FBDescription, FBImageUrl string
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

func blogPostPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: getTitle("Blog posts"),
		Content: "Test content",
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
		}
		t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/simpleMarkdown.html"))
		t.Execute(w, p)
	}
}

func main() {
	var blogPosts []data.BlogPost
	myJsonString := `[{"title":"post 1"}, {"id": 3, "title":"post 2"}]`
	err := json.Unmarshal([]byte(myJsonString), &blogPosts)
	if (err != nil) {
		fmt.Println("Error parsing JSON")
	} else {
		fmt.Printf("Blog posts: %+v", blogPosts)
	}


	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", blogPostPage)
	http.HandleFunc("/blog/filters", filtersPage)
	http.HandleFunc("/about", getMarkdownTemplate("About", "about.markdown"))
	http.HandleFunc("/other", getMarkdownTemplate("Other", "other.markdown"))
	http.HandleFunc("/contact", getMarkdownTemplate("Contact", "contact.markdown"))
	http.HandleFunc("/projects", getMarkdownTemplate("Projects", "projects.markdown"))
	http.HandleFunc("/advice", getMarkdownTemplate("Advice", "advice.markdown"))
	http.HandleFunc("/books", getMarkdownTemplate("Books", "books.markdown"))
	http.HandleFunc("/braille", getMarkdownTemplate("Braille", "braille.markdown"))
	http.HandleFunc("/compression", getMarkdownTemplate("Compression", "compression.markdown"))
	http.HandleFunc("/compression/huffman", getMarkdownTemplate("Huffman Compression", "compression/huffman.markdown"))
	http.HandleFunc("/compression/BWT", getMarkdownTemplate("Burrows-Wheeler", "compression/BWT.markdown"))
	http.HandleFunc("/compression/PPM", getMarkdownTemplate("Burrows-Wheeler", "compression/PPM.markdown"))
	http.HandleFunc("/math", getMarkdownTemplate("Mathematics", "math.markdown"))
	http.HandleFunc("/math/main", getMarkdownTemplate("Main", "math/main.markdown"))
	http.HandleFunc("/math/pi", getMarkdownTemplate("Pi", "math/pi.markdown"))
	http.HandleFunc("/math/primes", getMarkdownTemplate("Primes", "math/primes.markdown"))
	http.HandleFunc("/math/numberTheory", getMarkdownTemplate("Mathematics", "math/numberTheory.markdown"))
	http.HandleFunc("/math/graphTheory", getMarkdownTemplate("Mathematics", "math/graphTheory.markdown"))
	http.HandleFunc("/math/mathTricks", getMarkdownTemplate("Mathematics", "math/mathTricks.markdown"))

	http.HandleFunc("/morseCode", getMarkdownTemplate("Morse Code", "morseCode.markdown"))
	http.HandleFunc("/resume", getMarkdownTemplate("Resume", "resume.markdown"))
	http.HandleFunc("/running", getMarkdownTemplate("Running", "running.markdown"))
	http.ListenAndServe(":8080", nil)
}
