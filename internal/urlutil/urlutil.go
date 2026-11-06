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

func CompareURLs(a, b string) (bool, error) {
	normalizedA, err := normalizeHTTPURL(a)
	if err != nil {
		return false, err
	}

	normalizedB, err := normalizeHTTPURL(b)
	if err != nil {
		return false, err
	}

	return normalizedA == normalizedB, nil
}

func normalizeHTTPURL(rawURL string) (string, error) {
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

	if u.Path == "" {
		u.Path = "/"
	}

	return u.String(), nil
}
