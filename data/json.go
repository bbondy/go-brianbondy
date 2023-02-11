package data

import (
	"encoding/json"
	"log"
)

// MarshalJSON encodes the extension list into response JSON
func (blogPosts *BlogPosts) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &blogPosts)
	if err != nil {
		log.Panic(err)
	}
	return err
}
