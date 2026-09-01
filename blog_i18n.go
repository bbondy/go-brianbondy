package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bbondy/go-brianbondy/data"
)

type blogPostTranslation struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

var blogPostTranslations = make(map[string]map[int]blogPostTranslation)

// initializeBlogPostTranslations loads optional per-language post metadata.
// Missing files, posts, and fields intentionally fall back to English.
func initializeBlogPostTranslations() error {
	for _, language := range supportedLanguages {
		if language == defaultLanguage {
			continue
		}
		contents, err := os.ReadFile("data/blogPostTranslations." + language + ".json")
		if os.IsNotExist(err) {
			blogPostTranslations[language] = make(map[int]blogPostTranslation)
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s blog post translations: %w", language, err)
		}
		translations := make(map[int]blogPostTranslation)
		if err := json.Unmarshal(contents, &translations); err != nil {
			return fmt.Errorf("parse %s blog post translations: %w", language, err)
		}
		blogPostTranslations[language] = translations
	}
	return nil
}

func localizedBlogPost(post data.BlogPost, language string) data.BlogPost {
	language = normalizeLanguage(language)
	translation, ok := blogPostTranslations[language][post.Id]
	if !ok {
		return post
	}
	if translation.Title != "" {
		post.Title = translation.Title
	}
	if translation.Description != "" {
		description := translation.Description
		post.Description = &description
	}
	return post
}

func localizedBlogPostPointer(post *data.BlogPost, language string) *data.BlogPost {
	if post == nil {
		return nil
	}
	localized := localizedBlogPost(*post, language)
	return &localized
}

func blogPostURL(post data.BlogPost) string {
	title := post.Title
	if source, ok := blogPostIdMap[post.Id]; ok {
		title = source.Title
	}
	return "/blog/" + strconv.Itoa(post.Id) + "/" + slugifyTitle(title)
}

func localizedMarkdownSlug(slug, language string) string {
	language = normalizeLanguage(language)
	if language == defaultLanguage || !strings.HasPrefix(slug, "blog/") || !strings.HasSuffix(slug, ".markdown") {
		return slug
	}
	translated := strings.TrimSuffix(slug, ".markdown") + "." + language + ".markdown"
	if _, err := os.Stat("data/markdown/" + translated); err == nil {
		return translated
	}
	return slug
}

func getLocalizedMarkdownData(slug, language string) string {
	return getMarkdownData(localizedMarkdownSlug(slug, language))
}
