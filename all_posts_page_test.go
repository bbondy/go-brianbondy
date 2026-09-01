package main

import (
	"fmt"
	"html"
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllPostsPagePreservesPostData(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	allPostsHandler(w, httptest.NewRequest("GET", "/all", nil))

	assert.Equal(t, 200, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	content := html.UnescapeString(string(body))

	assert.Equal(t, len(blogPosts), strings.Count(content, `class="all-posts-card"`))
	for _, post := range blogPosts {
		postURL := fmt.Sprintf("/blog/%d/%s", post.Id, slugifyTitle(post.Title))
		assert.Contains(t, content, post.Title)
		assert.Contains(t, content, `href="`+postURL+`"`)
		assert.Contains(t, content, fmt.Sprintf(`/pictures?blog_id=%d`, post.Id))
		for _, tag := range post.Tags {
			assert.Contains(t, content, `href="/all?tag=`+url.QueryEscape(tag)+`"`)
		}
	}
}

func TestAllPostsPageUsesEditorialArchiveLayout(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	allPostsHandler(w, httptest.NewRequest("GET", "/all", nil))

	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	content := html.UnescapeString(string(body))

	assert.Contains(t, content, `<article class="all-posts-page editorial-page"`)
	assert.Contains(t, content, `<h1 class="editorial-title">`)
	assert.Contains(t, content, `All Blog Posts`)
	assert.Contains(t, content, `id="search-posts"`)
	assert.Contains(t, content, `/static/css/editorial.css?v=2`)
	assert.Contains(t, content, `/static/css/all-posts.css?v=2`)
	assert.Contains(t, content, `Brian Bondy's complete blog archive`)
}

func TestAllPostsPageShowsActiveFilters(t *testing.T) {
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	allPostsHandler(w, httptest.NewRequest("GET", "/all?tag=test&year=2023", nil))

	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	content := html.UnescapeString(string(body))

	assert.Equal(t, 2, strings.Count(content, `class="all-posts-card"`))
	assert.Contains(t, content, `class="all-posts-active-filters"`)
	assert.Contains(t, content, `test</span>`)
	assert.Contains(t, content, `2023</span>`)
	assert.Contains(t, content, `href="/all">Clear filters</a>`)
	assert.Contains(t, content, `https://brianbondy.com/all?tag=test&year=2023`)
}
