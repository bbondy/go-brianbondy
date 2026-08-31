package main

import (
	"fmt"
	"html/template"
	"reflect"
	"sync"
)

var (
	compiledTemplates     map[string]*template.Template
	compiledTemplatesOnce sync.Once
	compiledTemplatesErr  error
)

var templateFiles = map[string]string{
	"allPosts":       "templates/allPosts.html",
	"about":          "templates/about.html",
	"blogPost":       "templates/blogPost.html",
	"books":          "templates/books.html",
	"cheatsheets":    "templates/cheatsheets.html",
	"contact":        "templates/contact.html",
	"error":          "templates/error.html",
	"filters":        "templates/filters.html",
	"home":           "templates/home.html",
	"interviews":     "templates/interviews.html",
	"pictures":       "templates/pictures.html",
	"projects":       "templates/projects.html",
	"resume":         "templates/resume.html",
	"career":         "templates/career.html",
	"running":        "templates/running.html",
	"simpleMarkdown": "templates/simpleMarkdown.html",
}

// initializeTemplates parses every page template before the HTTP server accepts
// requests. The resulting templates are safe for concurrent execution.
func initializeTemplates() error {
	compiledTemplatesOnce.Do(func() {
		compiledTemplates = make(map[string]*template.Template, len(templateFiles))
		for name, pageFile := range templateFiles {
			t, err := template.New("base.html").Funcs(funcMap).ParseFiles("templates/base.html", pageFile)
			if err != nil {
				compiledTemplatesErr = fmt.Errorf("parse %s template: %w", name, err)
				return
			}
			compiledTemplates[name] = t
		}
	})
	return compiledTemplatesErr
}

func executeTemplate(wr templateWriter, name string, data interface{}) error {
	if err := initializeTemplates(); err != nil {
		return err
	}
	t, ok := compiledTemplates[name]
	if !ok {
		return fmt.Errorf("unknown template %q", name)
	}
	return t.Execute(wr, data)
}

type templateWriter interface {
	Write([]byte) (int, error)
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
