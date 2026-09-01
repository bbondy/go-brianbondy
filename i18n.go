package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	defaultLanguage = "en"
	languageCookie  = "site_language"
)

var supportedLanguages = []string{"en", "fr"}

//go:embed locales/*.json
var localeFiles embed.FS

var (
	translationCatalogs     map[string]map[string]string
	clientTranslationKeys   []string
	translationCatalogsOnce sync.Once
	translationCatalogsErr  error
)

func initializeTranslations() error {
	translationCatalogsOnce.Do(func() {
		translationCatalogs = make(map[string]map[string]string, len(supportedLanguages))
		for _, language := range supportedLanguages {
			contents, err := localeFiles.ReadFile("locales/" + language + ".json")
			if err != nil {
				translationCatalogsErr = fmt.Errorf("read %s translations: %w", language, err)
				return
			}

			catalog := make(map[string]string)
			if err := json.Unmarshal(contents, &catalog); err != nil {
				translationCatalogsErr = fmt.Errorf("parse %s translations: %w", language, err)
				return
			}
			translationCatalogs[language] = catalog
		}
		contents, err := localeFiles.ReadFile("locales/client.json")
		if err != nil {
			translationCatalogsErr = fmt.Errorf("read client translation keys: %w", err)
			return
		}
		if err := json.Unmarshal(contents, &clientTranslationKeys); err != nil {
			translationCatalogsErr = fmt.Errorf("parse client translation keys: %w", err)
		}
	})
	return translationCatalogsErr
}

func isSupportedLanguage(language string) bool {
	for _, supported := range supportedLanguages {
		if language == supported {
			return true
		}
	}
	return false
}

func parseLanguage(language string) (string, bool) {
	language = strings.ToLower(strings.TrimSpace(language))
	if separator := strings.IndexAny(language, "-_"); separator >= 0 {
		language = language[:separator]
	}
	if isSupportedLanguage(language) {
		return language, true
	}
	return "", false
}

func normalizeLanguage(language string) string {
	if normalized, ok := parseLanguage(language); ok {
		return normalized
	}
	return defaultLanguage
}

func languageFromWriter(writer interface{}) string {
	for depth := 0; depth < 8 && writer != nil; depth++ {
		if localized, ok := writer.(interface{ Language() string }); ok {
			if language, valid := parseLanguage(localized.Language()); valid {
				return language
			}
		}

		value := reflect.ValueOf(writer)
		for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr) {
			if value.IsNil() {
				return defaultLanguage
			}
			value = value.Elem()
		}
		if !value.IsValid() || value.Kind() != reflect.Struct {
			return defaultLanguage
		}
		underlying := value.FieldByName("ResponseWriter")
		if !underlying.IsValid() || !underlying.CanInterface() {
			return defaultLanguage
		}
		writer = underlying.Interface()
	}
	return defaultLanguage
}

func translate(language, english string, args ...interface{}) string {
	language = normalizeLanguage(language)
	translated := english
	if initializeTranslations() == nil {
		if value := translationCatalogs[language][english]; value != "" {
			translated = value
		} else if value := translationCatalogs[defaultLanguage][english]; value != "" {
			translated = value
		}
	}
	if len(args) > 0 {
		return fmt.Sprintf(translated, args...)
	}
	return translated
}

func translationsJSON(language string) template.JS {
	merged := make(map[string]string)
	if initializeTranslations() == nil {
		for _, key := range clientTranslationKeys {
			merged[key] = translate(language, key)
		}
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		return template.JS(`{}`)
	}
	return template.JS(encoded)
}

func localizedFuncMap(language string) template.FuncMap {
	localized := template.FuncMap{}
	for name, function := range funcMap {
		localized[name] = function
	}
	language = normalizeLanguage(language)
	localized["lang"] = func() string { return language }
	localized["alternateLang"] = func() string {
		if language == "fr" {
			return "en"
		}
		return "fr"
	}
	localized["t"] = func(english string, args ...interface{}) string {
		return translate(language, english, args...)
	}
	localized["pageTitle"] = func(title string) string {
		const suffix = " - Brian R. Bondy"
		if strings.HasSuffix(title, suffix) {
			return translate(language, strings.TrimSuffix(title, suffix)) + suffix
		}
		return translate(language, title)
	}
	localized["translationsJSON"] = func() template.JS { return translationsJSON(language) }
	localized["formatDate"] = func(dateStr string) string { return formatDateForLanguage(dateStr, language) }
	localized["formatFullDate"] = func(dateStr string) string { return formatDateForLanguage(dateStr, language) }
	return localized
}

func formatDateForLanguage(dateStr, language string) string {
	var parsed time.Time
	var err error
	for _, layout := range []string{"2006-01-02", "January 2, 2006"} {
		parsed, err = time.Parse(layout, dateStr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return dateStr
	}
	if normalizeLanguage(language) == "fr" {
		months := [...]string{"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"}
		return fmt.Sprintf("%d %s %d", parsed.Day(), months[parsed.Month()-1], parsed.Year())
	}
	return parsed.Format("January 2, 2006")
}

type localizedResponseWriter struct {
	http.ResponseWriter
	language string
}

func (writer localizedResponseWriter) Language() string { return writer.language }

func languageFromRequest(request *http.Request) string {
	if requested, ok := parseLanguage(request.URL.Query().Get("lang")); ok {
		return requested
	}
	if cookie, err := request.Cookie(languageCookie); err == nil {
		if requested, ok := parseLanguage(cookie.Value); ok {
			return requested
		}
	}
	return defaultLanguage
}

func languageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/static/") ||
			strings.HasPrefix(request.URL.Path, "/api/") ||
			request.URL.Path == "/rss" ||
			request.URL.Path == "/sitemap.xml" ||
			request.URL.Path == "/robots.txt" {
			next.ServeHTTP(writer, request)
			return
		}
		language := languageFromRequest(request)
		if requested, ok := parseLanguage(request.URL.Query().Get("lang")); ok && requested == language {
			http.SetCookie(writer, &http.Cookie{
				Name:     languageCookie,
				Value:    language,
				Path:     "/",
				MaxAge:   60 * 60 * 24 * 365,
				SameSite: http.SameSiteLaxMode,
			})
		}
		writer.Header().Set("Content-Language", language)
		writer.Header().Add("Vary", "Cookie")
		next.ServeHTTP(localizedResponseWriter{ResponseWriter: writer, language: language}, request)
	})
}
