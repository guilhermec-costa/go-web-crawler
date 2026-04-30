package crawler

import (
	"fmt"
	"net/url"
)

func ValidateUrl(u *url.URL) error {
	switch {
	case u.Host == "":
		return fmt.Errorf("url host should not be empty")
	case u.Scheme != "http" && u.Scheme != "https":
		return fmt.Errorf("url should be http or https only")
	}

	return nil
}
