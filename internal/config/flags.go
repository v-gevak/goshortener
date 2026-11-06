package config

import "flag"

func ParseFlags(cfg *Config) {
	flag.StringVar(&cfg.ServiceSettings.BaseURL, "b", "http://localhost:8080", "base URL for shortened links (scheme://host:port)")
	flag.StringVar(&cfg.ServerSettings.Address, "a", "localhost:8080", "server address (host:port)")

	flag.Parse()
}