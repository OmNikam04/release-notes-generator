package bugsby

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// Default token file location relative to home directory
	defaultTokenFileRel = ".local/state/artools_oauth2"
)

// TokenProvider handles authentication token retrieval for Bugsby API
type TokenProvider struct {
	tokenFile string
}

// NewTokenProvider creates a new token provider
func NewTokenProvider(tokenFile string) *TokenProvider {
	if tokenFile == "" {
		tokenFile = defaultTokenFileRel
	}
	return &TokenProvider{
		tokenFile: tokenFile,
	}
}

// GetToken retrieves the authentication token using the following priority:
// 1. BUGSBY_AUTH_TOKEN environment variable
// 2. Token file at ~/.local/state/artools_oauth2 (YAML format with "access_token" key)
func (tp *TokenProvider) GetToken() (string, error) {
	// Priority 1: Check environment variable
	if token := os.Getenv("BUGSBY_AUTH_TOKEN"); token != "" {
		return token, nil
	}

	// Priority 2: Read from token file
	return tp.readTokenFromFile()
}

// readTokenFromFile reads the OAuth2 token from the YAML file
func (tp *TokenProvider) readTokenFromFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	tokenPath := filepath.Join(homeDir, tp.tokenFile)
	
	file, err := os.Open(tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to open token file at %s: %w", tokenPath, err)
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return "", fmt.Errorf("failed to decode token YAML: %w", err)
	}

	accessToken, ok := data["access_token"].(string)
	if !ok || accessToken == "" {
		return "", errors.New("access_token not found in token file")
	}

	return accessToken, nil
}

// ValidateToken checks if a token is valid (non-empty)
func ValidateToken(token string) error {
	if token == "" {
		return errors.New("authentication token is empty")
	}
	return nil
}

