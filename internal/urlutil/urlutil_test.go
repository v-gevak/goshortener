package urlutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateHTTPURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{name: "valid http", rawURL: "http://example.com", wantErr: false},
		{name: "valid https with path and query", rawURL: "https://example.com/page?q=1#x", wantErr: false},
		{name: "valid with port", rawURL: "https://example.com:8080/path", wantErr: false},
		{name: "valid IPv6 with port", rawURL: "http://[::1]:8080/path", wantErr: false},
		{name: "unsupported scheme", rawURL: "ftp://example.com", wantErr: true},
		{name: "missing scheme", rawURL: "example.com", wantErr: true},
		{name: "empty host", rawURL: "http://", wantErr: true},
		{name: "port without hostname", rawURL: "http://:8080/path", wantErr: true},
		{name: "empty bracketed hostname", rawURL: "http://[]:8080/path", wantErr: true},
		{name: "not a url", rawURL: "not a url", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateHTTPURL(tt.rawURL)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
