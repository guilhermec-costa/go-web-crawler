package crawler

import (
	"fmt"
	"net/url"
)

func GetPage(pageLink string) {
	// resp, err := http.Get()
}

func ValidateUrl(u *url.URL) error {
	switch {
	case u.Scheme != "http" && u.Scheme != "https":
		return fmt.Errorf("url should be http or https only")
	case u.Host == "":
		return fmt.Errorf("url host should not be empty")
	}

	return nil
}
