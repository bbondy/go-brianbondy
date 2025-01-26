package data

type BlogPost struct {
	Id          int      `json:"id"`
	Title       string   `json:"title"`
	Created     string   `json:"created"`
	Tags        []string `json:"tags"`
	ImagePath   *string  `json:"fbImagePath,omitempty"`
	Description *string  `json:"fbDescription,omitempty"`
}
type BlogPosts []BlogPost
