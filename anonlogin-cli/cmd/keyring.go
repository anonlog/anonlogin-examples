package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const keyringSvc = "anonlogin-cli"
const keyringKey = "tokens"

// TokenStore holds the current access and refresh tokens.
type TokenStore struct {
	IssuerURL    string `json:"issuer_url"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// saveTokens persists tokens to the OS keychain, with a file fallback.
func saveTokens(ts *TokenStore) error {
	b, err := json.Marshal(ts)
	if err != nil {
		return err
	}
	if err := keyring.Set(keyringSvc, keyringKey, string(b)); err == nil {
		return nil
	}
	// Fallback: encrypted file in ~/.config/anonlogin/
	return saveTokenFile(b)
}

// loadTokens reads tokens from the OS keychain, falling back to file.
func loadTokens() (*TokenStore, error) {
	raw, err := keyring.Get(keyringSvc, keyringKey)
	if err != nil {
		raw2, err2 := loadTokenFile()
		if err2 != nil {
			return nil, fmt.Errorf("no stored credentials (keychain: %v, file: %v)", err, err2)
		}
		raw = string(raw2)
	}
	var ts TokenStore
	if err := json.Unmarshal([]byte(raw), &ts); err != nil {
		return nil, fmt.Errorf("invalid stored token data: %w", err)
	}
	return &ts, nil
}

// deleteTokens removes stored tokens from the keychain and file fallback.
func deleteTokens() {
	_ = keyring.Delete(keyringSvc, keyringKey)
	dir, err := configDir()
	if err == nil {
		_ = os.Remove(filepath.Join(dir, "tokens.json"))
	}
}

func tokenFilePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tokens.json"), nil
}

func saveTokenFile(b []byte) error {
	path, err := tokenFilePath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func loadTokenFile() ([]byte, error) {
	path, err := tokenFilePath()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}
