package main

import (
	"strings"
	"testing"

	"github.com/bbondy/go-brianbondy/data"
)

func TestPublishedBlogAudio(t *testing.T) {
	previous := publishedBlogAudioByID
	t.Cleanup(func() { publishedBlogAudioByID = previous })
	hash := strings.Repeat("a", 64)
	publishedBlogAudioByID = map[int]publishedBlogAudio{
		173: {
			AudioURL:     "https://storage.googleapis.com/example/blog-audio/173-test.mp3",
			TimingSHA256: hash,
		},
	}

	english := &data.BlogPostPage{}
	setBlogAudio(english, 173, "en")
	if english.AudioURL != publishedBlogAudioByID[173].AudioURL {
		t.Fatalf("published audio URL = %q", english.AudioURL)
	}
	if english.AudioTimingURL != "/static/audio/blog/173.json?v="+hash[:12] {
		t.Fatalf("timing URL = %q", english.AudioTimingURL)
	}

	french := &data.BlogPostPage{}
	setBlogAudio(french, 173, "fr")
	if french.AudioURL != "" || french.AudioTimingURL != "" {
		t.Fatal("English narration should not appear on a French post")
	}
}

func TestBlogAudioManifest(t *testing.T) {
	if err := loadBlogAudioManifest(); err != nil {
		t.Fatal(err)
	}
	if len(publishedBlogAudioByID) == 0 {
		t.Fatal("expected at least one published audio entry")
	}
}
