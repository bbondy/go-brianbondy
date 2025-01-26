package data

import (
	"html/template"
)

type SimpleMarkdownPage struct {
	Title, Content                         string
	MarkdownSlug                           string
	FBShareUrl, FBDescription, FBImagePath string
}

type BlogPostPage struct {
	Title, Content, MarkdownSlug           string
	BlogPost                               BlogPost
	BlogPostBody                           string
	BlogPostDate                           string
	NextPost                               *BlogPost
	PrevPost                               *BlogPost
	Tag                                    string
	Year                                   int
	FBShareUrl, FBDescription, FBImagePath string
}

type FiltersPage struct {
	Title, Content                         string
	MarkdownSlug                           string
	TagCountMap                            map[string]int
	SortedTags                             []string
	Years                                  []int
	FBShareUrl, FBDescription, FBImagePath string
}

type BlogPostPreview struct {
	BlogPost BlogPost
	Preview  template.HTML
	PostDate string
	PostUrl  string
}

type HomePage struct {
	Title        string
	Posts        []BlogPostPreview
	MarkdownSlug string
}
