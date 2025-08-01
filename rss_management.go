package main

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string, client *http.Client) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var unmarshaled RSSFeed

	err = xml.Unmarshal(data, &unmarshaled)
	if err != nil {
		return nil, err
	}

	unmarshaled.Channel.Title = html.UnescapeString(unmarshaled.Channel.Title)
	unmarshaled.Channel.Description = html.UnescapeString(unmarshaled.Channel.Description)

	for i, item := range unmarshaled.Channel.Item {
		unmarshaled.Channel.Item[i].Title = html.UnescapeString(item.Title)
		unmarshaled.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}

	return &unmarshaled, nil
}
