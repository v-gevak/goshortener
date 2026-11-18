package urlutil

import (
	"fmt"
	"net/url"
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
