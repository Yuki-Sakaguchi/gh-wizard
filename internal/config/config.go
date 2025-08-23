package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents application settings
type Config struct {
	DefaultPrivate   bool     `yaml:"default_private"`
	DefaultClone     bool     `yaml:"default_clone"`
	DefaultAddRemote bool     `yaml:"default_add_remote"`
	CacheTimeout     int      `yaml:"cache_timeout"`
	Theme            string   `yaml:"theme"`
	RecentTemplates  []string `yaml:"recent_templates"`
}

// GetConfigPath returns the configuration file path
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "gh-wizard")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.yaml"), nil

}

// Load reads the configuration file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return GetDefault(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration file: %w", err)
	}

	return &config, nil
}

// Save saves the configuration file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to convert configuration to YAML: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// Validate checks the validity of configuration values
func (c *Config) Validate() error {
	if c.CacheTimeout < 0 {
		return fmt.Errorf("cache timeout must be 0 or greater")
	}

	if c.Theme != "" && c.Theme != "default" && c.Theme != "dark" && c.Theme != "light" {
		return fmt.Errorf("theme must be one of: 'default', 'dark', 'light'")
	}

	return nil
}

// AddRecentTemplate adds a recently used template
func (c *Config) AddRecentTemplate(templateName string) {
	// Remove if already exists, to add to front
	for i, name := range c.RecentTemplates {
		if name == templateName {
			c.RecentTemplates = append(c.RecentTemplates[:i], c.RecentTemplates[i+1:]...)
			break
		}
	}

	// Add to front
	c.RecentTemplates = append([]string{templateName}, c.RecentTemplates...)

	// Limit to 10 items
	if len(c.RecentTemplates) > 10 {
		c.RecentTemplates = c.RecentTemplates[:10]
	}
}
