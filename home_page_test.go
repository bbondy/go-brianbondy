package main

import (
	"fmt"
	"html"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHomePagePreservesLatestPostData(t *testing.T) {
	setupTestEnvironment(t)
	imagePath := "/static/img/test.webp"
	blogPosts[0].ImagePath = &imagePath

	w := httptest.NewRecorder()
	homePageHandler(w, httptest.NewRequest("GET", "/", nil))
	require.Equal(t, 200, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	content := html.UnescapeString(string(body))

	expectedCount := len(blogPosts)
	if expectedCount > 4 {
		expectedCount = 4
	}
	assert.Equal(t, expectedCount, strings.Count(content, `class="home-post-card`))
	for _, post := range blogPosts[:expectedCount] {
		postURL := fmt.Sprintf("/blog/%d/%s", post.Id, slugifyTitle(post.Title))
		assert.Contains(t, content, post.Title)
		assert.Contains(t, content, `href="`+postURL+`"`)
		assert.Contains(t, content, "This is content for blog post "+post.Title)
	}
	assert.Contains(t, content, `src="/static/img/test.webp"`)
}

func TestHomePageUsesEditorialLayout(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	homePageHandler(w, httptest.NewRequest("GET", "/", nil))
	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	content := html.UnescapeString(string(body))

	assert.Contains(t, content, `<article class="home-page editorial-page"`)
	assert.Contains(t, content, `<h1 class="editorial-title">Running, work, and life</h1>`)
	assert.Contains(t, content, `href="/static/css/editorial.css?v=2"`)
	assert.Contains(t, content, `href="/static/css/home.css?v=2"`)
	assert.Contains(t, content, `href="/running"`)
	assert.Contains(t, content, `href="/projects"`)
	assert.Contains(t, content, `href="/resume">Work & Career</a>`)
	assert.Contains(t, content, `href="/all"`)
	assert.Contains(t, content, `href="/blog/filters"`)
	assert.Contains(t, content, `content="Brian Bondy's writing about software, running, work, and life."`)
}

func TestHomePageNavigationAndThemeControls(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	homePageHandler(w, httptest.NewRequest("GET", "/", nil))
	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	content := html.UnescapeString(string(body))

	intro := `I’m a father, software builder, and ultrarunner. I co-founded Brave after working at Mozilla and Khan Academy. This site is an archive of race reports, software work, and other things I wanted to write down.`
	assert.Contains(t, content, `<p class="home-intro editorial-intro">`+intro+`</p>`)

	writingIndex := strings.Index(content, `<strong>Writing</strong>`)
	runningIndex := strings.Index(content, `<strong>Running</strong>`)
	projectsIndex := strings.Index(content, `<strong>Projects</strong>`)
	require.NotEqual(t, -1, writingIndex)
	require.NotEqual(t, -1, runningIndex)
	require.NotEqual(t, -1, projectsIndex)
	assert.Less(t, writingIndex, runningIndex)
	assert.Less(t, runningIndex, projectsIndex)

	assert.Contains(t, content, `id="palette-toggle"`)
	assert.Contains(t, content, `id="theme-toggle"`)
	assert.Contains(t, content, `localStorage.getItem('palette')`)
	assert.Contains(t, content, `localStorage.setItem('palette', palette)`)
	for _, palette := range []string{"ocean", "forest", "ember", "violet", "rose"} {
		assert.Contains(t, content, palette)
	}
}
