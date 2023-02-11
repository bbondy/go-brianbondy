package data

import (
	"encoding/xml"
	"strconv"
)

type Item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Link        string   `xml:"link"`
	PubDate     string   `xml:"pubDate"`
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Title   string   `xml:"title"`
	Link    string   `xml:"link"`
	Items   []Item   `xml:"item"`
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

func ConvertToRSS(posts BlogPosts, title, link string) ([]byte, error) {
	var items []Item

	for _, post := range posts {
		items = append(items, Item{
			Title:       post.Title,
			Description: "",
			Link:        link + "/blog/" + strconv.Itoa(post.Id),
			PubDate:     post.Created,
		})
	}

	rss := RSS{
		Version: "2.0",
		Channel: Channel{
			Title: title,
			Link:  link,
			Items: items,
		},
	}

	return xml.MarshalIndent(rss, "", "  ")
}
