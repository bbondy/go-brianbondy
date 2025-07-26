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

type ActivityTypeBreakdown struct {
	Type       string  `json:"type"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type RunningPage struct {
	Title                 string
	MarkdownSlug          string
	Runs                  Runs
	ContributionGraph     *ContributionGraph
	StravaRunTotals       StravaRunTotals
	ActivityTypeBreakdown []ActivityTypeBreakdown
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
	Years                            []int
	ShareUrl, Description, ImagePath string
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
