package main

import (
	"encoding/json"
	"io/ioutil"
	"fmt"
	"sort"
	"time"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

var markdownMap = make(map[string]string)
var blogPostTagMap = make(map[string][]data.BlogPost)
var blogPostYearMap = make(map[int][]data.BlogPost)
var blogPosts []data.BlogPost
var blogPostIdMap = make(map[int]data.BlogPost)
var tagCountMap = make(map[string]int)
var sortedTags []string

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

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
