package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

type publishedBlogAudio struct {
	AudioURL     string `json:"audioURL"`
	TimingSHA256 string `json:"timingSHA256"`
}

var publishedBlogAudioByID map[int]publishedBlogAudio
var sha256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func loadBlogAudioManifest() error {
	contents, err := os.ReadFile("data/blogAudioManifest.json")
	if err != nil {
		return fmt.Errorf("read blog audio manifest: %w", err)
	}
	var entries map[string]publishedBlogAudio
	if err := json.Unmarshal(contents, &entries); err != nil {
		return fmt.Errorf("parse blog audio manifest: %w", err)
	}
	result := make(map[int]publishedBlogAudio, len(entries))
	for key, entry := range entries {
		id, err := strconv.Atoi(key)
		if err != nil || id <= 0 {
			return fmt.Errorf("invalid blog audio post ID %q", key)
		}
		parsed, err := url.Parse(entry.AudioURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host != "storage.googleapis.com" || parsed.Path == "" {
			return fmt.Errorf("invalid blog audio URL for post %d", id)
		}
		if !sha256Pattern.MatchString(entry.TimingSHA256) {
			return fmt.Errorf("invalid blog audio timing hash for post %d", id)
		}
		result[id] = entry
	}
	publishedBlogAudioByID = result
	return nil
}
