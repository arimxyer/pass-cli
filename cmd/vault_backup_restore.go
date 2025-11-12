package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"pass-cli/internal/security"
	"pass-cli/internal/storage"
	"pass-cli/internal/vault"
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

	if restoreVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Vault path: %s\n", vaultPath)
		fmt.Fprintf(os.Stderr, "[VERBOSE] Searching for backups...\n")
	}

	// Initialize storage service to access backup methods
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to initialize vault service: %w", err)
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

	if restoreVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Found backup: %s\n", newestBackup.Path)
		fmt.Fprintf(os.Stderr, "[VERBOSE] Backup type: %s\n", newestBackup.Type)
		fmt.Fprintf(os.Stderr, "[VERBOSE] Backup size: %d bytes\n", newestBackup.Size)
		fmt.Fprintf(os.Stderr, "[VERBOSE] Backup modified: %s\n", newestBackup.ModTime.Format("2006-01-02 15:04:05"))
	}

	// Dry-run mode: show what would be restored
	if restoreDryRun {
		fmt.Printf("Dry-run mode: No changes will be made\n\n")
		fmt.Printf("Would restore from:\n")
		fmt.Printf("  Backup: %s\n", newestBackup.Path)
		fmt.Printf("  Type: %s\n", newestBackup.Type)
		fmt.Printf("  Size: %.2f MB\n", float64(newestBackup.Size)/1024/1024)
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

	if restoreVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Starting restore operation...\n")
	}

	// Perform restore
	if err := storageService.RestoreFromBackup(newestBackup.Path); err != nil {
		// T030: Audit logging for restore failure (FR-017)
		vaultService.LogAudit(security.EventBackupRestore, security.OutcomeFailure, newestBackup.Path)
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// T030: Audit logging for restore success (FR-017)
	vaultService.LogAudit(security.EventBackupRestore, security.OutcomeSuccess, newestBackup.Path)

	if restoreVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Backup copied to vault location\n")
		fmt.Fprintf(os.Stderr, "[VERBOSE] Verifying vault file permissions...\n")
	}

	// Verify and set vault file permissions after restore (T028a, FR-014)
	if err := os.Chmod(vaultPath, storage.VaultPermissions); err != nil {
		return fmt.Errorf("failed to set vault permissions: %w", err)
	}

	if restoreVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Vault permissions set to %o\n", storage.VaultPermissions)
	}

	// Success message
	fmt.Printf("✅ Vault restored successfully from backup\n\n")
	fmt.Printf("Restored from: %s\n", newestBackup.Path)
	fmt.Printf("Backup type: %s\n", newestBackup.Type)
	fmt.Printf("\nYou can now unlock your vault with your master password.\n")

	if restoreVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Restore operation completed\n")
	}

	return nil
}
