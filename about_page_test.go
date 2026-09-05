package main

import (
	"html"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAboutPageContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/about", nil)
	aboutPageHandler(w, r)

	assert.Equal(t, 200, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	content := string(body)
	plainText := html.UnescapeString(content)

	assert.Contains(t, plainText, "About")
	assert.Contains(t, plainText, "I’m Brian R. Bondy.")
	assert.Contains(t, plainText, "Shannon, Link, Ronnie, and Asher")
	assert.Contains(t, plainText, "264.7 miles")
	assert.Contains(t, plainText, "Tahoe 200")
	assert.Contains(t, plainText, "Hutter Prize")
	assert.Contains(t, plainText, "Ray Charles")
	assert.Contains(t, plainText, "Microsoft MVP")
	assert.Contains(t, plainText, "Nymeria")
	assert.Contains(t, plainText, "Remy and Emile")
	assert.NotContains(t, plainText, "Loki")
	assert.Contains(t, content, "href=\"/resume\"")
	assert.Contains(t, content, "href=\"/running\"")
	assert.Contains(t, content, "id=\"personal-index\"")
	assert.NotContains(t, content, "id=\"north-star\"")
	assert.NotContains(t, plainText, "What matters to me")
	assert.Contains(t, plainText, "raison d’être")
	assert.Contains(t, plainText, "turning ideas into working systems")
	assert.Contains(t, plainText, "difficult things that demand growth")
	assert.NotContains(t, content, `<footer class="about-local-footer">\n    <span>Belle River, Ontario, Canada</span>`)
}
