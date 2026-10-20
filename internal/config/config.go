package config

import (
	"fmt"
	"net"
	"strconv"

	"github.com/v-gevak/goshortener/internal/urlutil"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func New() (*Config, error) {
	cfg := new(Config)
	ParseFlags(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	if err := validateServerAddress(cfg.ServerAddress); err != nil {
		return fmt.Errorf("server address: %w", err)
	}

	if err := validateURL(cfg.BaseURL); err != nil {
		return fmt.Errorf("base URL: %w", err)
	}

	return nil
}

func validateServerAddress(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("expected host:port: %w", err)
	}

	if err := validatePort(port); err != nil {
		return err
	}

	return nil
}

func validateURL(rawURL string) error {
	u, err := urlutil.ValidateHTTPURL(rawURL)
	if err != nil {
		return err
	}

	if port := u.Port(); port != "" {
		if err := validatePort(port); err != nil {
			return err
		}
	}

	if u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("query parameters and fragments are not allowed")
	}

	return nil
}

func validatePort(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("invalid port %q: must be between 1 and 65535", port)
	}

	return nil
}
