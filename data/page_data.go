package data

import (
	"html/template"
)

type SimpleMarkdownPage struct {
	Title        string
	Content      string
	MarkdownSlug string
	ShareUrl     string
	ErrorCode    int // Added for error pages
}

type RunningPage struct {
	Title        string
	MarkdownSlug string
	Runs         Runs
}

type BlogPostPage struct {
	Title, Content, MarkdownSlug     string
	BlogPost                         BlogPost
	BlogPostBody                     string
	BlogPostDate                     string
	NextPost                         *BlogPost
	PrevPost                         *BlogPost
	Tag                              string
	Year                             int
	ShareUrl, Description, ImagePath string
}

type FiltersPage struct {
	Title, Content                   string
	MarkdownSlug                     string
	TagCountMap                      map[string]int
	SortedTags                       []string
	Years                            []int
	ShareUrl, Description, ImagePath string
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
	AllPosts     []BlogPostPreview
	MarkdownSlug string
}

type AllPostsPage struct {
	Title        string
	Posts        []BlogPostPreview
	MarkdownSlug string
	Tag          string
	Year         int
}

type ProjectsPage struct {
	Title        string
	MarkdownSlug string
	Projects     Projects
}
