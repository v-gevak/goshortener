package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	oldArgs, oldFlags := os.Args, flag.CommandLine
	t.Cleanup(func() {
		os.Args, flag.CommandLine = oldArgs, oldFlags
	})

	tests := []struct {
		name               string
		osArgs             []string
		envVars            map[string]string
		expectedServerAddr string
		expectedBaseURL    string
	}{
		{
			name:               "default variables",
			osArgs:             []string{"shortener"},
			envVars:            map[string]string{},
			expectedServerAddr: "localhost:8080",
			expectedBaseURL:    "http://localhost:8080",
		},
		{
			name:               "set config from args",
			osArgs:             []string{"shortener", "-a", "localhost:8081", "-b", "http://localhost:8081"},
			envVars:            map[string]string{},
			expectedServerAddr: "localhost:8081",
			expectedBaseURL:    "http://localhost:8081",
		},
		{
			name:   "env variables override args",
			osArgs: []string{"shortener", "-a", "localhost:8081", "-b", "http://localhost:8081"},
			envVars: map[string]string{
				"SERVER_ADDRESS": "localhost:9090",
				"BASE_URL":       "https://short.example",
			},
			expectedServerAddr: "localhost:9090",
			expectedBaseURL:    "https://short.example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("shortener", flag.ContinueOnError)
			os.Args = tt.osArgs

			for key, envVar := range tt.envVars {
				t.Setenv(key, envVar)
			}

			cfg := New()

			assert.Equal(t, tt.expectedServerAddr, cfg.ServerSettings.Address)
			assert.Equal(t, tt.expectedBaseURL, cfg.ServiceSettings.BaseURL)
		})
	}

}
