package data

import (
	"encoding/json"
	"os"
)

type Book struct {
	Title        string `json:"title"`
	Author       string `json:"author"`
	Year         int    `json:"year"`
	Publisher    string `json:"publisher"`
	Pages        int    `json:"pages"`
	GoodreadsURL string `json:"goodreads_url"`
	AudibleURL   string `json:"audible_url"`
}

type Books []Book

var cachedBooks Books
var booksLoaded bool

func GetBooks() (Books, error) {
	if booksLoaded {
		return cachedBooks, nil
	}
	f, err := os.Open("data/booksManifest.json")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			// log error if needed
		}
	}()
	var books Books
	if err := json.NewDecoder(f).Decode(&books); err != nil {
		return nil, err
	}
	cachedBooks = books
	booksLoaded = true
	return cachedBooks, nil
}

func ClearBooksCache() {
	booksLoaded = false
	cachedBooks = nil
}
