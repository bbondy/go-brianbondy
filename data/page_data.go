package data

type SimpleMarkdownPage struct {
	Title, Content                        string
	MarkdownSlug                          string
	FBShareUrl, FBDescription, FBImageUrl string
}

type BlogPostPage struct {
	Title, Content, MarkdownSlug          string
	BlogPost                              BlogPost
	BlogPostBody                          string
	BlogPostUri                           string
	BlogPostDate                          string
	NextPage                              int
	PrevPage                              int
	MaxPage                               int
	Tag                                   string
	Year                                  int
	FBShareUrl, FBDescription, FBImageUrl string
}

type FiltersPage struct {
	Title, Content                        string
	MarkdownSlug                          string
	TagCountMap                           map[string]int
	SortedTags                            []string
	Years                                 []int
	FBShareUrl, FBDescription, FBImageUrl string
}
