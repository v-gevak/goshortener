package config

import (
	"encoding/base64"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Setenv("AUTH_SIGNING_KEY", "")
	envSigningKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
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
		expectedSigningKey string
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
				"SERVER_ADDRESS":   "localhost:9090",
				"BASE_URL":         "https://short.example",
				"AUTH_SIGNING_KEY": envSigningKey,
			},
			expectedServerAddr: "localhost:9090",
			expectedBaseURL:    "https://short.example",
			expectedSigningKey: envSigningKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("shortener", flag.ContinueOnError)
			os.Args = tt.osArgs

			for key, envVar := range tt.envVars {
				t.Setenv(key, envVar)
			}

			cfg, err := New()
			require.NoError(t, err)

			assert.Equal(t, tt.expectedServerAddr, cfg.ServerSettings.Address)
			assert.Equal(t, tt.expectedBaseURL, cfg.ServiceSettings.BaseURL)
			if tt.expectedSigningKey != "" {
				assert.Equal(t, tt.expectedSigningKey, cfg.AuthSettings.SigningKey)
			}
			key, err := base64.StdEncoding.DecodeString(cfg.AuthSettings.SigningKey)
			require.NoError(t, err)
			assert.Len(t, key, 32)
		})
	}

}
