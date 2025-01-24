package data

type BlogPost struct {
	Id      int      `json:"id"`
	Title   string   `json:"title"`
	Created string   `json:"created"`
	Tags    []string `json:"tags"`
	FBImagePath *string `json:"fbImagePath,omitempty"`
	FBDescription *string `json:"fbDescription,omitempty"`
}
type BlogPosts []BlogPost
