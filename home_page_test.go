package main

import (
	"fmt"
	"html"
	"io"
	"net/http/httptest"
	"regexp"
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
	assert.Contains(t, content, `href="/static/css/editorial.css?v=4"`)
	assert.Contains(t, content, `href="/static/css/home.css?v=3"`)
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
	assert.Contains(t, content, `System theme. Click for light mode`)
	assert.Contains(t, content, `enum: ['light', 'dark', 'system']`)
	assert.Contains(t, content, `setTheme('system', false)`)
	assert.Contains(t, content, `localStorage.getItem('palette')`)
	assert.Contains(t, content, `localStorage.setItem('palette', palette)`)
	// Ocean is the default, so its variables live in the bare :root and
	// [data-theme="dark"] blocks rather than in palette specific ones.
	for _, palette := range []string{"forest", "ember", "violet", "rose", "gold", "slate", "teal", "indigo", "moss", "crimson", "plum"} {
		assert.Contains(t, content, fmt.Sprintf(`:root[data-palette=%q] {`, palette))
		assert.Contains(t, content, fmt.Sprintf(`:root[data-theme="dark"][data-palette=%q] {`, palette))
		assert.Contains(t, content, fmt.Sprintf(`'%s'`, palette))
	}
	assert.Contains(t, content, `'ocean'`)
}

func TestPaletteTogglePreviewsNextPaletteColor(t *testing.T) {
	setupTestEnvironment(t)

	recorder := httptest.NewRecorder()
	homePageHandler(recorder, httptest.NewRequest("GET", "/", nil))
	require.Equal(t, 200, recorder.Code)
	body, err := io.ReadAll(recorder.Result().Body)
	require.NoError(t, err)
	content := html.UnescapeString(string(body))

	assert.Contains(t, content, `color: var(--next-palette-swatch);`)
	readColor := func(selector, variable string) string {
		block := regexp.MustCompile(regexp.QuoteMeta(selector) + `\s*\{([^{}]*)\}`).FindStringSubmatch(content)
		require.Len(t, block, 2, "missing CSS block for %s", selector)
		property := regexp.MustCompile(regexp.QuoteMeta(variable) + `:\s*([^;]+);`).FindStringSubmatch(block[1])
		require.Len(t, property, 2, "missing %s in %s", variable, selector)
		return strings.TrimSpace(property[1])
	}

	palettes := []struct {
		name  string
		light string
		dark  string
	}{
		{"ocean", "#1483b3", "#59c0ff"},
		{"forest", "#177a57", "#61d3a4"},
		{"ember", "#b5512c", "#ff9a67"},
		{"violet", "#7350ad", "#bc9cff"},
		{"rose", "#b34b70", "#ff8fb5"},
		{"gold", "#a67816", "#e6be66"},
		{"slate", "#607b96", "#9db2c7"},
		{"teal", "#0f7f80", "#4fd2d2"},
		{"indigo", "#4453a6", "#8f9dff"},
		{"moss", "#5f7a1c", "#b3d15f"},
		{"crimson", "#b32f38", "#ff8b91"},
		{"plum", "#973d8a", "#eb92dc"},
	}
	for index, palette := range palettes {
		nextPalette := palettes[(index+1)%len(palettes)]
		lightSelector := fmt.Sprintf(`:root[data-palette=%q]`, palette.name)
		darkSelector := fmt.Sprintf(`:root[data-theme="dark"][data-palette=%q]`, palette.name)
		if palette.name == "ocean" {
			lightSelector = ":root"
			darkSelector = `[data-theme="dark"]`
		}
		assert.Equal(t, nextPalette.light, readColor(lightSelector, "--next-palette-swatch"), "%s in light mode", palette.name)
		assert.Equal(t, nextPalette.dark, readColor(darkSelector, "--next-palette-swatch"), "%s in dark mode", palette.name)
	}
}
