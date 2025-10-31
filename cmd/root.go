package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"pass-cli/internal/config"
	"pass-cli/internal/vault"
)

var (
	cfgFile string
	verbose bool

	// Version information (set via ldflags during build)
	version = "dev"
	commit  = "none"
	date    = "unknown"

	rootCmd = &cobra.Command{
		Use:   "pass-cli",
		Short: "A secure CLI password and API key manager",
		Long: `Pass-CLI is a secure, cross-platform command-line password and API key manager
designed for developers. It provides local encrypted storage with optional system
keychain integration, allowing developers to securely manage credentials without
relying on cloud services.

Features:
  • AES-256-GCM encryption with PBKDF2 key derivation
  • Native OS keychain integration (Windows Credential Manager, macOS Keychain, Linux Secret Service)
  • Script-friendly output for CI/CD integration
  • Automatic usage tracking
  • Offline-first design with no cloud dependencies

Examples:
  # Initialize a new vault
  pass-cli init

  # Add a new credential
  pass-cli add github

  # Retrieve a credential
  pass-cli get github

  # List all credentials
  pass-cli list

For more information, visit: https://github.com/username/pass-cli`,
		PersistentPreRunE: checkFirstRun,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pass-cli/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// GetVaultPath returns the vault path from config or default
// Exits with error if config validation fails (FR-012)
func GetVaultPath() string {
	// Load config and check validation
	cfg, result := config.Load()

	// FR-012: Validate vault_path during config loading and report errors
	if !result.Valid {
		fmt.Fprintf(os.Stderr, "Configuration validation failed:\n")
		for _, err := range result.Errors {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", err.Field, err.Message)
		}
		fmt.Fprintf(os.Stderr, "\nPlease fix your configuration file and try again.\n")
		os.Exit(1)
	}

	var vaultPath string
	if cfg.VaultPath != "" {
		vaultPath = cfg.VaultPath
	} else {
		// Default vault path
		home, err := os.UserHomeDir()
		if err != nil {
			return ".pass-cli/vault.enc"
		}
		return filepath.Join(home, ".pass-cli", "vault.enc")
	}

	// Expand environment variables
	vaultPath = os.ExpandEnv(vaultPath)

	// Expand ~ prefix
	if strings.HasPrefix(vaultPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return vaultPath // Return as-is if home unknown
		}
		vaultPath = filepath.Join(home, vaultPath[1:])
	}

	// Convert relative to absolute path
	if !filepath.IsAbs(vaultPath) {
		home, err := os.UserHomeDir()
		if err == nil {
			vaultPath = filepath.Join(home, vaultPath)
		}
	}

	return vaultPath
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose || viper.GetBool("verbose")
}

// checkFirstRun detects first-run scenarios and triggers guided initialization
// T065: PersistentPreRunE hook for first-run detection
func checkFirstRun(cmd *cobra.Command, args []string) error {
	// Detect first-run scenario (no longer uses vault flag)
	state := vault.DetectFirstRun(cmd.Name(), "")

	// If guided init should be triggered
	if state.ShouldPrompt {
		// Check if running in TTY
		isTTY := term.IsTerminal(int(os.Stdin.Fd()))

		// Get actual vault path (flag or default)
		actualVaultPath := GetVaultPath()

		// Run guided initialization
		if err := vault.RunGuidedInit(actualVaultPath, isTTY); err != nil {
			return err
		}
	}

	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".pass-cli" (without extension).
		viper.AddConfigPath(home + "/.pass-cli")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
