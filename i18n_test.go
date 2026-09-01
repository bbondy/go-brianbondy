package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/bbondy/go-brianbondy/data"
	"github.com/gorilla/mux"
)

func TestTranslateUsesEnglishFallback(t *testing.T) {
	const source = "A string that is deliberately absent from every catalog"
	if got := translate("fr", source); got != source {
		t.Fatalf("translate() = %q, want English fallback %q", got, source)
	}
}

func TestFrenchCatalogIsComplete(t *testing.T) {
	if err := initializeTranslations(); err != nil {
		t.Fatal(err)
	}
	for source := range translationCatalogs[defaultLanguage] {
		if translated := translationCatalogs["fr"][source]; translated == "" {
			t.Errorf("French translation is missing for %q", source)
		}
	}
}

func TestTranslationPlaceholdersMatch(t *testing.T) {
	if err := initializeTranslations(); err != nil {
		t.Fatal(err)
	}
	placeholder := regexp.MustCompile(`%[sd]`)
	for source, translated := range translationCatalogs["fr"] {
		if !reflect.DeepEqual(placeholder.FindAllString(source, -1), placeholder.FindAllString(translated, -1)) {
			t.Errorf("format placeholders do not match for %q => %q", source, translated)
		}
	}
}

func TestCatalogContainsExtractedStrings(t *testing.T) {
	if err := initializeTranslations(); err != nil {
		t.Fatal(err)
	}
	templatePattern := regexp.MustCompile(`\{\{t\s+"([^"\\]+)"`)
	clientPattern := regexp.MustCompile(`window\.site(?:T|Format)\('([^'\\]+)'`)
	clientKeys := make(map[string]bool, len(clientTranslationKeys))
	for _, key := range clientTranslationKeys {
		clientKeys[key] = true
	}

	patterns := []string{"templates/*.html", "static/js/*.js"}
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			contents, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			for _, match := range templatePattern.FindAllSubmatch(contents, -1) {
				assertCatalogKey(t, file, string(match[1]))
			}
			for _, match := range clientPattern.FindAllSubmatch(contents, -1) {
				key := string(match[1])
				assertCatalogKey(t, file, key)
				if !clientKeys[key] {
					t.Errorf("client translation key %q from %s is missing from locales/client.json", key, file)
				}
			}
		}
	}
}

func assertCatalogKey(t *testing.T, file, key string) {
	t.Helper()
	for _, language := range supportedLanguages {
		if _, ok := translationCatalogs[language][key]; !ok {
			t.Errorf("%s catalog is missing %q extracted from %s", language, key, file)
		}
	}
}

func TestLanguageFromRequest(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		cookie string
		want   string
	}{
		{name: "default", url: "/", want: "en"},
		{name: "query", url: "/?lang=fr", want: "fr"},
		{name: "regional query", url: "/?lang=fr-CA", want: "fr"},
		{name: "cookie", url: "/", cookie: "fr", want: "fr"},
		{name: "query wins", url: "/?lang=en", cookie: "fr", want: "en"},
		{name: "invalid query does not hide cookie", url: "/?lang=de", cookie: "fr", want: "fr"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.url, nil)
			if test.cookie != "" {
				request.AddCookie(&http.Cookie{Name: languageCookie, Value: test.cookie})
			}
			if got := languageFromRequest(request); got != test.want {
				t.Fatalf("languageFromRequest() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestLanguageMiddlewareExposesLocale(t *testing.T) {
	handler := languageMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		localized, ok := writer.(interface{ Language() string })
		if !ok {
			t.Fatal("wrapped response writer does not expose its language")
		}
		_, _ = writer.Write([]byte(localized.Language()))
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/?lang=fr", nil))

	if recorder.Body.String() != "fr" {
		t.Fatalf("body = %q, want fr", recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Language"); got != "fr" {
		t.Fatalf("Content-Language = %q, want fr", got)
	}
	if len(recorder.Result().Cookies()) != 1 || recorder.Result().Cookies()[0].Value != "fr" {
		t.Fatal("French query selection was not persisted")
	}
}

func TestFrenchTemplateRendering(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := localizedResponseWriter{ResponseWriter: recorder, language: "fr"}
	page := &data.HomePage{
		Title:        "Brian R. Bondy",
		Description:  "Brian Bondy's writing about software, running, work, and life.",
		MarkdownSlug: "home",
		ShareUrl:     "/",
	}
	if err := executeTemplate(writer, "home", page); err != nil {
		t.Fatal(err)
	}

	html := recorder.Body.String()
	for _, expected := range []string{
		`<html lang="fr">`,
		`<h1 class="editorial-title">Course, travail et vie</h1>`,
		`id="language-toggle"`,
		`data-language="en"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("French HTML does not contain %q", expected)
		}
	}
}

func TestFrenchRenderingThroughRouter(t *testing.T) {
	router := mux.NewRouter()
	initializeRoutes(router)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://localhost:8080/?lang=fr", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	for _, expected := range []string{`<html lang="fr">`, `Course, travail et vie`, `data-language="en"`} {
		if !strings.Contains(recorder.Body.String(), expected) {
			t.Errorf("routed French HTML does not contain %q", expected)
		}
	}
}
