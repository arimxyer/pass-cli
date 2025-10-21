package cmd

import "github.com/spf13/cobra"

// vaultCmd represents the vault command
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage pass-cli vault files",
	Long: `Manage pass-cli vault files and their lifecycle.

Available commands:
  remove  - Permanently delete a vault file and its keychain entry`,
}

func init() {
	rootCmd.AddCommand(vaultCmd)
}
