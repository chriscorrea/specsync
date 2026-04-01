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

// getConfigDir returns the config directory path
func getConfigDir() (string, error) {
	if configDir != "" {
		return configDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "specsync"), nil
}

// getTokenFromFile reads the token from the credentials file (if exists)
func getTokenFromFile() string {
	dir, err := getConfigDir()
	if err != nil {
		return ""
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

// SaveToken writes token to ~/.config/specsync/credentials.json
func SaveToken(token string) error {
	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	creds := struct {
		Token string `json:"token"`
	}{Token: token}

	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	credPath := filepath.Join(dir, "credentials.json")
	return os.WriteFile(credPath, data, 0600)
}

// DeleteToken removes the credentials file
func DeleteToken() error {
	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	credPath := filepath.Join(dir, "credentials.json")
	err = os.Remove(credPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// TokenSource indicates where auth is configured
type TokenSource int

const (
	TokenSourceNone TokenSource = iota
	TokenSourceEnv
	TokenSourceFile
)

// GetTokenSource returns where the token is configured (env, file, or none)
func GetTokenSource() TokenSource {
	if os.Getenv(envVar) != "" {
		return TokenSourceEnv
	}
	if getTokenFromFile() != "" {
		return TokenSourceFile
	}
	return TokenSourceNone
}
