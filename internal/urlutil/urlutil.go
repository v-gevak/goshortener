package urlutil

import (
	"fmt"
	"net/url"
	"strings"
)

func ValidateHTTPURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("scheme must be http or https")
	}

	if u.Hostname() == "" {
		return nil, fmt.Errorf("host must not be empty")
	}

	return u, nil
}

func NormalizeHTTPURL(rawURL string) (string, error) {
	u, err := ValidateHTTPURL(rawURL)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(u.Host, "[") {
		u.Host = strings.ToLower(u.Host)
	}

	port := u.Port()
	if (u.Scheme == "http" && port == "80") ||
		(u.Scheme == "https" && port == "443") {
		u.Host = strings.TrimSuffix(u.Host, ":"+port)
	}

	escapedPath := strings.TrimRight(u.EscapedPath(), "/")
	u.Path, err = url.PathUnescape(escapedPath)
	if err != nil {
		return "", err
	}
	u.RawPath = escapedPath

	return u.String(), nil
}
