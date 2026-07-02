package model

// ProxyConfig represents proxy configuration
type ProxyConfig struct {
	Enable   bool   `toml:"enable" env:"ENABLE" env-default:"false"`
	Type     string `toml:"type" env:"TYPE" env-default:"http"`
	Host     string `toml:"host" env:"HOST"`
	Port     int    `toml:"port" env:"PORT" env-default:"0"`
	Username string `toml:"username" env:"USERNAME"`
	Password string `toml:"password" env:"PASSWORD"`
}
