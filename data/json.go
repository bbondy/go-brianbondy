package data

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func ReadJsonFile(fileName string, v interface{}) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, v)
	if err != nil {
		return err
	}
	return nil
}

// MarshalJSON encodes the extension list into response JSON
func (blogPosts *BlogPosts) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &blogPosts)
	if err != nil {
		log.Panic(err)
	}
	return err
}
