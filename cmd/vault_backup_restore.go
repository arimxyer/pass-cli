package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/arimxyer/pass-cli/internal/security"
	"github.com/arimxyer/pass-cli/internal/storage"
)

var (
	restoreForce       bool
	restoreVerbose     bool
	restoreDryRun      bool
	restoreFile        string
	restoreInteractive bool
)

var vaultBackupRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore vault from a backup",
	Long: `Restore vault from a backup file (automatic or manual).

By default, the system selects the newest valid backup. Use --file to specify
a specific backup, or --interactive to choose from a list.

⚠️  WARNING: This command will overwrite your current vault file with the backup.
After restore, you must use the backup's master password to unlock the vault.
If you changed your password since the backup was created, use the OLD password.`,
	Example: `  # Restore from newest backup (with confirmation prompt)
  pass-cli vault backup restore

  # Restore from a specific backup file
  pass-cli vault backup restore --file vault.enc.20241210-143022.manual.backup

  # Interactive mode: choose from available backups
  pass-cli vault backup restore --interactive

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
	vaultBackupRestoreCmd.Flags().StringVar(&restoreFile, "file", "", "restore from a specific backup file")
	vaultBackupRestoreCmd.Flags().BoolVarP(&restoreInteractive, "interactive", "i", false, "interactively select a backup to restore")
}

func runVaultBackupRestore(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()
	logVerbose(restoreVerbose, "Vault path: %s", vaultPath)

	// Initialize storage service to access backup methods
	vaultService, err := initVaultAndStorage(vaultPath)
	if err != nil {
		return err
	}

	storageService := vaultService.GetStorageService()

	// Determine which backup to restore
	var selectedBackup *storage.BackupInfo

	if restoreFile != "" && restoreInteractive {
		return fmt.Errorf("cannot use both --file and --interactive flags")
	}

	if restoreFile != "" {
		// Use specified file
		logVerbose(restoreVerbose, "Using specified backup file: %s", restoreFile)
		info, err := os.Stat(restoreFile)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("backup file not found: %s", restoreFile)
			}
			return fmt.Errorf("failed to access backup file: %w", err)
		}
		// Determine backup type from filename
		backupType := storage.BackupTypeManual
		if strings.HasSuffix(restoreFile, ".backup") && !strings.Contains(restoreFile, ".manual.") {
			backupType = storage.BackupTypeAutomatic
		}
		selectedBackup = &storage.BackupInfo{
			Path:    restoreFile,
			ModTime: info.ModTime(),
			Size:    info.Size(),
			Type:    backupType,
		}
	} else if restoreInteractive {
		// Interactive selection
		logVerbose(restoreVerbose, "Entering interactive mode...")
		backups, err := storageService.ListBackups()
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}
		if len(backups) == 0 {
			return fmt.Errorf("no backups available\n\nNo backup files found. Create a backup with: pass-cli vault backup create")
		}
		selectedBackup, err = selectBackupInteractively(backups)
		if err != nil {
			return err
		}
	} else {
		// Default: find newest backup
		logVerbose(restoreVerbose, "Searching for newest backup...")
		newestBackup, err := storageService.FindNewestBackup()
		if err != nil {
			return fmt.Errorf("failed to find backups: %w", err)
		}
		if newestBackup == nil {
			return fmt.Errorf("no backup available\n\nNo backup files found. Create a backup with: pass-cli vault backup create")
		}
		selectedBackup = newestBackup
	}

	logVerbose(restoreVerbose, "Selected backup: %s", selectedBackup.Path)
	logVerbose(restoreVerbose, "Backup type: %s", selectedBackup.Type)
	logVerbose(restoreVerbose, "Backup size: %s", formatSize(selectedBackup.Size))
	logVerbose(restoreVerbose, "Backup modified: %s", selectedBackup.ModTime.Format("2006-01-02 15:04:05"))

	// Dry-run mode: show what would be restored
	if restoreDryRun {
		fmt.Printf("Dry-run mode: No changes will be made\n\n")
		fmt.Printf("Would restore from:\n")
		fmt.Printf("  Backup: %s\n", selectedBackup.Path)
		fmt.Printf("  Type: %s\n", selectedBackup.Type)
		fmt.Printf("  Size: %s\n", formatSize(selectedBackup.Size))
		fmt.Printf("  Modified: %s\n", selectedBackup.ModTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("\nTo actually restore, run without --dry-run flag.\n")
		return nil
	}

	// Confirmation prompt (unless --force)
	if !restoreForce {
		fmt.Printf("⚠️  Warning: This will overwrite your current vault with the backup.\n\n")
		fmt.Printf("Backup to restore:\n")
		fmt.Printf("  File: %s\n", selectedBackup.Path)
		fmt.Printf("  Type: %s\n", selectedBackup.Type)
		fmt.Printf("  Modified: %s (%s ago)\n", selectedBackup.ModTime.Format("2006-01-02 15:04:05"), formatAge(time.Since(selectedBackup.ModTime)))
		fmt.Printf("\n📌 Note: After restore, use the backup's master password to unlock.\n")
		fmt.Printf("   If you changed your password since this backup, use the OLD password.\n")
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
	if err := storageService.RestoreFromBackup(selectedBackup.Path); err != nil {
		// T030: Audit logging for restore failure (FR-017)
		vaultService.LogAudit(security.EventBackupRestore, security.OutcomeFailure, selectedBackup.Path)
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// T030: Audit logging for restore success (FR-017)
	vaultService.LogAudit(security.EventBackupRestore, security.OutcomeSuccess, selectedBackup.Path)

	logVerbose(restoreVerbose, "Backup copied to vault location")
	logVerbose(restoreVerbose, "Verifying vault file permissions...")

	// Verify and set vault file permissions after restore (T028a, FR-014)
	if err := os.Chmod(vaultPath, storage.VaultPermissions); err != nil {
		return fmt.Errorf("failed to set vault permissions: %w", err)
	}

	logVerbose(restoreVerbose, "Vault permissions set to %o", storage.VaultPermissions)

	// Success message
	fmt.Printf("✅ Vault restored successfully from backup\n\n")
	fmt.Printf("Restored from: %s\n", selectedBackup.Path)
	fmt.Printf("Backup type: %s\n", selectedBackup.Type)
	fmt.Printf("\n📌 Remember: Use the backup's master password to unlock your vault.\n")

	logVerbose(restoreVerbose, "Restore operation completed")

	return nil
}

// selectBackupInteractively displays a table of available backups and prompts user to select one
func selectBackupInteractively(backups []storage.BackupInfo) (*storage.BackupInfo, error) {
	fmt.Println("Available backups:")
	fmt.Println()

	// Display backups in a styled table
	var builder strings.Builder
	table := tablewriter.NewWriter(&builder)
	table.Header([]string{"#", "Type", "Age", "Size", "Modified"})

	for i, backup := range backups {
		age := formatAge(time.Since(backup.ModTime))
		size := formatSize(backup.Size)
		modified := backup.ModTime.Format("2006-01-02 15:04")
		status := ""
		if backup.IsCorrupted {
			status = " ⚠️"
		}
		_ = table.Append([]string{
			fmt.Sprintf("%d", i+1),
			backup.Type + status,
			age,
			size,
			modified,
		})
	}

	_ = table.Render()
	fmt.Print(builder.String())
	fmt.Println()

	// Prompt for selection
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Select backup number (1-%d) or 'q' to cancel: ", len(backups))
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read selection: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "q" || input == "Q" {
			return nil, fmt.Errorf("restore cancelled")
		}

		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(backups) {
			fmt.Printf("Invalid selection. Please enter a number between 1 and %d.\n", len(backups))
			continue
		}

		selected := &backups[num-1]
		if selected.IsCorrupted {
			fmt.Printf("⚠️  Warning: This backup may be corrupted. Are you sure? (y/n): ")
			confirm, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("failed to read confirmation: %w", err)
			}
			confirm = strings.TrimSpace(strings.ToLower(confirm))
			if confirm != "y" && confirm != "yes" {
				continue
			}
		}

		return selected, nil
	}
}
