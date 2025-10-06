// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/pezware/samedi.dev/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage samedi configuration",
		Long: `View and modify samedi configuration.

Configuration is stored in ~/.samedi/config.toml

Examples:
  samedi config list                    # Show all settings
  samedi config get llm.provider        # Get specific setting
  samedi config set llm.provider claude # Set specific setting
  samedi config edit                    # Edit in $EDITOR`,
	}

	cmd.AddCommand(configListCmd())
	cmd.AddCommand(configGetCmd())
	cmd.AddCommand(configSetCmd())
	cmd.AddCommand(configEditCmd())
	cmd.AddCommand(configInitCmd())

	return cmd
}

func configListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration settings",
		Run: func(cmd *cobra.Command, _ []string) {
			cfg, err := getConfig(cmd)
			if err != nil {
				exitWithError("Failed to load config: %v", err)
			}

			jsonOutput, err := cmd.Flags().GetBool("json")
			if err != nil {
				exitWithError("Failed to get json flag: %v", err)
			}
			if jsonOutput {
				data, err := json.MarshalIndent(cfg, "", "  ")
				if err != nil {
					exitWithError("Failed to marshal config: %v", err)
				}
				fmt.Println(string(data))
			} else {
				// Pretty print as YAML
				data, err := yaml.Marshal(cfg)
				if err != nil {
					exitWithError("Failed to marshal config: %v", err)
				}
				fmt.Println(string(data))
			}
		},
	}
}

func configGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := getConfig(cmd)
			if err != nil {
				exitWithError("Failed to load config: %v", err)
			}

			key := args[0]
			value := getConfigValue(cfg, key)
			if value == nil {
				exitWithError("Unknown config key: %s", key)
			}

			fmt.Println(value)
		},
	}
}

func configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := getConfig(cmd)
			if err != nil {
				// If config doesn't exist, create default
				cfg = config.DefaultConfig()
			}

			key := args[0]
			value := args[1]

			if err := setConfigValue(cfg, key, value); err != nil {
				exitWithError("Failed to set config: %v", err)
			}

			if err := config.Save(cfg); err != nil {
				exitWithError("Failed to save config: %v", err)
			}

			fmt.Printf("✓ Set %s = %s\n", key, value)
		},
	}
}

func configEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit configuration in $EDITOR",
		Run: func(_ *cobra.Command, _ []string) {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}

			configPath := config.Path()

			// Ensure config exists
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				if err := config.InitConfig(); err != nil {
					exitWithError("Failed to create config: %v", err)
				}
			}

			// Open in editor
			editorCmd := exec.Command(editor, configPath)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr

			if err := editorCmd.Run(); err != nil {
				exitWithError("Failed to edit config: %v", err)
			}

			// Validate after editing
			if _, err := config.Load(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Config validation failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "Please fix the config file and try again.\n")
			}
		},
	}
}

func configInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize default configuration file",
		Run: func(_ *cobra.Command, _ []string) {
			if err := config.InitConfig(); err != nil {
				exitWithError("Failed to initialize config: %v", err)
			}

			fmt.Printf("✓ Configuration created at %s\n", config.Path())
		},
	}
}

var configValueResolvers = map[string]func(*config.Config) interface{}{
	"user.email":                     func(cfg *config.Config) interface{} { return cfg.User.Email },
	"user.username":                  func(cfg *config.Config) interface{} { return cfg.User.Username },
	"user.timezone":                  func(cfg *config.Config) interface{} { return cfg.User.Timezone },
	"llm.provider":                   func(cfg *config.Config) interface{} { return cfg.LLM.Provider },
	"llm.cli_command":                func(cfg *config.Config) interface{} { return cfg.LLM.CLICommand },
	"llm.default_model":              func(cfg *config.Config) interface{} { return cfg.LLM.DefaultModel },
	"llm.timeout_seconds":            func(cfg *config.Config) interface{} { return cfg.LLM.TimeoutSeconds },
	"storage.data_dir":               func(cfg *config.Config) interface{} { return cfg.Storage.DataDir },
	"storage.backup_enabled":         func(cfg *config.Config) interface{} { return cfg.Storage.BackupEnabled },
	"storage.backup_dir":             func(cfg *config.Config) interface{} { return cfg.Storage.BackupDir },
	"storage.auto_backup_days":       func(cfg *config.Config) interface{} { return cfg.Storage.AutoBackupDays },
	"sync.enabled":                   func(cfg *config.Config) interface{} { return cfg.Sync.Enabled },
	"sync.cloudflare_endpoint":       func(cfg *config.Config) interface{} { return cfg.Sync.CloudflareEndpoint },
	"sync.sync_interval_minutes":     func(cfg *config.Config) interface{} { return cfg.Sync.SyncIntervalMinutes },
	"tui.theme":                      func(cfg *config.Config) interface{} { return cfg.TUI.Theme },
	"tui.date_format":                func(cfg *config.Config) interface{} { return cfg.TUI.DateFormat },
	"tui.time_format":                func(cfg *config.Config) interface{} { return cfg.TUI.TimeFormat },
	"tui.first_day_of_week":          func(cfg *config.Config) interface{} { return cfg.TUI.FirstDayOfWeek },
	"learning.default_chunk_minutes": func(cfg *config.Config) interface{} { return cfg.Learning.DefaultChunkMinutes },
	"learning.reminder_enabled":      func(cfg *config.Config) interface{} { return cfg.Learning.ReminderEnabled },
	"learning.reminder_message":      func(cfg *config.Config) interface{} { return cfg.Learning.ReminderMessage },
	"learning.streak_tracking":       func(cfg *config.Config) interface{} { return cfg.Learning.StreakTracking },
}

// getConfigValue retrieves a nested config value by dot-notation key.
func getConfigValue(cfg *config.Config, key string) interface{} {
	if resolver, ok := configValueResolvers[key]; ok {
		return resolver(cfg)
	}
	return nil
}

var stringConfigSetters = map[string]func(*config.Config, string){
	"user.email":                func(cfg *config.Config, value string) { cfg.User.Email = value },
	"user.username":             func(cfg *config.Config, value string) { cfg.User.Username = value },
	"user.timezone":             func(cfg *config.Config, value string) { cfg.User.Timezone = value },
	"llm.provider":              func(cfg *config.Config, value string) { cfg.LLM.Provider = value },
	"llm.cli_command":           func(cfg *config.Config, value string) { cfg.LLM.CLICommand = value },
	"llm.default_model":         func(cfg *config.Config, value string) { cfg.LLM.DefaultModel = value },
	"storage.data_dir":          func(cfg *config.Config, value string) { cfg.Storage.DataDir = value },
	"storage.backup_dir":        func(cfg *config.Config, value string) { cfg.Storage.BackupDir = value },
	"sync.cloudflare_endpoint":  func(cfg *config.Config, value string) { cfg.Sync.CloudflareEndpoint = value },
	"tui.theme":                 func(cfg *config.Config, value string) { cfg.TUI.Theme = value },
	"tui.date_format":           func(cfg *config.Config, value string) { cfg.TUI.DateFormat = value },
	"tui.time_format":           func(cfg *config.Config, value string) { cfg.TUI.TimeFormat = value },
	"tui.first_day_of_week":     func(cfg *config.Config, value string) { cfg.TUI.FirstDayOfWeek = value },
	"learning.reminder_message": func(cfg *config.Config, value string) { cfg.Learning.ReminderMessage = value },
}

var intConfigSetters = map[string]func(*config.Config, int){
	"llm.timeout_seconds":            func(cfg *config.Config, value int) { cfg.LLM.TimeoutSeconds = value },
	"storage.auto_backup_days":       func(cfg *config.Config, value int) { cfg.Storage.AutoBackupDays = value },
	"sync.sync_interval_minutes":     func(cfg *config.Config, value int) { cfg.Sync.SyncIntervalMinutes = value },
	"learning.default_chunk_minutes": func(cfg *config.Config, value int) { cfg.Learning.DefaultChunkMinutes = value },
}

var boolConfigSetters = map[string]func(*config.Config, bool){
	"storage.backup_enabled":    func(cfg *config.Config, value bool) { cfg.Storage.BackupEnabled = value },
	"sync.enabled":              func(cfg *config.Config, value bool) { cfg.Sync.Enabled = value },
	"learning.reminder_enabled": func(cfg *config.Config, value bool) { cfg.Learning.ReminderEnabled = value },
	"learning.streak_tracking":  func(cfg *config.Config, value bool) { cfg.Learning.StreakTracking = value },
}

// setConfigValue sets a nested config value by dot-notation key.
func setConfigValue(cfg *config.Config, key, value string) error {
	if setter, ok := stringConfigSetters[key]; ok {
		setter(cfg, value)
		return nil
	}

	if setter, ok := intConfigSetters[key]; ok {
		parsed, err := parseInt(value, key)
		if err != nil {
			return err
		}
		setter(cfg, parsed)
		return nil
	}

	if setter, ok := boolConfigSetters[key]; ok {
		parsed, err := parseBool(value, key)
		if err != nil {
			return err
		}
		setter(cfg, parsed)
		return nil
	}

	return fmt.Errorf("unknown config key: %s", key)
}

func parseBool(value, key string) (bool, error) {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean for %s: %w", key, err)
	}
	return parsed, nil
}

func parseInt(value, key string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s: %w", key, err)
	}
	return parsed, nil
}
