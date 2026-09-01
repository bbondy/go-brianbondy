package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func renderFiltersPage(t *testing.T) string {
	t.Helper()
	setupTestEnvironment(t)

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/blog/filters", nil)
	if err != nil {
		t.Fatal(err)
	}

	filtersPageHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("filters page status = %d, want %d", w.Code, http.StatusOK)
	}

	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func TestFiltersPagePreservesArchiveNavigation(t *testing.T) {
	body := renderFiltersPage(t)

	for _, expected := range []string{
		`href="/all?year=2023"`,
		`href="/all?year=2022"`,
		`href="/all?tag=test"`,
		`href="/all?tag=golang"`,
		`2 posts`,
		`1 post`,
		`class="filters-tag-count">2</span>`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("filters page is missing %q", expected)
		}
	}
}

func TestFiltersPageUsesEditorialLayout(t *testing.T) {
	body := renderFiltersPage(t)

	for _, expected := range []string{
		`class="filters-page editorial-page"`,
		`class="filters-hero editorial-hero"`,
		`Browse the archive`,
		`Posts by year`,
		`Posts by tag`,
		`id="search-tags"`,
		`href="/static/css/editorial.css?v=1"`,
		`href="/static/css/filters.css?v=1"`,
		`content="Browse Brian Bondy&#39;s complete blog archive by year and topic."`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("filters page is missing %q", expected)
		}
	}

	if strings.Contains(body, "Test content - filters") {
		t.Error("filters page exposes placeholder copy")
	}
}
