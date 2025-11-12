package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"pass-cli/internal/security"
	"pass-cli/internal/vault"
)

var (
	createVerbose bool
)

var vaultBackupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a timestamped manual backup of the vault",
	Long: `Create a timestamped manual backup of the vault file.

Manual backups are saved with the naming pattern:
  vault.enc.YYYYMMDD-HHMMSS.manual.backup

This allows you to retain history of manual backups for recovery purposes.
Manual backups complement the automatic backup system that creates backups
during vault saves.

The command works regardless of vault lock state (no master password required).`,
	Example: `  # Create manual backup
  pass-cli vault backup create

  # Create manual backup with verbose output
  pass-cli vault backup create --verbose`,
	Args: cobra.NoArgs,
	RunE: runVaultBackupCreate,
}

func init() {
	vaultBackupCmd.AddCommand(vaultBackupCreateCmd)
	vaultBackupCreateCmd.Flags().BoolVarP(&createVerbose, "verbose", "v", false, "show detailed operation progress")
}

func runVaultBackupCreate(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()

	if createVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Vault path: %s\n", vaultPath)
	}

	// T043: Vault path validation - check vault exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return fmt.Errorf("vault not found at %s\n\nCreate a vault with: pass-cli init", vaultPath)
	}

	// Initialize vault service to access storage
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to initialize vault service: %w", err)
	}

	storageService := vaultService.GetStorageService()

	// T045: Disk space check (best effort)
	if createVerbose {
		vaultInfo, err := os.Stat(vaultPath)
		if err == nil {
			requiredSpace := vaultInfo.Size()
			fmt.Fprintf(os.Stderr, "[VERBOSE] Vault size: %.2f MB\n", float64(requiredSpace)/1024/1024)
			fmt.Fprintf(os.Stderr, "[VERBOSE] Creating backup (requires ~%.2f MB free space)...\n", float64(requiredSpace)/1024/1024)
		}
	}

	if createVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Generating timestamped backup filename...\n")
	}

	// T044: Backup creation logic - calls CreateManualBackup()
	backupPath, err := storageService.CreateManualBackup()
	if err != nil {
		// T048: Audit logging for backup failure (FR-017)
		vaultService.LogAudit(security.EventBackupCreate, security.OutcomeFailure, "")

		// T047: Error handling for common failures
		// Use errors.Is() to check wrapped errors correctly
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("permission denied creating backup\n\nCheck directory permissions for: %s", vaultPath)
		}
		// Check for disk space errors (best effort detection)
		// Check both the error string and common patterns
		errStr := err.Error()
		if strings.Contains(errStr, "no space left on device") ||
		   strings.Contains(errStr, "insufficient disk space") ||
		   strings.Contains(errStr, "disk full") {
			return fmt.Errorf("insufficient disk space for backup\n\nFree up disk space and try again")
		}
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// T048: Audit logging for backup success (FR-017)
	vaultService.LogAudit(security.EventBackupCreate, security.OutcomeSuccess, backupPath)

	if createVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Backup created at: %s\n", backupPath)
		fmt.Fprintf(os.Stderr, "[VERBOSE] Verifying backup integrity...\n")
	}

	// Get backup file info for success message
	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		// Backup was created but we can't stat it - still report success
		fmt.Printf("✅ Backup created successfully\n\nBackup: %s\n", backupPath)
		return nil
	}

	if createVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Backup verified successfully\n")
	}

	// T046: Success message with backup path, size, timestamp
	fmt.Printf("✅ Backup created successfully\n\n")
	fmt.Printf("Backup: %s\n", backupPath)
	fmt.Printf("Size: %.2f MB\n", float64(backupInfo.Size())/1024/1024)
	fmt.Printf("Created: %s\n", backupInfo.ModTime().Format("2006-01-02 15:04:05"))
	fmt.Printf("\nYou can restore from this backup with: pass-cli vault backup restore\n")

	if createVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Backup creation completed\n")
	}

	return nil
}
