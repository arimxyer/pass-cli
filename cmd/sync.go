package cmd

import (
	"github.com/spf13/cobra"
)

// syncCmd is the parent command for sync-related subcommands
var syncCmd = &cobra.Command{
	Use:     "sync",
	GroupID: "security",
	Short:   "Manage cloud sync for your vault",
	Long: `Manage cloud synchronization for your pass-cli vault.

Cloud sync uses rclone to synchronize your encrypted vault with cloud storage
providers like Google Drive, Dropbox, OneDrive, and many others.

Prerequisites:
  - rclone must be installed and configured with at least one remote
  - Run 'rclone config' to set up a remote if you haven't already

Examples:
  # Enable sync on an existing vault
  pass-cli sync enable

  # Check sync status (via doctor)
  pass-cli doctor`,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
