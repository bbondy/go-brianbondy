package main

import (
	"fmt"
	"strings"
	"html/template"
	"net/url"
	"reflect"
)

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
	"getYearCount": func(year int) int {
		posts, ok := blogPostYearMap[year]
		if !ok {
			return 0
		}
		return len(posts)
	},
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
