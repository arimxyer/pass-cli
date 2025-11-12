package cmd

import "github.com/spf13/cobra"

// vaultBackupCmd represents the vault backup parent command
var vaultBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage vault backup and restore operations",
	Long: `Manage vault backup and restore operations.

Create manual backups before risky operations, restore from backup when vault
is corrupted or deleted, and view backup status and history.

Manual backups complement the automatic backup system that creates backups
during vault saves. Manual backups are timestamped and retain history, while
automatic backups use an N-1 strategy (single .backup file).`,
}

func init() {
	vaultCmd.AddCommand(vaultBackupCmd)
}
