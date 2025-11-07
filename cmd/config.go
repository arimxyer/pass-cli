package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"pass-cli/internal/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Pass-CLI configuration",
	Long: `Manage Pass-CLI configuration settings for terminal warnings and keyboard shortcuts.

Configuration file location:
  All platforms: ~/.pass-cli/config.yml`,
}

// configInitCmd creates a new config file with examples
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create configuration file with examples",
	Long: `Create a new configuration file at the default location with commented examples.

If a configuration file already exists, this command will fail. Use 'config reset' to overwrite.`,
	Run: runConfigInit,
}

// configEditCmd opens the config file in an editor
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration file in editor",
	Long: `Open the configuration file in your default editor.

The editor is determined by:
  1. EDITOR environment variable
  2. Platform-specific defaults (notepad on Windows, nano/vim/vi on Linux/macOS)

If no configuration file exists, it will be created with defaults.`,
	Run: runConfigEdit,
}

// configValidateCmd validates the config file
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long: `Check the configuration file for errors and display validation results.

Exit codes:
  0 - Configuration is valid
  1 - Configuration has errors
  2 - File system error (cannot read config file)

Validation checks:
  - Terminal size ranges (1-10000 width, 1-1000 height)
  - Keybinding conflicts (no duplicate key assignments)
  - Unknown actions (all keybindings must map to known actions)
  - Key format (valid key syntax)`,
	Run: runConfigValidate,
}

// configResetCmd resets config to defaults
var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long: `Reset the configuration file to default values.

A backup of the current configuration will be created at:
  <config-path>.backup

If a backup already exists, it will be overwritten.`,
	Run: runConfigReset,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configResetCmd)
}

// runConfigInit creates a new config file with examples
func runConfigInit(cmd *cobra.Command, args []string) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Cannot determine config path: %v\n", err)
		os.Exit(2)
	}

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: Config file already exists at %s\n", configPath)
		fmt.Fprintf(os.Stderr, "Use 'pass-cli config edit' to modify or 'pass-cli config reset' to overwrite\n")
		os.Exit(2)
	}

	// Write default config template
	template := config.GetDefaultConfigTemplate()
	if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create config file: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("Config file created at %s\n", configPath)
	fmt.Printf("Edit with: pass-cli config edit\n")
}

// runConfigEdit opens the config file in an editor
func runConfigEdit(cmd *cobra.Command, args []string) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Cannot determine config path: %v\n", err)
		os.Exit(2)
	}

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		template := config.GetDefaultConfigTemplate()
		if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create config file: %v\n", err)
			os.Exit(2)
		}
		fmt.Printf("Created new config file at %s\n", configPath)
	}

	// Open in editor
	if err := config.OpenEditor(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open editor: %v\n", err)
		os.Exit(2)
	}
}

// runConfigValidate validates the config file
func runConfigValidate(cmd *cobra.Command, args []string) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Cannot determine config path: %v\n", err)
		os.Exit(2)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("No config file found, using defaults")
		fmt.Printf("Run 'pass-cli config init' to create a config file at %s\n", configPath)
		os.Exit(0)
	}

	// Load and validate config
	cfg, result := config.LoadFromPath(configPath)

	// Display validation results
	if result.Valid {
		fmt.Println("✓ Config valid")

		// Display config summary
		fmt.Printf("\nTerminal: warning_enabled=%v, min_width=%d, min_height=%d\n",
			cfg.Terminal.WarningEnabled,
			cfg.Terminal.MinWidth,
			cfg.Terminal.MinHeight)

		fmt.Printf("Keybindings: %d custom bindings loaded\n", len(cfg.Keybindings))

		// Display warnings if any
		if len(result.Warnings) > 0 {
			fmt.Println("\nWarnings:")
			for _, warning := range result.Warnings {
				if warning.Field != "" {
					fmt.Printf("  - %s: %s\n", warning.Field, warning.Message)
				} else {
					fmt.Printf("  - %s\n", warning.Message)
				}
			}
		}

		os.Exit(0)
	} else {
		fmt.Println("✗ Config has errors:")
		for i, err := range result.Errors {
			if err.Line > 0 {
				fmt.Printf("  %d. %s (line %d): %s\n", i+1, err.Field, err.Line, err.Message)
			} else {
				fmt.Printf("  %d. %s: %s\n", i+1, err.Field, err.Message)
			}
		}
		fmt.Println("\nUsing default settings. Fix errors and run 'pass-cli config validate' again.")
		os.Exit(1)
	}
}

// runConfigReset resets config to defaults
func runConfigReset(cmd *cobra.Command, args []string) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Cannot determine config path: %v\n", err)
		os.Exit(2)
	}

	// Create backup if config exists
	if _, err := os.Stat(configPath); err == nil {
		backupPath := configPath + ".backup"

		// Read current config
		currentConfig, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to read current config: %v\n", err)
			os.Exit(2)
		}

		// Write backup
		if err := os.WriteFile(backupPath, currentConfig, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create backup: %v\n", err)
			os.Exit(2)
		}

		fmt.Printf("Config file backed up to %s\n", backupPath)
	}

	// Write default config template
	template := config.GetDefaultConfigTemplate()
	if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to write config file: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("Config file reset to defaults at %s\n", configPath)
}
