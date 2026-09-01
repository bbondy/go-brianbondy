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

type CareerPage struct {
	Title        string
	Description  string
	MarkdownSlug string
	ShareUrl     string
	Profile      CareerProfile
}

type AboutPage struct {
	Title        string
	Description  string
	ImagePath    string
	MarkdownSlug string
	ShareUrl     string
}

type ContactPage struct {
	Title        string
	Description  string
	MarkdownSlug string
	ShareUrl     string
}

type ActivityTypeBreakdown struct {
	Type       string  `json:"type"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type RunningPage struct {
	Title                 string
	Description           string
	MarkdownSlug          string
	ShareUrl              string
	Runs                  Runs
	ContributionGraph     *ContributionGraph
	ContributionGraph2D   *ContributionGraph2D
	StravaRunTotals       StravaRunTotals
	ActivityTypeBreakdown []ActivityTypeBreakdown
	Years                 []int
	SelectedYear          string
	LastUpdatedDate       string
	GraphJSON             template.JS
}

type StravaRunTotals struct {
	TotalRuns        int
	TotalDistanceKm  float64
	TotalElevationM  int
	TotalTimeDays    int
	TotalTimeHours   int
	TotalTimeMinutes int
}

type BlogPostPage struct {
	Title, Content, MarkdownSlug     string
	BlogPost                         BlogPost
	BlogPostBody                     string
	BlogPostDate                     string
	ReadingTime                      int
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
	TagGroups                        []TagGroup
	Years                            []int
	ShareUrl, Description, ImagePath string
}

// TagGroup is a named collection of related blog tags for navigation.
type TagGroup struct {
	Name string
	Tags []string
}

type BlogPostPreview struct {
	BlogPost    BlogPost
	Preview     template.HTML
	PostDate    string
	PostUrl     string
	ReadingTime int
}

type HomePage struct {
	Title        string
	Posts        []BlogPostPreview
	AllPosts     []BlogPostPreview
	MarkdownSlug string
}

type AllPostsPage struct {
	Title        string
	Description  string
	Posts        []BlogPostPreview
	MarkdownSlug string
	ShareUrl     string
	Tag          string
	Year         int
}

type ProjectsPage struct {
	Title        string
	Description  string
	MarkdownSlug string
	ShareUrl     string
	Projects     Projects
	BlogPostMap  map[int]BlogPost
}

type Picture struct {
	Id    int      `json:"id"`
	Image string   `json:"image"`
	Tags  []string `json:"tags"`
}

type PicturesPage struct {
	Title         string
	MarkdownSlug  string
	Pictures      []Picture
	FilterTag     string
	FilterBlogID  int
	BlogPostTitle string
}
