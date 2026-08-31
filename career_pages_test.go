package main

import (
	"html"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResumePageContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/resume", nil)
	resumePageHandler(w, r)

	assert.Equal(t, 200, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	content := string(body)
	plainText := html.UnescapeString(content)

	assert.Contains(t, plainText, "Brian R. Bondy")
	assert.Contains(t, plainText, "Co-founder & CTO, Brave Software")
	assert.Contains(t, content, "career-print-signature")
	assert.Contains(t, plainText, "120M+")
	assert.Contains(t, content, "https://brave.com/about/")
	assert.Contains(t, content, "Khan Academy")
	assert.Contains(t, content, "Mozilla")
	assert.Contains(t, content, "April 2014–2015")
	assert.Contains(t, content, "href=\"/career\"")
	assert.Contains(t, content, "data-print-page")
	assert.Contains(t, plainText, "Save as PDF")

	assert.NotContains(t, content, "Microsoft SourceSafe")
	assert.NotContains(t, content, "Silverlight")
	assert.NotContains(t, strings.ToLower(content), "hard working")
}

func TestCareerArchiveContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/career", nil)
	careerPageHandler(w, r)

	assert.Equal(t, 200, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	content := string(body)
	plainText := html.UnescapeString(content)

	assert.Contains(t, plainText, "Career & Technical Archive")
	assert.Contains(t, plainText, "Brian R. Bondy")
	assert.Contains(t, plainText, "Co-founder & CTO, Brave Software")
	assert.Contains(t, content, "career-print-signature")
	assert.Contains(t, plainText, "Microsoft SourceSafe")
	assert.Contains(t, plainText, "Silverlight")
	assert.Contains(t, plainText, "Borland C++ 4.2")
	assert.Contains(t, plainText, "File Access Manager")
	assert.Contains(t, plainText, "Evernote")
	assert.Contains(t, content, "<details")
	assert.Contains(t, content, "href=\"/resume\"")
	assert.Contains(t, content, "data-print-page")
	assert.Contains(t, plainText, "Save as PDF")
}
