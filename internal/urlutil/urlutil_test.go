package urlutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCompareURLs(t *testing.T) {
	tests := []struct {
		name      string
		a         string
		b         string
		wantEqual bool
		wantErr   bool
	}{
		{name: "identical URLs", a: "https://example.com/page?q=1#part", b: "https://example.com/page?q=1#part", wantEqual: true},
		{name: "scheme case", a: "HTTPS://example.com/page", b: "https://example.com/page", wantEqual: true},
		{name: "host case", a: "https://EXAMPLE.COM/page", b: "https://example.com/page", wantEqual: true},
		{name: "default HTTP port", a: "http://example.com:80/page", b: "http://example.com/page", wantEqual: true},
		{name: "default HTTPS port", a: "https://example.com:443/page", b: "https://example.com/page", wantEqual: true},
		{name: "empty root path", a: "https://example.com", b: "https://example.com/", wantEqual: true},
		{name: "empty fragment", a: "https://example.com#", b: "https://example.com", wantEqual: true},
		{name: "different schemes", a: "http://example.com", b: "https://example.com"},
		{name: "different hosts", a: "https://example.ru", b: "https://example.com"},
		{name: "www is a different host", a: "https://www.example.com", b: "https://example.com"},
		{name: "different ports", a: "https://example.com:8080/page", b: "https://example.com:9090/page"},
		{name: "nondefault port", a: "https://example.com:80/page", b: "https://example.com/page"},
		{name: "different paths", a: "https://example.com/a", b: "https://example.com/b"},
		{name: "query values", a: "https://example.com/?a=b", b: "https://example.com/?a=c"},
		{name: "different fragments", a: "https://example.com/#first", b: "https://example.com/#second"},
		{name: "fragment presence", a: "https://example.com/#part", b: "https://example.com/"},
		{name: "invalid first URL", a: "https://example.com/%zz", b: "https://example.com", wantErr: true},
		{name: "unsupported scheme", a: "ftp://example.com", b: "https://example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEqual, err := CompareURLs(tt.a, tt.b)
			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, isEqual)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantEqual, isEqual)
		})
	}
}
