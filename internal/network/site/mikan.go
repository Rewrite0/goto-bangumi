// Package site provides RSS parsing functionality for different sites
package site

import (
	"goto-bangumi/internal/model"
)

// RSSParser parses RSS feed and extracts torrent information
// Returns three slices: titles, URLs, and homepages
func RSSParser(rss model.RSSXml) ([]string, []string, []string) {
	torrentTitles := make([]string, 0, len(rss.Torrents))
	torrentURLs := make([]string, 0, len(rss.Torrents))
	torrentHomepages := make([]string, 0, len(rss.Torrents))

	for _, item := range rss.Torrents {
		// Add title
		torrentTitles = append(torrentTitles, item.Name)

		// Check if torrent URL exists
		if item.Enclosure.URL != "" {
			torrentURLs = append(torrentURLs, item.Enclosure.URL)
		} else {
			// No torrent URL, use link instead
			torrentURLs = append(torrentURLs, item.Link)
		}
		torrentHomepages = append(torrentHomepages, item.Homepage)
	}

	return torrentTitles, torrentURLs, torrentHomepages
}

// MikanTitle extracts title from RSS feed (for compatibility)
func MikanTitle(rss model.RSSXml) string {
	return rss.Title
}
