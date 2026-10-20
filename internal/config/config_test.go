package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		serverAddr  string
		baseURL     string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid config",
			serverAddr: "localhost:8080",
			baseURL:    "http://localhost:8080",
			wantErr:    false,
		},
		{
			name:       "server address without host",
			serverAddr: ":8080",
			baseURL:    "http://localhost:8080",
			wantErr:    false,
		},
		{
			name:        "server address without port",
			serverAddr:  "localhost",
			baseURL:     "http://localhost:8080",
			wantErr:     true,
			errContains: "server address",
		},
		{
			name:        "server port out of range",
			serverAddr:  "localhost:0",
			baseURL:     "http://localhost:8080",
			wantErr:     true,
			errContains: "invalid port",
		},
		{
			name:        "base url with unsupported scheme",
			serverAddr:  "localhost:8080",
			baseURL:     "ftp://localhost:8080",
			wantErr:     true,
			errContains: "scheme must be http or https",
		},
		{
			name:        "base url with empty host",
			serverAddr:  "localhost:8080",
			baseURL:     "http://",
			wantErr:     true,
			errContains: "host must not be empty",
		},
		{
			name:        "base url with query parameters",
			serverAddr:  "localhost:8080",
			baseURL:     "http://localhost:8080?x=1",
			wantErr:     true,
			errContains: "query parameters and fragments are not allowed",
		},
		{
			name:        "base url port out of range",
			serverAddr:  "localhost:8080",
			baseURL:     "http://localhost:65536",
			wantErr:     true,
			errContains: "invalid port",
		},
		{
			name:        "server port is not numeric",
			serverAddr:  "localhost:abc",
			baseURL:     "http://localhost:8080",
			wantErr:     true,
			errContains: "invalid port",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ServerAddress: tt.serverAddr,
				BaseURL:       tt.baseURL,
			}

			err := cfg.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
