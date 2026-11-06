package config

type ServiceSettings struct {
	BaseURL string
}

type ServerSettings struct {
	Address string
}

type Config struct {
	ServiceSettings ServiceSettings
	ServerSettings  ServerSettings
}

func New() *Config {
	cfg := new(Config)
	ParseFlags(cfg)
	return cfg
}
