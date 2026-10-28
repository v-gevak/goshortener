package config

import "os"

func parseEnvs(cfg *Config) {
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerSettings.Address = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.ServiceSettings.BaseURL = envBaseURL
	}
}
