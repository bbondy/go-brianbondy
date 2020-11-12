package data

import (
	"encoding/json"
	"log"
)

type BlogPost struct {
  Id int `json:"id"`
  Title string `json:"title"`
  Created string `json:"created"`
  Tags []string `json:"tags"`
}
type BlogPosts []BlogPost

// MarshalJSON encodes the extension list into response JSON
func (blogPosts *BlogPosts) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &blogPosts)
	if err != nil {
		log.Panic(err)
	}
	return err
}
