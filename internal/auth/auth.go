package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// configDir can be overridden for testing
var configDir = ""

const envVar = "SPECSYNC_TOKEN"

// GetToken returns auth token from env or credentials file
// returns empty string if no auth configured (auth is optional)
func GetToken() string {
	// env var takes precedence
	if token := os.Getenv(envVar); token != "" {
		return token
	}

	// else, try credentials file
	return getTokenFromFile()
}

// getTokenFromFile reads the token from the credentials file (if exists)
func getTokenFromFile() string {
	dir := configDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".config", "specsync")
	}

	credPath := filepath.Join(dir, "credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		return ""
	}

	var creds struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return ""
	}

	return creds.Token
}
