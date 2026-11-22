// Usage: go run scripts/blog_audio_source.go 186
// Prints the rendered prose blocks used by the audio generator.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"golang.org/x/net/html"
)

func textOf(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}
	var text strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		text.WriteString(textOf(child))
	}
	return text.String()
}

func isProseBlock(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}
	switch node.Data {
	case "p", "h1", "h2", "h3", "h4", "h5", "h6":
		return true
	case "li":
		// Markdown sometimes wraps list item text in paragraphs. In that case,
		// the paragraphs are the narration blocks, avoiding duplicate text.
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if containsProseBlock(child) {
				return false
			}
		}
		return true
	}
	return false
}

func containsProseBlock(node *html.Node) bool {
	if node.Type == html.ElementNode {
		switch node.Data {
		case "p", "h1", "h2", "h3", "h4", "h5", "h6":
			return true
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if containsProseBlock(child) {
			return true
		}
	}
	return false
}

func blocks(node *html.Node, result *[]string) {
	if isProseBlock(node) {
		value := strings.Join(strings.Fields(textOf(node)), " ")
		if value != "" {
			*result = append(*result, value)
		}
		return
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		blocks(child, result)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/blog_audio_source.go POST_ID")
		os.Exit(2)
	}
	id, err := strconv.Atoi(os.Args[1])
	if err != nil || id <= 0 {
		fmt.Fprintln(os.Stderr, "POST_ID must be a positive integer")
		os.Exit(2)
	}
	input, err := os.ReadFile(fmt.Sprintf("data/markdown/blog/%d.markdown", id))
	if err != nil {
		panic(err)
	}
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
	markup := markdown.ToHTML(input, p, nil)
	root, err := html.Parse(strings.NewReader(string(markup)))
	if err != nil {
		panic(err)
	}
	var result []string
	blocks(root, &result)
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
