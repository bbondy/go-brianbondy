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

func displayPage(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: "Blog posts - Brian R. Bondy",
		Content: "Test content",
	}
	funcMap := template.FuncMap{
		"avail": avail,
	}
	t := template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", "templates/blogPost.html"))
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
	http.HandleFunc("/", displayPage)
	http.ListenAndServe(":8080", nil)
}
