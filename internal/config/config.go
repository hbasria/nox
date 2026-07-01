// Package config loads and, on first run, creates nox's global TOML config.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const defaultTemplate = `[default]
provider = "openai"
model = "gpt-4o-mini"
temperature = 0.2
max_tokens = 400

[providers.openai]
base_url = "https://api.openai.com/v1"
api_key_env = "OPENAI_API_KEY"

[providers.groq]
base_url = "https://api.groq.com/openai/v1"
api_key_env = "GROQ_API_KEY"

[providers.ollama]
base_url = "http://localhost:11434/v1"
api_key_env = ""
`

type Default struct {
	Provider    string  `toml:"provider"`
	Model       string  `toml:"model"`
	Temperature float64 `toml:"temperature"`
	MaxTokens   int     `toml:"max_tokens"`
}

type Provider struct {
	BaseURL   string `toml:"base_url"`
	APIKeyEnv string `toml:"api_key_env"`
}

type Config struct {
	Default   Default             `toml:"default"`
	Providers map[string]Provider `toml:"providers"`
}

// Path returns the global config file path, honoring NOX_CONFIG for overrides.
func Path() (string, error) {
	if p := os.Getenv("NOX_CONFIG"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dizini bulunamadı: %w", err)
	}
	return filepath.Join(home, ".config", "nox", "config.toml"), nil
}

// Load reads the global config, creating it with defaults on first run.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := create(path); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "nox: yeni config oluşturuldu: %s\n", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config okunamadı: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config parse edilemedi (%s): %w", path, err)
	}
	return &cfg, nil
}

func create(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("config dizini oluşturulamadı: %w", err)
	}
	return os.WriteFile(path, []byte(defaultTemplate), 0o600)
}

// ActiveProvider resolves the provider config for cfg.Default.Provider.
func (c *Config) ActiveProvider() (Provider, error) {
	p, ok := c.Providers[c.Default.Provider]
	if !ok {
		return Provider{}, fmt.Errorf("provider %q config.toml içinde tanımlı değil", c.Default.Provider)
	}
	return p, nil
}

// APIKey reads the actual key from the environment variable named by APIKeyEnv.
func (p Provider) APIKey() string {
	if p.APIKeyEnv == "" {
		return ""
	}
	return os.Getenv(p.APIKeyEnv)
}
