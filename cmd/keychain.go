package cmd

import "github.com/spf13/cobra"

// keychainCmd represents the keychain command
var keychainCmd = &cobra.Command{
	Use:     "keychain",
	GroupID: "security",
	Short:   "Manage keychain integration for pass-cli vaults",
	Long: `Manage system keychain integration for pass-cli vaults.

The keychain integration stores your vault master password securely in the
operating system's native credential storage (Windows Credential Manager,
macOS Keychain, or Linux Secret Service).`,
}

func init() {
	rootCmd.AddCommand(keychainCmd)
}
