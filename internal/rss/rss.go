// Package rss fetches and parses RSS feeds into torrent models.
package rss

import (
	"context"
	"encoding/xml"
	"fmt"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/utils"
)

// Getter provides the network dependency needed to fetch RSS content.
type Getter interface {
	Get(ctx context.Context, url string) ([]byte, error)
}

// Feed represents an RSS feed starting from channel level.
type Feed struct {
	Title string `xml:"channel>title"`
	Link  string `xml:"channel>link"`
	Items []Item `xml:"channel>item"`
}

// Item represents a single RSS item.
type Item struct {
	Name      string    `xml:"title"`
	Link      string    `xml:"link"`
	Enclosure Enclosure `xml:"enclosure"`
}

// Enclosure represents the RSS enclosure element used for torrent links.
type Enclosure struct {
	URL string `xml:"url,attr"`
}

// Parse parses RSS XML bytes into a Feed.
func Parse(data []byte) (*Feed, error) {
	var feed Feed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, &apperrors.ParseError{Err: fmt.Errorf("failed to parse RSS XML: %w", err)}
	}
	return &feed, nil
}

// Fetch fetches and parses an RSS feed.
func Fetch(ctx context.Context, getter Getter, url string) (*Feed, error) {
	// 需要能判断出来是网络不好还是空的 XML。
	// 空 RSS 示例: https://mikanani.me/RSS/Search?searchstr=ANININI
	resp, err := getter.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	return Parse(resp)
}

// ToTorrents converts RSS items into torrent models.
func ToTorrents(feed *Feed, fallbackURL string) []*model.Torrent {
	if feed == nil {
		return nil
	}

	torrents := make([]*model.Torrent, 0, len(feed.Items))
	for _, item := range feed.Items {
		name := utils.ProcessTitle(item.Name)
		torrent := &model.Torrent{
			Name:     name,
			Homepage: item.Enclosure.URL,
			Link:     fallbackURL,
		}

		if item.Enclosure.URL != "" {
			torrent.Link = item.Enclosure.URL
			torrent.Homepage = item.Link
		} else {
			torrent.Link = item.Link
		}

		torrents = append(torrents, torrent)
	}

	return torrents
}

// GetTorrents fetches and parses RSS feed to extract torrents.
// 返回错误主要是为了区分网络请求错误和确实没有种子。
func GetTorrents(ctx context.Context, getter Getter, url string) ([]*model.Torrent, error) {
	feed, err := Fetch(ctx, getter, url)
	if err != nil {
		return nil, err
	}
	return ToTorrents(feed, url), nil
}

// GetTitle fetches RSS feed and returns the channel title.
func GetTitle(ctx context.Context, getter Getter, url string) (string, error) {
	feed, err := Fetch(ctx, getter, url)
	if err != nil {
		return "", err
	}
	return feed.Title, nil
}
