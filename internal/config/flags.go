package config

import "flag"

func ParseFlags(cfg *Config) {
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "server address (host:port)")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base URL for shortened links (scheme://host:port)")

	flag.Parse()
}
