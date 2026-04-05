package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	BaseURL   string `toml:"base_url"`
	ClientID  string `toml:"client_id"`
	SecretKey string `toml:"secret_key"`
}

type TokenCache struct {
	Token     string    `toml:"token"`
	ExpiresAt time.Time `toml:"expires_at"`
}

func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "synthwaves")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "synthwaves")
}

func ConfigPath() string {
	return filepath.Join(Dir(), "config.toml")
}

func TokenPath() string {
	return filepath.Join(Dir(), "token.toml")
}

func Load() (*Config, error) {
	return LoadFrom(ConfigPath())
}

func LoadFrom(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("loading config from %s: %w", path, err)
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if c.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("secret_key is required")
	}
	return nil
}

func (c *Config) Save() error {
	return c.SaveTo(ConfigPath())
}

func (c *Config) SaveTo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}

func LoadToken() (*TokenCache, error) {
	var tc TokenCache
	if _, err := toml.DecodeFile(TokenPath(), &tc); err != nil {
		return nil, err
	}
	return &tc, nil
}

func SaveToken(tc *TokenCache) error {
	path := TokenPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(tc)
}

func (tc *TokenCache) Valid() bool {
	return tc.Token != "" && time.Until(tc.ExpiresAt) > 60*time.Second
}
