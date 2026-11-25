package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

type ServiceSettings struct {
	BaseURL string
}

type ServerSettings struct {
	Address string
}

type LogSettings struct {
	Level         string
	EnableFile    bool
	EnableConsole bool
	FilePath      string
}

type StoreSettings struct {
	FilePath string
	DBDSN    string
}

type AuthSettings struct {
	SigningKey string
}

type Config struct {
	ServiceSettings ServiceSettings
	ServerSettings  ServerSettings
	LogSettings     LogSettings
	StoreSettings   StoreSettings
	AuthSettings    AuthSettings
}

func New() (*Config, error) {
	cfg := new(Config)
	setDefaults(cfg)
	parseFlags(cfg)
	if err := parseEnvs(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getRandomSigningKey() string {
	var signingKey [sha256.Size]byte
	rand.Read(signingKey[:])
	return base64.StdEncoding.EncodeToString(signingKey[:])
}

func setDefaults(cfg *Config) {
	cfg.AuthSettings.SigningKey = getRandomSigningKey()

	cfg.LogSettings.Level = "info"
	cfg.LogSettings.EnableConsole = true
	cfg.LogSettings.EnableFile = false
	cfg.LogSettings.FilePath = "goshortener.log"

	cfg.ServiceSettings.BaseURL = "http://localhost:8080"
	cfg.ServerSettings.Address = "localhost:8080"

	cfg.StoreSettings.FilePath = ""
	cfg.StoreSettings.DBDSN = ""
}
