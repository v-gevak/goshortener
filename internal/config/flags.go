package config

import "flag"

func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.ServiceSettings.BaseURL, "b", cfg.ServiceSettings.BaseURL, "base URL for shortened links (scheme://host:port)")
	flag.StringVar(&cfg.ServerSettings.Address, "a", cfg.ServerSettings.Address, "server address (host:port)")
	flag.StringVar(&cfg.LogSettings.Level, "l", cfg.LogSettings.Level, "log level (debug, info, warn, error)")
	flag.StringVar(&cfg.StoreSettings.Provider, "p", cfg.StoreSettings.Provider, "store provider (file, in-memory)")
	flag.StringVar(&cfg.StoreSettings.FilePath, "f", cfg.StoreSettings.FilePath, "path to storage file")

	flag.Parse()
}
