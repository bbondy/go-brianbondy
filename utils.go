package main

import (
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

func slugifyTitle(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace any non-alphanumeric characters (except hyphens) with a hyphen
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

func GetTitle(titleSlug string) string {
	return titleSlug + " - " + "Brian R. Bondy"
}

func directToHttps(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Host == "localhost:8080" ||
		r.URL.Scheme == "https" ||
		strings.HasPrefix(r.Proto, "HTTPS") ||
		r.Header.Get("X-Forwarded-Proto") == "https" {
		next(w, r)
	} else {
		target := "https://" + r.Host + r.URL.Path
		http.Redirect(w, r, target,
			http.StatusTemporaryRedirect)
	}
}

func extractFirstParagraph(content string) string {
	// First try to find content between first <p> tags
	re := regexp.MustCompile(`<p>(.*?)</p>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		// Get the content inside the first <p> tags
		preview := matches[1]

		// Remove any remaining HTML tags
		tagRegex := regexp.MustCompile(`<[^>]*>`)
		preview = tagRegex.ReplaceAllString(preview, "")

		if preview != "" {
			// If preview is too long, truncate it
			if len(preview) > 300 {
				// Try to cut at a word boundary
				lastSpace := strings.LastIndex(preview[:300], " ")
				if lastSpace > 0 {
					preview = preview[:lastSpace] + "..."
				} else {
					preview = preview[:300] + "..."
				}
			}
			return preview
		}
	}

	// If no paragraph found or it was empty, remove all HTML tags from content
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	cleanContent := tagRegex.ReplaceAllString(content, " ")

	// Remove extra whitespace
	cleanContent = strings.Join(strings.Fields(cleanContent), " ")

	// Take first 300 characters or less
	if len(cleanContent) > 300 {
		lastSpace := strings.LastIndex(cleanContent[:300], " ")
		if lastSpace > 0 {
			cleanContent = cleanContent[:lastSpace] + "..."
		} else {
			cleanContent = cleanContent[:300] + "..."
		}
	}

	return cleanContent
}

func getImageMimeType(imagePath string) string {
	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	default:
		return ""
	}
}
