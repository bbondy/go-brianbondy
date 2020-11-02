package main

import (
	"html/template"
	"net/http"
	"fmt"
	"reflect"
	"github.com/bbondy/go-brianbondy/data"
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
		fmt.Println("----avail")
    v := reflect.ValueOf(data)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }
    if v.Kind() != reflect.Struct {
			  fmt.Println("----avail false")
        return false
    }
    if (v.FieldByName(name).IsValid()) {
	    fmt.Println("----avail valid true")
	} else {
	    fmt.Println("----avail valid false")
		}
    return v.FieldByName(name).IsValid()
}

func blogPostPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: getTitle("Blog posts"),
		Content: "Test content",
	}
	funcMap := template.FuncMap{
		"avail": avail,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/blogPost.html"))
	t.Execute(w, p)
}

func aboutPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: getTitle("About"),
		Content: "Test content - about",
	}
	funcMap := template.FuncMap{
		"avail": avail,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/about.html"))
	t.Execute(w, p)
}

func contactPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: getTitle("Contact"),
		Content: "Test content - contact",
	}
	funcMap := template.FuncMap{
		"avail": avail,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/contact.html"))
	t.Execute(w, p)
}

func projectsPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: getTitle("Projects"),
		Content: "Test content - projects",
	}
	funcMap := template.FuncMap{
		"avail": avail,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/projects.html"))
	t.Execute(w, p)
}

func otherPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: getTitle("Other"),
		Content: "Test content - other",
	}
	funcMap := template.FuncMap{
		"avail": avail,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/other.html"))
	t.Execute(w, p)
}

func filtersPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: getTitle("Filters"),
		Content: "Test content - filters",
	}
	funcMap := template.FuncMap{
		"avail": avail,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/filters.html"))
	t.Execute(w, p)
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
	http.HandleFunc("/about", aboutPage)
	http.HandleFunc("/other", otherPage)
	http.HandleFunc("/contact", contactPage)
	http.HandleFunc("/projects", projectsPage)
	http.ListenAndServe(":8080", nil)
}
