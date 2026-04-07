package model

// Torrent represents a torrent item from RSS feed
// type Torrent struct {
// 	Name     string
// 	Link     string
// 	Homepage string
// }

// ProxyConfig represents proxy configuration
type ProxyConfig struct {
	Enable   bool   `yaml:"enable" env:"ENABLE" env-default:"false"`
	Type     string `yaml:"type" env:"TYPE" env-default:"http"`
	Host     string `yaml:"host" env:"HOST"`
	Port     int    `yaml:"port" env:"PORT" env-default:"0"`
	Username string `yaml:"username" env:"USERNAME"`
	Password string `yaml:"password" env:"PASSWORD"`
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
