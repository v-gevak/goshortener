package urlutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeHTTPURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{name: "no path", url: "https://example.com", want: "https://example.com"},
		{name: "root slash", url: "https://example.com/", want: "https://example.com"},
		{name: "path slash", url: "https://example.com/page/", want: "https://example.com/page"},
		{name: "multiple trailing slashes", url: "https://example.com/page///", want: "https://example.com/page"},
		{name: "inner slashes preserved", url: "https://example.com/a//b/", want: "https://example.com/a//b"},
		{name: "query and fragment preserved", url: "HTTPS://EXAMPLE.COM:443/page/?next=/x/#part/", want: "https://example.com/page?next=/x/#part/"},
		{name: "encoded trailing slash preserved", url: "https://example.com/page%2F", want: "https://example.com/page%2F"},
		{name: "encoded inner slash preserved", url: "https://example.com/a%2Fb/", want: "https://example.com/a%2Fb"},
		{name: "literal slash after encoded slash", url: "https://example.com/page%2F/", want: "https://example.com/page%2F"},
		{name: "unicode path", url: "https://example.com/я/", want: "https://example.com/%D1%8F"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeHTTPURL(tt.url)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			again, err := NormalizeHTTPURL(got)
			require.NoError(t, err)
			assert.Equal(t, got, again)
		})
	}
}

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
