package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type FeedItem struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	PubDate     time.Time `json:"pubDate"`
	Source      string    `json:"source"`
}

type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title string    `xml:"title"`
		Items []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type AtomFeed struct {
	XMLName xml.Name `xml:"feed"`
	Title   string   `xml:"title"`
	Entries []struct {
		Title   string `xml:"title"`
		Link    string `xml:"link"`
		Content string `xml:"content"`
		Summary string `xml:"summary"`
		Updated string `xml:"updated"`
	} `xml:"entry"`
}

func FetchFeed(ctx context.Context, url string) ([]FeedItem, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "rss-digest/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Try RSS first
	var rss RSSFeed
	if err := xml.Unmarshal(body, &rss); err == nil && len(rss.Channel.Items) > 0 {
		return parseRSS(rss), nil
	}

	// Try Atom
	var atom AtomFeed
	if err := xml.Unmarshal(body, &atom); err == nil && len(atom.Entries) > 0 {
		return parseAtom(atom), nil
	}

	return nil, fmt.Errorf("could not parse feed")
}

func parseRSS(feed RSSFeed) []FeedItem {
	var items []FeedItem
	for _, item := range feed.Channel.Items {
		pubDate, _ := parseDate(item.PubDate)
		items = append(items, FeedItem{
			Title:       strings.TrimSpace(item.Title),
			Link:        strings.TrimSpace(item.Link),
			Description: cleanHTML(strings.TrimSpace(item.Description)),
			PubDate:     pubDate,
			Source:      feed.Channel.Title,
		})
	}
	return items
}

func parseAtom(feed AtomFeed) []FeedItem {
	var items []FeedItem
	for _, entry := range feed.Entries {
		content := entry.Content
		if content == "" {
			content = entry.Summary
		}
		pubDate, _ := parseDate(entry.Updated)
		items = append(items, FeedItem{
			Title:       strings.TrimSpace(entry.Title),
			Link:        strings.TrimSpace(entry.Link),
			Description: cleanHTML(strings.TrimSpace(content)),
			PubDate:     pubDate,
			Source:      feed.Title,
		})
	}
	return items
}

func parseDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date: %s", s)
}

func cleanHTML(s string) string {
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = strings.ReplaceAll(s, "</p>", "\n")
	// Basic tag stripping
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}
