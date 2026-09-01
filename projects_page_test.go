package main

import (
	"html"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectsPagePreservesManifestData(t *testing.T) {
	data.ClearProjectsCache()
	projects, err := data.GetProjects()
	require.NoError(t, err)
	require.Len(t, projects, 57)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/projects", nil)
	projectsHandler(w, r)

	assert.Equal(t, 200, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	content := string(body)
	unescaped := html.UnescapeString(content)

	assert.Equal(t, len(projects), strings.Count(content, `class="projects-card"`))
	assert.Contains(t, content, `id="search-projects"`)
	assert.NotContains(t, content, `data-project-filter`)

	for _, project := range projects {
		assert.Contains(t, unescaped, project.Title)
		assert.Contains(t, unescaped, project.Description)
		if project.URL != "" {
			assert.Contains(t, unescaped, `href="`+project.URL+`"`)
		}
		if project.Website != "" {
			assert.Contains(t, unescaped, `href="`+project.Website+`"`)
		}
		if project.Github != "" {
			assert.Contains(t, unescaped, `href="`+project.Github+`"`)
		}
	}
}

func TestProjectsPageCopyStaysFactual(t *testing.T) {
	w := httptest.NewRecorder()
	projectsHandler(w, httptest.NewRequest("GET", "/projects", nil))

	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	content := html.UnescapeString(string(body))

	assert.Contains(t, content, `<h1 class="editorial-title">Projects</h1>`)
	assert.NotContains(t, content, "Some I'm the primary developer")
	assert.NotContains(t, content, "Some are full-fledged products")
}
