// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Load reads configuration from file and environment variables.
// It returns the default config if no config file exists.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Set up viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")

	// Add config search paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".samedi")
	v.AddConfigPath(configDir)

	// Allow environment variable overrides
	v.SetEnvPrefix("SAMEDI")
	v.AutomaticEnv()

	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		// If config file doesn't exist, use defaults
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into config struct
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Save writes the configuration to disk.
func Save(cfg *Config) error {
	// Validate before saving
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure config directory exists
	configPath := Path()
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set up viper for writing
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("toml")

	// Set values in viper
	v.Set("user", cfg.User)
	v.Set("llm", cfg.LLM)
	v.Set("storage", cfg.Storage)
	v.Set("sync", cfg.Sync)
	v.Set("tui", cfg.TUI)
	v.Set("learning", cfg.Learning)

	// Write config file
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Set secure permissions (owner read/write only)
	if err := os.Chmod(configPath, 0o600); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}

	return nil
}

// InitConfig creates the default config file if it doesn't exist.
func InitConfig() error {
	configPath := Path()

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// Create default config
	cfg := DefaultConfig()

	// Save to disk
	if err := Save(cfg); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	return nil
}
