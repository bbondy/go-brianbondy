package main

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (w gzipResponseWriter) Write(data []byte) (int, error) { return w.writer.Write(data) }

// gzipHTML compresses dynamic HTML when the front proxy has not already done so.
func gzipHTML(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		next(w, r)
		return
	}
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Add("Vary", "Accept-Encoding")
	gz := gzip.NewWriter(w)
	defer func() {
		_ = gz.Close()
	}()
	next(gzipResponseWriter{ResponseWriter: w, writer: gz}, r)
}

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
	if r.Host == "localhost:8080" {
		next(w, r)
		return
	}

	isHTTPS := r.URL.Scheme == "https" || strings.HasPrefix(r.Proto, "HTTPS") || r.Header.Get("X-Forwarded-Proto") == "https"
	if isHTTPS && r.Host == "brianbondy.com" {
		next(w, r)
		return
	}

	// Use a fixed hostname rather than the request Host header so every public
	// URL resolves to the site's single canonical origin.
	http.Redirect(w, r, siteURL+r.URL.RequestURI(), http.StatusPermanentRedirect)
}

// canonicalURL returns the preferred public URL for a rendered page.
func canonicalURL(data interface{}) string {
	pathBySlug := map[string]string{
		"home":             "/",
		"all":              "/all",
		"filters":          "/blog/filters",
		"running":          "/running",
		"pictures":         "/pictures",
		"projects":         "/projects",
		"interviews":       "/interviews",
		"cheatsheets":      "/cheatsheets",
		"about.markdown":   "/about",
		"advice.markdown":  "/advice",
		"contact.markdown": "/contact",
		"resume.markdown":  "/resume",
		"books":            "/books",
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Map {
		if page := v.MapIndex(reflect.ValueOf("Page")); page.IsValid() {
			return canonicalURL(page.Interface())
		}
	}
	if v.Kind() != reflect.Struct {
		return ""
	}
	if shareURL := v.FieldByName("ShareUrl"); shareURL.IsValid() && shareURL.Kind() == reflect.String && shareURL.String() != "" {
		return siteURL + shareURL.String()
	}
	if slug := v.FieldByName("MarkdownSlug"); slug.IsValid() && slug.Kind() == reflect.String {
		return siteURL + pathBySlug[slug.String()]
	}
	return ""
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

const responsiveImageSizes = "(max-width: 800px) calc(100vw - 44px), 756px"

// responsiveImageSizesFor returns the sizes attribute for an image slot. It
// defaults to the article body width; templates rendering smaller slots (such
// as gallery thumbnails) pass their own value so browsers pick a smaller
// srcset candidate.
func responsiveImageSizesFor(override ...string) string {
	if len(override) > 0 && override[0] != "" {
		return override[0]
	}
	return responsiveImageSizes
}

func responsiveImageSrcset(src string) string {
	ext := strings.ToLower(filepath.Ext(src))
	if !strings.HasPrefix(src, "/static/") || (ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp") {
		return ""
	}

	base := strings.TrimSuffix(src, filepath.Ext(src))
	variants := make([]string, 0, 3)
	for _, width := range []int{640, 960, 1200} {
		variant := fmt.Sprintf("%s-%d.webp", base, width)
		if _, err := os.Stat(strings.TrimPrefix(variant, "/")); err == nil {
			variants = append(variants, fmt.Sprintf("%s %dw", variant, width))
		}
	}
	return strings.Join(variants, ", ")
}

// Image optimization functions for PageSpeed improvements.
func optimizeImageTag(imgTag string) string {
	// Extract src attribute
	srcRegex := regexp.MustCompile(`src=["']([^"']+)["']`)
	srcMatch := srcRegex.FindStringSubmatch(imgTag)
	if len(srcMatch) < 2 {
		return imgTag
	}

	src := srcMatch[1]

	// Skip data URLs and external URLs
	if strings.HasPrefix(src, "data:") || strings.HasPrefix(src, "http") {
		return imgTag
	}

	// Only process static images
	if !strings.HasPrefix(src, "/static/") {
		return imgTag
	}

	if strings.ToLower(filepath.Ext(src)) == ".gif" {
		return imgTag
	}

	responsiveImg := imgTag

	// Add loading="lazy" if not already present
	if !strings.Contains(responsiveImg, `loading="`) {
		responsiveImg = strings.Replace(responsiveImg, ">", ` loading="lazy">`, 1)
	}

	if !strings.Contains(responsiveImg, "data-lightbox-src=") {
		responsiveImg = strings.Replace(responsiveImg, ">", ` data-lightbox-src="`+src+`">`, 1)
	}

	if srcset := responsiveImageSrcset(src); srcset != "" && !strings.Contains(responsiveImg, "srcset=") {
		responsiveImg = strings.Replace(responsiveImg, ">", ` srcset="`+srcset+`" sizes="`+responsiveImageSizes+`">`, 1)
	}

	// Add decoding="async" for better performance
	if !strings.Contains(responsiveImg, `decoding="`) {
		responsiveImg = strings.Replace(responsiveImg, ">", ` decoding="async">`, 1)
	}

	return responsiveImg
}

func optimizeImagesInContent(content string) string {
	// Find all img tags
	imgRegex := regexp.MustCompile(`<img[^>]+>`)
	return imgRegex.ReplaceAllStringFunc(content, func(match string) string {
		return optimizeImageTag(match)
	})
}
