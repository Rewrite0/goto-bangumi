package model

// Torrent represents a torrent item from RSS feed
// type Torrent struct {
// 	Name     string
// 	Link     string
// 	Homepage string
// }

// ProxyConfig represents proxy configuration
type ProxyConfig struct {
	Enable   bool   `json:"enable" mapstructure:"enable"`
	Type     string `json:"type" mapstructure:"type"` // http, https (socks5 not supported yet)
	Host     string `json:"host" mapstructure:"host"`
	Port     int    `json:"port" mapstructure:"port"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
}

// RSSXml represents RSS feed starting from channel level
type RSSXml struct {
	Title    string       `xml:"channel>title"`
	Link     string       `xml:"channel>link"`
	Torrents []RSSTorrent `xml:"channel>item"`
}

// RSSTorrent represents a single torrent item
type RSSTorrent struct {
	Name string `xml:"title"`
	Link string `xml:"link"`
	// Homepage string `xml:"guid"`
	Enclosure Enclosure `xml:"enclosure"`
	// Homepage struct {
	// 	URL string `xml:"url,attr"`
	// } `xml:"enclosure"`
}

type Enclosure struct {
	URL string `xml:"url,attr"`
}
