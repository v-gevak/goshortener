package config

import (
	"fmt"
	"os"
	"strconv"
)

func parseEnvs(cfg *Config) error {
	setEnvString("SERVER_ADDRESS", &cfg.ServerSettings.Address)
	setEnvString("BASE_URL", &cfg.ServiceSettings.BaseURL)
	setEnvString("LOG_LEVEL", &cfg.LogSettings.Level)
	setEnvString("FILE_LOGGING_PATH", &cfg.LogSettings.FilePath)

	if err := setEnvBool("ENABLE_FILE_LOGGING", &cfg.LogSettings.EnableFile); err != nil {
		return err
	}

	if err := setEnvBool("ENABLE_CONSOLE_LOGGING", &cfg.LogSettings.EnableConsole); err != nil {
		return err
	}

	return nil
}

func setEnvString(key string, target *string) {
	if value := os.Getenv(key); value != "" {
		*target = value
	}
}
func setEnvBool(key string, target *bool) error {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", key, value, err)
	}

	*target = parsed
	return nil
}
