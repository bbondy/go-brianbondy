package data

type BlogPost struct {
	Id      int      `json:"id"`
	Title   string   `json:"title"`
	Created string   `json:"created"`
	Tags    []string `json:"tags"`
}
type BlogPosts []BlogPost
