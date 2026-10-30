package config

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

type Config struct {
	ServiceSettings ServiceSettings
	ServerSettings  ServerSettings
	LogSettings     LogSettings
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

func setDefaults(cfg *Config) {
	cfg.LogSettings.Level = "info"
	cfg.LogSettings.EnableConsole = true
	cfg.LogSettings.EnableFile = false
	cfg.LogSettings.FilePath = "goshortener.log"
	cfg.ServiceSettings.BaseURL = "http://localhost:8080"
	cfg.ServerSettings.Address = "localhost:8080"
}
