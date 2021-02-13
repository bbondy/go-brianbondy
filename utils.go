package main

import (
	"strings"
	"regexp"
)

func SlugifyTitle(title string) string {
	title = strings.ToLower(title)
	str := strings.ReplaceAll(title, " ", "-")
	reg, _ := regexp.Compile("[^a-zA-Z0-9-]+")
	str = reg.ReplaceAllString(str, "")
	return strings.ReplaceAll(str, "--", "-")
}

func GetTitle(titleSlug string) string {
	return titleSlug + " - " + "Brian R. Bondy"
}
