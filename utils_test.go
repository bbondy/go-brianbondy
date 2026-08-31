package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bbondy/go-brianbondy/data"
)

func TestSlugifyTitle(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"This is a test!", "this-is-a-test"},
		{"Another_Test 123", "another-test-123"},
	}

	for _, tt := range tests {
		got := slugifyTitle(tt.title)
		if got != tt.want {
			t.Errorf("slugifyTitle(%q) = %v, want %v", tt.title, got, tt.want)
		}
	}
}

func TestGetTitle(t *testing.T) {
	titleSlug := "test-slug"
	want := "test-slug - Brian R. Bondy"
	got := GetTitle(titleSlug)
	if got != want {
		t.Errorf("GetTitle(%q) = %v, want %v", titleSlug, got, want)
	}
}

func TestDirectToHttps(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		proto          string
		host           string
		forwardedProto string
		wantRedirect   bool
		wantLocation   string
	}{
		{"http to canonical https", "http://example.com/test?x=1", "HTTP/1.1", "example.com", "", true, "https://brianbondy.com/test?x=1"},
		{"www redirects to apex", "https://www.brianbondy.com", "HTTP/2.0", "www.brianbondy.com", "", true, "https://brianbondy.com/"},
		{"canonical https", "https://brianbondy.com", "HTTP/2.0", "brianbondy.com", "", false, ""},
		{"localhost no redirect", "http://localhost:8080", "HTTP/1.1", "localhost:8080", "", false, ""},
		{"forwarded canonical https", "http://brianbondy.com", "HTTP/1.1", "brianbondy.com", "https", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			req.Proto = tt.proto
			req.Host = tt.host
			req.Header.Set("X-Forwarded-Proto", tt.forwardedProto)

			w := httptest.NewRecorder()
			directToHttps(w, req, func(w http.ResponseWriter, r *http.Request) {})

			if tt.wantRedirect {
				if w.Code != http.StatusPermanentRedirect {
					t.Errorf("expected status %v, got %v", http.StatusPermanentRedirect, w.Code)
				}
				if got := w.Header().Get("Location"); got != tt.wantLocation {
					t.Errorf("expected location %q, got %q", tt.wantLocation, got)
				}
			} else {
				if w.Code != http.StatusOK {
					t.Errorf("expected status %v, got %v", http.StatusOK, w.Code)
				}
			}
		})
	}
}

func TestCanonicalURL(t *testing.T) {
	if got := canonicalURL(&data.SimpleMarkdownPage{MarkdownSlug: "about.markdown"}); got != "https://brianbondy.com/about" {
		t.Errorf("canonicalURL() = %q, want about URL", got)
	}
	if got := canonicalURL(&data.BlogPostPage{ShareUrl: "/blog/1/test-post"}); got != "https://brianbondy.com/blog/1/test-post" {
		t.Errorf("canonicalURL() = %q, want blog URL", got)
	}
	if got := canonicalURL(&data.CareerPage{MarkdownSlug: "resume"}); got != "https://brianbondy.com/resume" {
		t.Errorf("canonicalURL() = %q, want resume URL", got)
	}
	if got := canonicalURL(&data.CareerPage{MarkdownSlug: "career"}); got != "https://brianbondy.com/career" {
		t.Errorf("canonicalURL() = %q, want career URL", got)
	}
}

func TestExtractFirstParagraph(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			"p tags",
			"<p>This is a test paragraph.</p><p>Another paragraph.</p>",
			"This is a test paragraph.",
		},
		{
			"no p tags",
			"This is a test paragraph. Another paragraph.",
			"This is a test paragraph. Another paragraph.",
		},
		{
			"long p tag",
			"<p>" + strings.Repeat("A", 400) + "</p>",
			strings.Repeat("A", 300) + "...",
		},
		{
			"long content no p tags",
			strings.Repeat("A", 400),
			strings.Repeat("A", 300) + "...",
		},
		{
			"empty p tag",
			"<p></p>Some content",
			"Some content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFirstParagraph(tt.content)
			if got != tt.want {
				t.Errorf("extractFirstParagraph() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetImageMimeType(t *testing.T) {
	tests := []struct {
		imagePath string
		want      string
	}{
		{"image.png", "image/png"},
		{"image.webp", "image/webp"},
		{"image.jpg", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"image.gif", "image/gif"},
		{"image.bmp", ""},
	}

	for _, tt := range tests {
		got := getImageMimeType(tt.imagePath)
		if got != tt.want {
			t.Errorf("getImageMimeType(%q) = %v, want %v", tt.imagePath, got, tt.want)
		}
	}
}

func TestOptimizeImageTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"basic image optimization",
			`<img src="/static/img/test.jpg" alt="test">`,
			`<img src="/static/img/test.jpg" alt="test" loading="lazy" data-lightbox-src="/static/img/test.jpg" decoding="async">`,
		},
		{
			"image with existing loading attribute",
			`<img src="/static/img/test.png" alt="test" loading="lazy">`,
			`<img src="/static/img/test.png" alt="test" loading="lazy" data-lightbox-src="/static/img/test.png" decoding="async">`,
		},
		{
			"external image - no optimization",
			`<img src="https://example.com/image.jpg" alt="test">`,
			`<img src="https://example.com/image.jpg" alt="test">`,
		},
		{
			"data URL - no optimization",
			`<img src="data:image/png;base64,test" alt="test">`,
			`<img src="data:image/png;base64,test" alt="test">`,
		},
		{
			"non-static image - no optimization",
			`<img src="/other/image.jpg" alt="test">`,
			`<img src="/other/image.jpg" alt="test">`,
		},
		{
			"image with responsive variants",
			`<img src="/static/img/blogpost_190/tahoe-start.webp" alt="test">`,
			`<img src="/static/img/blogpost_190/tahoe-start.webp" alt="test" loading="lazy" data-lightbox-src="/static/img/blogpost_190/tahoe-start.webp" srcset="/static/img/blogpost_190/tahoe-start-640.webp 640w, /static/img/blogpost_190/tahoe-start-960.webp 960w, /static/img/blogpost_190/tahoe-start-1200.webp 1200w" sizes="(max-width: 800px) calc(100vw - 44px), 756px" decoding="async">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optimizeImageTag(tt.input)
			if got != tt.expected {
				t.Errorf("optimizeImageTag() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOptimizeImagesInContent(t *testing.T) {
	input := `<p>Some text</p><img src="/static/img/test.jpg" alt="test"><p>More text</p><img src="/static/img/test2.png" alt="test2">`
	expected := `<p>Some text</p><img src="/static/img/test.jpg" alt="test" loading="lazy" data-lightbox-src="/static/img/test.jpg" decoding="async"><p>More text</p><img src="/static/img/test2.png" alt="test2" loading="lazy" data-lightbox-src="/static/img/test2.png" decoding="async">`

	got := optimizeImagesInContent(input)
	if got != expected {
		t.Errorf("optimizeImagesInContent() = %v, want %v", got, expected)
	}
}

func TestResponsiveImageSrcset(t *testing.T) {
	got := responsiveImageSrcset("/static/img/blogpost_190/tahoe-start.webp")
	want := "/static/img/blogpost_190/tahoe-start-640.webp 640w, /static/img/blogpost_190/tahoe-start-960.webp 960w, /static/img/blogpost_190/tahoe-start-1200.webp 1200w"
	if got != want {
		t.Errorf("responsiveImageSrcset() = %q, want %q", got, want)
	}
}

func TestResponsiveImageSizesFor(t *testing.T) {
	if got := responsiveImageSizesFor(); got != responsiveImageSizes {
		t.Errorf("responsiveImageSizesFor() = %q, want %q", got, responsiveImageSizes)
	}
	if got := responsiveImageSizesFor(""); got != responsiveImageSizes {
		t.Errorf("responsiveImageSizesFor(\"\") = %q, want %q", got, responsiveImageSizes)
	}
	if got := responsiveImageSizesFor("160px"); got != "160px" {
		t.Errorf("responsiveImageSizesFor(\"160px\") = %q, want %q", got, "160px")
	}
}
