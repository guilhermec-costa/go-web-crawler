package crawler

import (
	"fmt"
	"net/http"
	"net/url"
)

var statusHandlers = map[int]func(u *url.URL) error{
	http.StatusForbidden: func(u *url.URL) error {
		return fmt.Errorf("[ERROR] not allowed to crawl url %s", u)
	},

	http.StatusNotFound: func(u *url.URL) error {
		return fmt.Errorf("[ERROR] url \"%s\" page does not exist", u)
	},

	http.StatusInternalServerError: func(u *url.URL) error {
		return fmt.Errorf("[ERROR] url \"%s\" internal server error", u)
	},
}
