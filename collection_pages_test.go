package main

import (
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderCollectionPage(t *testing.T, path string, handler http.HandlerFunc) string {
	t.Helper()
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodGet, path, nil))
	require.Equal(t, http.StatusOK, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	return string(body)
}

func TestBooksPagePreservesManifestData(t *testing.T) {
	data.ClearBooksCache()
	books, err := data.GetBooks()
	require.NoError(t, err)
	require.Len(t, books, 241)

	body := renderCollectionPage(t, "/books", booksHandler)
	unescaped := html.UnescapeString(body)
	assert.Equal(t, len(books), strings.Count(body, `class="collection-card book-card"`))
	assert.Contains(t, body, `id="search-books"`)

	for _, book := range books {
		assert.Contains(t, unescaped, book.Title)
		assert.Contains(t, unescaped, book.Author)
		assert.Contains(t, unescaped, book.Publisher)
		assert.Contains(t, unescaped, strconv.Itoa(book.Year))
		assert.Contains(t, unescaped, strconv.Itoa(book.Pages)+" pages")
		assert.Contains(t, unescaped, `href="`+book.GoodreadsURL+`"`)
		assert.Contains(t, unescaped, `href="`+book.AudibleURL+`"`)
	}
}

func TestInterviewsPagePreservesManifestData(t *testing.T) {
	data.ClearInterviewsCache()
	interviews, err := data.GetInterviews()
	require.NoError(t, err)
	require.Len(t, interviews, 11)

	body := renderCollectionPage(t, "/interviews", interviewsHandler)
	unescaped := html.UnescapeString(body)
	assert.Equal(t, len(interviews), strings.Count(body, `class="collection-card interview-card"`))

	for _, interview := range interviews {
		assert.Contains(t, unescaped, interview.Title)
		assert.Contains(t, unescaped, interview.Date)
		assert.Contains(t, unescaped, interview.Platform)
		assert.Contains(t, unescaped, interview.Channel)
		assert.Contains(t, unescaped, interview.Description)
		assert.Contains(t, unescaped, `href="`+interview.URL+`"`)
	}
}

func TestCheatsheetsPagePreservesManifestData(t *testing.T) {
	data.ClearCheatsheetsCache()
	cheatsheets, err := data.GetCheatsheets()
	require.NoError(t, err)
	require.Len(t, cheatsheets, 15)

	body := renderCollectionPage(t, "/cheatsheets", cheatsheetsHandler)
	unescaped := html.UnescapeString(body)
	assert.Equal(t, len(cheatsheets), strings.Count(body, `class="collection-card cheatsheet-card"`))

	for _, cheatsheet := range cheatsheets {
		assert.Contains(t, unescaped, cheatsheet.Title)
		assert.Contains(t, unescaped, cheatsheet.Description)
		assert.Contains(t, unescaped, `href="/cheatsheets/`+cheatsheet.Slug+`"`)
	}
}

func TestPicturesPagePreservesManifestData(t *testing.T) {
	ClearPicturesCache()
	pictures, err := getCachedPictures()
	require.NoError(t, err)
	require.Len(t, pictures, 318)

	body := renderCollectionPage(t, "/pictures", picturesHandler)
	assert.Equal(t, len(pictures), strings.Count(body, `class="picture-entry"`))
	for _, picture := range pictures {
		assert.Contains(t, body, `data-lightbox-src="/`+picture.Image+`"`)
	}
}

func TestAdvicePagePreservesEveryEntry(t *testing.T) {
	originalMarkdownMap := markdownMap
	markdownMap = make(map[string]string)
	t.Cleanup(func() { markdownMap = originalMarkdownMap })

	body := renderCollectionPage(t, "/advice", adviceHandler)
	unescaped := html.UnescapeString(body)
	assert.Equal(t, 22, strings.Count(body, `class="card-entry"`))
	assert.Contains(t, unescaped, "anything is understandable with the right background")
	assert.Contains(t, unescaped, "Arnold's Pump Club")
}

func TestCollectionPagesUseSharedEditorialStyles(t *testing.T) {
	tests := []struct {
		path    string
		title   string
		handler http.HandlerFunc
	}{
		{"/books", "Books", booksHandler},
		{"/interviews", "Interviews", interviewsHandler},
		{"/cheatsheets", "Cheatsheets", cheatsheetsHandler},
		{"/pictures", "Pictures gallery", picturesHandler},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			body := renderCollectionPage(t, test.path, test.handler)
			assert.Contains(t, body, `class="collection-page`)
			assert.Contains(t, body, `class="collection-hero editorial-hero"`)
			assert.Contains(t, body, `href="/static/css/editorial.css?v=3"`)
			assert.Contains(t, body, `href="/static/css/collections.css?v=2"`)
			assert.Contains(t, html.UnescapeString(body), test.title)
		})
	}
}
