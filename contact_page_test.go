package main

import (
	"html"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContactPageContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/contact", nil)
	contactPageHandler(w, r)

	assert.Equal(t, 200, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	content := string(body)
	plainText := html.UnescapeString(content)

	assert.Contains(t, plainText, "Contact me.")
	assert.NotContains(t, plainText, "The inbox is the shortest path.")
	assert.NotContains(t, plainText, "A little context in the first note is always helpful.")
	assert.Contains(t, plainText, "bbondy [at] gmail.com")
	assert.Contains(t, content, "href=\"mailto:bbondy@gmail.com\"")
	assert.Contains(t, content, "href=\"https://github.com/bbondy/\"")
	assert.Contains(t, content, "href=\"https://www.strava.com/athletes/6932803\"")
	assert.Contains(t, content, "rel=\"noopener noreferrer\"")
	assert.Contains(t, content, "aria-labelledby=\"online-heading\"")
}
