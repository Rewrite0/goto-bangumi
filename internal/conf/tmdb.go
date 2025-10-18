// Package conf provides configuration constants and utilities
package conf

import "fmt"

const (
	// TMDBURL is the base URL for TMDB API
	TMDBURL = "https://api.themoviedb.org"

	// TMDBImgURL is the base URL for TMDB images
	TMDBImgURL = "https://image.tmdb.org/t/p/w780"

	// DefaultTMDBAPIKey is the default TMDB API key
	// Note: Using a public key is not recommended for production use.
	// It is better to set your own key in the configuration.
	DefaultTMDBAPIKey = "291237f90b24267380d6176c98f7619f"
)

// Language maps language codes to TMDB API language parameters
var Language = map[string]string{
	"zh": "zh-CN",
	"jp": "ja-JP",
	"en": "en-US",
}

// GetAPIKey returns the TMDB API key
// TODO: Get from settings when configuration system is implemented
func GetAPIKey() string {
	// For now, return the default key
	// In the future, this should read from settings.rss_parser.tmdb_api_key
	return DefaultTMDBAPIKey
}

// SearchURL generates TMDB search URL for TV shows
func SearchURL(keyword string) string {
	return fmt.Sprintf("%s/3/search/tv?api_key=%s&page=1&query=%s&include_adult=false",
		TMDBURL, GetAPIKey(), keyword)
}

// InfoURL generates TMDB info URL for a specific TV show
func InfoURL(showID string, language string) string {
	lang, ok := Language[language]
	if !ok {
		lang = Language["en"] // Default to English if language not found
	}
	return fmt.Sprintf("%s/3/tv/%s?api_key=%s&language=%s",
		TMDBURL, showID, GetAPIKey(), lang)
}
