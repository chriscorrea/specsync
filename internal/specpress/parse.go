package specpress

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// ExtractProjectID parses a project ID from UUID string
// sec note: client-side parsing only, never fetch a URL
func ExtractProjectID(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("empty input")
	}

	if _, err := uuid.Parse(input); err == nil {
		return input, nil
	}

	// else, parse as URL
	u, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid input")
	}

	// strict host allowlist to allow project ID extraction
	// we can relax in the future if needed but want to avoid SSRF attempts by default
	if u.Host != "spec.press" && u.Host != "www.spec.press" {
		return "", fmt.Errorf("URL host is invalid")
	}

	// extract UUID from /project/{uuid} or /projects/{uuid}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 2 && (parts[0] == "project" || parts[0] == "projects" || parts[0] == "p") {
		if _, err := uuid.Parse(parts[1]); err == nil {
			return parts[1], nil
		}
	}

	return "", fmt.Errorf("could not extract project ID from URL")
}
