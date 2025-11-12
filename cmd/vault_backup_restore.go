package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"pass-cli/internal/security"
	"pass-cli/internal/storage"
)

var (
	restoreForce   bool
	restoreVerbose bool
	restoreDryRun  bool
)

var vaultBackupRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore vault from the most recent backup",
	Long: `Restore vault from the most recent backup file (automatic or manual).

The system automatically selects the newest valid backup based on file modification
time. Automatic backups (vault.enc.backup) and manual backups (vault.enc.*.manual.backup)
are both considered.

⚠️  WARNING: This command will overwrite your current vault file with the backup.
Make sure this is what you want before proceeding.`,
	Example: `  # Restore from newest backup (with confirmation prompt)
  pass-cli vault backup restore

  # Restore without confirmation prompt
  pass-cli vault backup restore --force

  # Preview which backup would be restored (no changes)
  pass-cli vault backup restore --dry-run

  # Restore with detailed progress
  pass-cli vault backup restore --verbose`,
	Args: cobra.NoArgs,
	RunE: runVaultBackupRestore,
}

func init() {
	vaultBackupCmd.AddCommand(vaultBackupRestoreCmd)
	vaultBackupRestoreCmd.Flags().BoolVarP(&restoreForce, "force", "f", false, "skip confirmation prompt")
	vaultBackupRestoreCmd.Flags().BoolVarP(&restoreVerbose, "verbose", "v", false, "show detailed operation progress")
	vaultBackupRestoreCmd.Flags().BoolVar(&restoreDryRun, "dry-run", false, "show which backup would be restored without making changes")
}

func runVaultBackupRestore(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()
	logVerbose(restoreVerbose, "Vault path: %s", vaultPath)
	logVerbose(restoreVerbose, "Searching for backups...")

	// Initialize storage service to access backup methods
	vaultService, err := initVaultAndStorage(vaultPath)
	if err != nil {
		return err
	}

	storageService := vaultService.GetStorageService()

	// Find newest backup
	newestBackup, err := storageService.FindNewestBackup()
	if err != nil {
		return fmt.Errorf("failed to find backups: %w", err)
	}

	if newestBackup == nil {
		return fmt.Errorf("no backup available\n\nNo backup files found. Create a backup with: pass-cli vault backup create")
	}

	logVerbose(restoreVerbose, "Found backup: %s", newestBackup.Path)
	logVerbose(restoreVerbose, "Backup type: %s", newestBackup.Type)
	logVerbose(restoreVerbose, "Backup size: %s", formatSize(newestBackup.Size))
	logVerbose(restoreVerbose, "Backup modified: %s", newestBackup.ModTime.Format("2006-01-02 15:04:05"))

	// Dry-run mode: show what would be restored
	if restoreDryRun {
		fmt.Printf("Dry-run mode: No changes will be made\n\n")
		fmt.Printf("Would restore from:\n")
		fmt.Printf("  Backup: %s\n", newestBackup.Path)
		fmt.Printf("  Type: %s\n", newestBackup.Type)
		fmt.Printf("  Size: %s\n", formatSize(newestBackup.Size))
		fmt.Printf("  Modified: %s\n", newestBackup.ModTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("\nTo actually restore, run without --dry-run flag.\n")
		return nil
	}

	// Confirmation prompt (unless --force)
	if !restoreForce {
		fmt.Printf("⚠️  Warning: This will overwrite your current vault with the backup.\n\n")
		fmt.Printf("Backup to restore:\n")
		fmt.Printf("  File: %s\n", newestBackup.Path)
		fmt.Printf("  Type: %s\n", newestBackup.Type)
		fmt.Printf("  Modified: %s\n", newestBackup.ModTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("\nAre you sure you want to continue? (y/n): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(os.Stderr, "Restore cancelled.")
			os.Exit(1)
		}
		fmt.Println()
	}

	logVerbose(restoreVerbose, "Starting restore operation...")

	// Perform restore
	if err := storageService.RestoreFromBackup(newestBackup.Path); err != nil {
		// T030: Audit logging for restore failure (FR-017)
		vaultService.LogAudit(security.EventBackupRestore, security.OutcomeFailure, newestBackup.Path)
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// T030: Audit logging for restore success (FR-017)
	vaultService.LogAudit(security.EventBackupRestore, security.OutcomeSuccess, newestBackup.Path)

	logVerbose(restoreVerbose, "Backup copied to vault location")
	logVerbose(restoreVerbose, "Verifying vault file permissions...")

	// Verify and set vault file permissions after restore (T028a, FR-014)
	if err := os.Chmod(vaultPath, storage.VaultPermissions); err != nil {
		return fmt.Errorf("failed to set vault permissions: %w", err)
	}

	logVerbose(restoreVerbose, "Vault permissions set to %o", storage.VaultPermissions)

	// Success message
	fmt.Printf("✅ Vault restored successfully from backup\n\n")
	fmt.Printf("Restored from: %s\n", newestBackup.Path)
	fmt.Printf("Backup type: %s\n", newestBackup.Type)
	fmt.Printf("\nYou can now unlock your vault with your master password.\n")

	logVerbose(restoreVerbose, "Restore operation completed")

	return nil
}
