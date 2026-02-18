// Package config provides configuration management for switcher.
// It handles reading/writing the ~/.switcher.json config file
// and managing AI provider settings.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Provider represents an AI provider configuration.
// It stores the base URL, API key, default model, and any additional environment variables.
type Provider struct {
	Description string            `json:"description"`
	BaseURL     string            `json:"base_url"`
	APIKey      string            `json:"api_key"`
	Model       string            `json:"model"`
	Env         map[string]string `json:"env"`
}

// Config represents the complete switcher configuration.
// It maps provider names to their Provider configurations.
type Config struct {
	Providers map[string]Provider `json:"providers"`
}

// ConfigPath returns the path to the switcher configuration file (~/.switcher.json).
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".switcher.json"), nil
}

func defaultConfig() Config {
	return Config{
		Providers: map[string]Provider{
			"moonshot": {
				Description: "Moonshot AI",
				BaseURL:     "https://api.moonshot.cn/v1",
				APIKey:      "",
				Model:       "kimi-k2.5",
				Env:         map[string]string{},
			},
			"zai": {
				Description: "ZhipuAI",
				BaseURL:     "https://api.z.ai/api/anthropic",
				APIKey:      "your_zai_api_key",
				Model:       "glm-5",
				Env: map[string]string{
					"API_TIMEOUT_MS": "3000000",
				},
			},
			"openrouter": {
				Description: "OpenRouter",
				BaseURL:     "https://openrouter.ai/api/v1",
				APIKey:      "",
				Model:       "",
				Env:         map[string]string{},
			},
			"deepseek": {
				Description: "DeepSeek",
				BaseURL:     "https://api.deepseek.com/v1",
				APIKey:      "",
				Model:       "deepseek-chat",
				Env:         map[string]string{},
			},
			"minimax": {
				Description: "MiniMax",
				BaseURL:     "https://api.minimax.io/anthropic",
				APIKey:      "",
				Model:       "MiniMax-M2.5",
				Env:         map[string]string{},
			},
		},
	}
}

// Init creates a config file with default providers if it doesn't already exist.
// Returns the config file path, a bool indicating if it was created, and any error.
func Init() (string, bool, error) {
	path, err := ConfigPath()
	if err != nil {
		return "", false, err
	}

	if _, err := os.Stat(path); err == nil {
		return path, false, nil
	} else if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("checking config file: %w", err)
	}

	cfg := defaultConfig()
	if err := Save(cfg); err != nil {
		return "", false, fmt.Errorf("creating config: %w", err)
	}

	return path, true, nil
}

// Load reads and parses the config file from ~/.switcher.json.
// If the file doesn't exist, it creates one with default providers.
func Load() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := defaultConfig()
			if err := Save(cfg); err != nil {
				return Config{}, fmt.Errorf("creating default config: %w", err)
			}
			return cfg, nil
		}
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Providers == nil {
		cfg.Providers = make(map[string]Provider)
	}

	return cfg, nil
}

// Save writes the config to ~/.switcher.json as formatted JSON.
func Save(cfg Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// Get retrieves a provider by name.
// Returns the provider and a bool indicating if it exists.
func (c *Config) Get(name string) (Provider, bool) {
	p, ok := c.Providers[name]
	return p, ok
}

// Add inserts or updates a provider in the config.
func (c *Config) Add(name string, p Provider) {
	c.Providers[name] = p
}

// Remove deletes a provider from the config.
// Returns true if the provider was found and deleted, false otherwise.
func (c *Config) Remove(name string) bool {
	if _, ok := c.Providers[name]; !ok {
		return false
	}
	delete(c.Providers, name)
	return true
}

// List returns all configured providers.
func (c *Config) List() map[string]Provider {
	return c.Providers
}

// DefaultProvider returns a built-in default provider by name.
// Returns the provider and a bool indicating if it exists.
func DefaultProvider(name string) (Provider, bool) {
	defaults := defaultConfig()
	p, ok := defaults.Providers[name]
	return p, ok
}
