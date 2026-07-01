// Package config loads and, on first run, creates nox's global TOML config.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Each [providers.*] entry supports either:
//   api_key_env = "SOME_ENV_VAR"   (reads the key from that env var, recommended)
//   api_key     = "sk-..."         (stores the key directly in this file)
// If both are set, api_key wins.
const defaultTemplate = `[default]
provider = "openai"
model = "gpt-4o-mini"
temperature = 0.2
max_tokens = 400
format = false   # true = always format command output into readable columns (same as passing --format)

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
	// Format, when true, makes nox always format command output into
	// readable columns, as if --format were passed on every call.
	Format bool `toml:"format"`
}

type Provider struct {
	BaseURL string `toml:"base_url"`
	// APIKey is an optional direct key value, for convenience. Takes
	// precedence over APIKeyEnv when set.
	APIKey string `toml:"api_key"`
	// APIKeyEnv names an environment variable to read the key from.
	// Preferred over APIKey when the config file may be shared/synced.
	APIKeyEnv string `toml:"api_key_env"`
}

type Config struct {
	Default   Default             `toml:"default"`
	Providers map[string]Provider `toml:"providers"`
}

// Dir returns nox's data directory (~/.nox), honoring NOX_HOME for overrides.
func Dir() (string, error) {
	if d := os.Getenv("NOX_HOME"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".nox"), nil
}

// Path returns the global config file path, honoring NOX_CONFIG for overrides.
func Path() (string, error) {
	if p := os.Getenv("NOX_CONFIG"); p != "" {
		return p, nil
	}
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// legacyPath is the pre-~/.nox config location (~/.config/nox/config.toml).
func legacyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "nox", "config.toml"), nil
}

// migrateLegacy moves a config found at the old ~/.config/nox location to
// newPath, if present and newPath doesn't already exist.
func migrateLegacy(newPath string) error {
	oldPath, err := legacyPath()
	if err != nil || oldPath == newPath {
		return nil
	}
	if _, err := os.Stat(oldPath); err != nil {
		return nil
	}
	if _, err := os.Stat(newPath); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("could not migrate legacy config: %w", err)
	}
	fmt.Fprintf(os.Stderr, "nox: migrated config from %s to %s\n", oldPath, newPath)
	return nil
}

// Load reads the global config, creating it with defaults on first run.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	if err := migrateLegacy(path); err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := create(path); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "nox: created new config at %s\n", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config (%s): %w", path, err)
	}
	return &cfg, nil
}

func create(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	return os.WriteFile(path, []byte(defaultTemplate), 0o600)
}

// ActiveProvider resolves the provider config for cfg.Default.Provider.
func (c *Config) ActiveProvider() (Provider, error) {
	p, ok := c.Providers[c.Default.Provider]
	if !ok {
		return Provider{}, fmt.Errorf("provider %q is not defined in config.toml", c.Default.Provider)
	}
	return p, nil
}

// ResolveAPIKey returns the provider's API key: the direct APIKey value if
// set, otherwise the value of the APIKeyEnv environment variable.
func (p Provider) ResolveAPIKey() string {
	if p.APIKey != "" {
		return p.APIKey
	}
	if p.APIKeyEnv == "" {
		return ""
	}
	return os.Getenv(p.APIKeyEnv)
}
