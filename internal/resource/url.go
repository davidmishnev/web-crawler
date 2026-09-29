package resource

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeURL returns an absolute HTTP(S) URL suitable for queue deduplication.
func NormalizeURL(u url.URL) (url.URL, error) {
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.Opaque != "" || u.User != nil {
		return url.URL{}, fmt.Errorf("expected an absolute HTTP(S) URL without credentials")
	}
	u.Fragment, u.RawFragment = "", ""
	if u.Path == "" {
		u.Path = "/"
	}
	if (u.Scheme == "http" && u.Port() == "80") || (u.Scheme == "https" && u.Port() == "443") {
		u.Host = u.Hostname()
		if strings.Contains(u.Host, ":") {
			u.Host = "[" + u.Host + "]"
		}
	}
	return u, nil
}
