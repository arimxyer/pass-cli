package cmd

import "github.com/spf13/cobra"

// keychainCmd represents the keychain command
var keychainCmd = &cobra.Command{
	Use:   "keychain",
	Short: "Manage keychain integration for pass-cli vaults",
	Long: `Manage system keychain integration for pass-cli vaults.

The keychain integration stores your vault master password securely in the
operating system's native credential storage (Windows Credential Manager,
macOS Keychain, or Linux Secret Service).

Available commands:
  enable  - Enable keychain for an existing vault
  status  - Check keychain integration status`,
}

func init() {
	rootCmd.AddCommand(keychainCmd)
}
