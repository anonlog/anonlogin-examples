package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const configFileName = "config.json"

// CLIConfig is the persisted CLI configuration.
type CLIConfig struct {
	IssuerURL string `json:"issuer_url"`
	ClientID  string `json:"client_id"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "anonlogin")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func loadConfig() (*CLIConfig, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(dir, configFileName))
	if os.IsNotExist(err) {
		return &CLIConfig{
			IssuerURL: "https://anonlog.in",
			ClientID:  "anonlogin-cli",
		}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg CLIConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveConfig(cfg *CLIConfig) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, configFileName), b, 0600)
}

func newConfigCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value (keys: issuer, client_id)",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			switch args[0] {
			case "issuer":
				cfg.IssuerURL = args[1]
			case "client_id":
				cfg.ClientID = args[1]
			default:
				return fmt.Errorf("unknown config key %q (known: issuer, client_id)", args[0])
			}
			if err := saveConfig(cfg); err != nil {
				return err
			}
			fmt.Printf("Set %s = %s\n", args[0], args[1])
			return nil
		},
	}

	c.AddCommand(setCmd)
	return c
}
