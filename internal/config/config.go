// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"path/filepath"
	"time"
)

// Config holds all user configuration for samedi.
type Config struct {
	User     UserConfig     `mapstructure:"user"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Sync     SyncConfig     `mapstructure:"sync"`
	TUI      TUIConfig      `mapstructure:"tui"`
	Learning LearningConfig `mapstructure:"learning"`
}

// UserConfig holds user identity and preferences.
type UserConfig struct {
	Email    string `mapstructure:"email"`
	Username string `mapstructure:"username"`
	Timezone string `mapstructure:"timezone"`
}

// LLMConfig holds LLM provider configuration.
type LLMConfig struct {
	Provider       string `mapstructure:"provider"`
	CLICommand     string `mapstructure:"cli_command"`
	DefaultModel   string `mapstructure:"default_model"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
}

// StorageConfig holds storage paths and backup settings.
type StorageConfig struct {
	DataDir        string `mapstructure:"data_dir"`
	BackupEnabled  bool   `mapstructure:"backup_enabled"`
	BackupDir      string `mapstructure:"backup_dir"`
	AutoBackupDays int    `mapstructure:"auto_backup_days"`
}

// SyncConfig holds cloud sync settings (Phase 2).
type SyncConfig struct {
	Enabled             bool   `mapstructure:"enabled"`
	CloudflareEndpoint  string `mapstructure:"cloudflare_endpoint"`
	SyncIntervalMinutes int    `mapstructure:"sync_interval_minutes"`
}

// TUIConfig holds TUI theme and display preferences.
type TUIConfig struct {
	Theme          string `mapstructure:"theme"`
	DateFormat     string `mapstructure:"date_format"`
	TimeFormat     string `mapstructure:"time_format"`
	FirstDayOfWeek string `mapstructure:"first_day_of_week"`
}

// LearningConfig holds learning session preferences.
type LearningConfig struct {
	DefaultChunkMinutes int    `mapstructure:"default_chunk_minutes"`
	ReminderEnabled     bool   `mapstructure:"reminder_enabled"`
	ReminderMessage     string `mapstructure:"reminder_message"`
	StreakTracking      bool   `mapstructure:"streak_tracking"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "." // Fallback to current directory
	}

	return &Config{
		User: UserConfig{
			Email:    "",
			Username: os.Getenv("USER"),
			Timezone: time.Local.String(),
		},
		LLM: LLMConfig{
			Provider:       "auto",
			CLICommand:     "",
			DefaultModel:   "", // Empty allows each provider to use its own default
			TimeoutSeconds: 300,
		},
		Storage: StorageConfig{
			DataDir:        filepath.Join(homeDir, ".samedi"),
			BackupEnabled:  true,
			BackupDir:      filepath.Join(homeDir, "samedi-backups"),
			AutoBackupDays: 7,
		},
		Sync: SyncConfig{
			Enabled:             false,
			CloudflareEndpoint:  "",
			SyncIntervalMinutes: 30,
		},
		TUI: TUIConfig{
			Theme:          "dracula",
			DateFormat:     "2006-01-02",
			TimeFormat:     "15:04",
			FirstDayOfWeek: "monday",
		},
		Learning: LearningConfig{
			DefaultChunkMinutes: 60,
			ReminderEnabled:     true,
			ReminderMessage:     "What did you learn today?",
			StreakTracking:      true,
		},
	}
}

// Path returns the default config file path.
func Path() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir is unavailable
		return filepath.Join(".samedi", "config.toml")
	}
	return filepath.Join(homeDir, ".samedi", "config.toml")
}
