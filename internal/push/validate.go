package push

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// cloud metadata hostnames to block; IP-based blocking handled by safeurl)
var blockedMetadataHosts = []string{
	"metadata.google.internal",
	"metadata.amazonaws.com",
}

// ValidateAPIURL checks URL structure, blocks metadata hostnames, requires HTTPS.
// IP-based SSRF protection is handled by safeurl at connection time.
func ValidateAPIURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("api_url is empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid api_url: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return errors.New("invalid api_url: missing scheme or host")
	}

	host := u.Hostname()

	// block known/common cloud metadata hostnames
	for _, blocked := range blockedMetadataHosts {
		if strings.EqualFold(host, blocked) {
			return fmt.Errorf("api_url cannot target metadata endpoints")
		}
	}

	// dev exception; localhost can use http
	isLocalhostDev := host == "localhost" || host == "127.0.0.1"
	if u.Scheme == "http" && !isLocalhostDev {
		return fmt.Errorf("HTTPS required for non-localhost hosts")
	}

	return nil
}
