package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Provider struct {
	Description string            `json:"description"`
	BaseURL     string            `json:"base_url"`
	APIKey      string            `json:"api_key"`
	Model       string            `json:"model"`
	Env         map[string]string `json:"env"`
}

type Config struct {
	Providers map[string]Provider `json:"providers"`
}

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
				BaseURL:     "https://open.bigmodel.cn/api/paas/v4",
				APIKey:      "",
				Model:       "glm-5",
				Env:         map[string]string{},
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

func Init() (string, bool, error) {
	path, err := ConfigPath()
	if err != nil {
		return "", false, err
	}

	if _, err := os.Stat(path); err == nil {
		return path, false, nil
	}

	cfg := defaultConfig()
	if err := Save(cfg); err != nil {
		return "", false, fmt.Errorf("creating config: %w", err)
	}

	return path, true, nil
}

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

func (c *Config) Get(name string) (Provider, bool) {
	p, ok := c.Providers[name]
	return p, ok
}

func (c *Config) Add(name string, p Provider) {
	c.Providers[name] = p
}

func (c *Config) Remove(name string) bool {
	if _, ok := c.Providers[name]; !ok {
		return false
	}
	delete(c.Providers, name)
	return true
}

func (c *Config) List() map[string]Provider {
	return c.Providers
}
