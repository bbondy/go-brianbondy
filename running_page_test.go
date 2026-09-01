package main

import (
	"html"
	"io"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunningPagePreservesMemorableRunData(t *testing.T) {
	data.ClearRunsCache()
	runs, err := data.GetRuns()
	require.NoError(t, err)

	w := httptest.NewRecorder()
	runningHandler(w, httptest.NewRequest("GET", "/running", nil))

	assert.Equal(t, 200, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	content := string(body)
	unescaped := html.UnescapeString(content)

	assert.Equal(t, len(runs), strings.Count(content, `class="run-card"`))
	assert.Contains(t, content, `id="contribution-graph"`)
	assert.Contains(t, content, `id="year-select"`)
	assert.Contains(t, content, `id="search-runs"`)

	for _, run := range runs {
		assert.Contains(t, unescaped, run.Title)
		assert.Contains(t, unescaped, run.Description)
		assert.Contains(t, unescaped, run.Time)
		assert.Contains(t, unescaped, run.Distance)
		assert.Contains(t, unescaped, run.Elevation)
		for _, url := range run.StravaURLs {
			assert.Contains(t, unescaped, `href="`+url+`"`)
		}
		if run.ImagePath != "" {
			assert.Contains(t, unescaped, `/static/`+run.ImagePath)
		}
		for _, imagePath := range run.ImagePaths {
			assert.Contains(t, unescaped, `/static/`+imagePath)
		}
		if run.BlogPostID != 0 {
			assert.Contains(t, unescaped, `/blog/`+strconv.Itoa(run.BlogPostID))
		}
	}
}

func TestRunningPageUsesEditorialArchiveLayout(t *testing.T) {
	w := httptest.NewRecorder()
	runningHandler(w, httptest.NewRequest("GET", "/running", nil))

	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	content := html.UnescapeString(string(body))

	assert.Contains(t, content, `<article class="running-page editorial-page"`)
	assert.Contains(t, content, `<h1 class="editorial-title">Runs</h1>`)
	assert.Contains(t, content, `/static/css/editorial.css?v=2`)
	assert.Contains(t, content, `/static/css/running.css?v=3`)
	assert.Contains(t, content, `Brian Bondy's running history`)
	assert.NotContains(t, content, `K<article class="running-page`)
}
