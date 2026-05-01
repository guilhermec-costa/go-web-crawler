package crawler

import (
	"fmt"
	"net/url"
)

func ValidateURL(u *url.URL) error {
	switch {
	case u.Host == "":
		return fmt.Errorf("url host should not be empty")
	case u.Scheme != "http" && u.Scheme != "https":
		return fmt.Errorf("url should be http or https only")
	}

	return nil
}

func ParseAndValidateURL(rawUrl string) (*url.URL, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %s: %w", rawUrl, err)
	}

	if err := ValidateURL(parsedUrl); err != nil {
		return nil, fmt.Errorf("url %s is not valid : %w", rawUrl, err)
	}

	return parsedUrl, nil
}
