// Copyright (c) 2025 Samedi Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

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

// getConfigValue retrieves a nested config value by dot-notation key
func getConfigValue(cfg *config.Config, key string) interface{} {
	// Map of config keys to values
	configMap := map[string]interface{}{
		"user.email":                     cfg.User.Email,
		"user.username":                  cfg.User.Username,
		"user.timezone":                  cfg.User.Timezone,
		"llm.provider":                   cfg.LLM.Provider,
		"llm.cli_command":                cfg.LLM.CLICommand,
		"llm.default_model":              cfg.LLM.DefaultModel,
		"llm.timeout_seconds":            cfg.LLM.TimeoutSeconds,
		"storage.data_dir":               cfg.Storage.DataDir,
		"storage.backup_enabled":         cfg.Storage.BackupEnabled,
		"storage.backup_dir":             cfg.Storage.BackupDir,
		"storage.auto_backup_days":       cfg.Storage.AutoBackupDays,
		"tui.theme":                      cfg.TUI.Theme,
		"tui.date_format":                cfg.TUI.DateFormat,
		"tui.time_format":                cfg.TUI.TimeFormat,
		"learning.default_chunk_minutes": cfg.Learning.DefaultChunkMinutes,
		"learning.reminder_enabled":      cfg.Learning.ReminderEnabled,
		"learning.streak_tracking":       cfg.Learning.StreakTracking,
	}

	value, ok := configMap[key]
	if !ok {
		return nil
	}
	return value
}

// setConfigValue sets a nested config value by dot-notation key
func setConfigValue(cfg *config.Config, key, value string) error {
	switch key {
	case "user.email":
		cfg.User.Email = value
	case "user.username":
		cfg.User.Username = value
	case "user.timezone":
		cfg.User.Timezone = value
	case "llm.provider":
		cfg.LLM.Provider = value
	case "llm.cli_command":
		cfg.LLM.CLICommand = value
	case "llm.default_model":
		cfg.LLM.DefaultModel = value
	case "storage.data_dir":
		cfg.Storage.DataDir = value
	case "storage.backup_dir":
		cfg.Storage.BackupDir = value
	case "tui.theme":
		cfg.TUI.Theme = value
	case "tui.date_format":
		cfg.TUI.DateFormat = value
	case "tui.time_format":
		cfg.TUI.TimeFormat = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}
