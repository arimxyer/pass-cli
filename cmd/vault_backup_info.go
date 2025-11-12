package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"pass-cli/internal/storage"
	"pass-cli/internal/vault"
)

var (
	infoVerbose bool
)

var vaultBackupInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "View backup status and information",
	Long: `View all available backups with status, age warnings, and disk usage information.

Displays both automatic backups (vault.enc.backup) and manual backups
(vault.enc.*.manual.backup) with metadata including:
- Backup type (automatic or manual)
- File size
- Creation time and age
- Integrity status
- Which backup would be used for restore

Provides warnings for:
- Backups older than 30 days
- More than 5 manual backups (disk space usage)`,
	Example: `  # View all backups
  pass-cli vault backup info

  # View with detailed information
  pass-cli vault backup info --verbose`,
	Args: cobra.NoArgs,
	RunE: runVaultBackupInfo,
}

func init() {
	vaultBackupCmd.AddCommand(vaultBackupInfoCmd)
	vaultBackupInfoCmd.Flags().BoolVarP(&infoVerbose, "verbose", "v", false, "show detailed backup information")
}

func runVaultBackupInfo(cmd *cobra.Command, args []string) error {
	vaultPath := GetVaultPath()

	if infoVerbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] Vault path: %s\n", vaultPath)
		fmt.Fprintf(os.Stderr, "[VERBOSE] Searching for backups...\n")
	}

	// Initialize vault service to access storage
	vaultService, err := vault.New(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to initialize vault service: %w", err)
	}

	storageService := vaultService.GetStorageService()

	// T063: List all backups
	backups, err := storageService.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	// T065: Handle no backups case
	if len(backups) == 0 {
		fmt.Println("No backups found.")
		fmt.Println("\nCreate a backup with: pass-cli vault backup create")
		return nil
	}

	// T064: Categorize backups by type
	var automaticBackup *storage.BackupInfo
	var manualBackups []storage.BackupInfo

	for i := range backups {
		if backups[i].Type == "automatic" {
			automaticBackup = &backups[i]
		} else {
			manualBackups = append(manualBackups, backups[i])
		}
	}

	// Display header
	fmt.Println("📦 Vault Backup Status")
	fmt.Println()

	// T066: Display automatic backup
	if automaticBackup != nil {
		fmt.Println("Automatic Backup:")
		displayBackup(automaticBackup, infoVerbose)
		fmt.Println()
	}

	// T067: Display manual backups
	if len(manualBackups) > 0 {
		fmt.Printf("Manual Backups (%d total):\n", len(manualBackups))
		for i := range manualBackups {
			fmt.Printf("\n%d. ", i+1)
			displayBackup(&manualBackups[i], infoVerbose)
		}
		fmt.Println()
	}

	// T073: Total backup size
	totalSize := int64(0)
	for _, b := range backups {
		totalSize += b.Size
	}
	fmt.Printf("Total backup size: %s\n", formatSize(totalSize))

	// T074: Restore priority
	newestBackup, err := storageService.FindNewestBackup()
	if err == nil && newestBackup != nil {
		fmt.Printf("\n✓ Restore priority: %s backup (%s)\n", newestBackup.Type, formatAge(time.Since(newestBackup.ModTime)))
	}

	// T072: Warning for old backups
	for _, b := range backups {
		age := time.Since(b.ModTime)
		if age > 30*24*time.Hour && !b.IsCorrupted {
			fmt.Printf("\n⚠️  Warning: Backup is %s old. Consider creating a fresh backup.\n", formatAge(age))
			break // Only show warning once
		}
	}

	// T071: Warning for too many manual backups
	if len(manualBackups) > 5 {
		fmt.Printf("\n⚠️  Warning: %d manual backups found. Consider removing old backups to free disk space.\n", len(manualBackups))
	}

	return nil
}

// displayBackup shows backup information
func displayBackup(b *storage.BackupInfo, verbose bool) {
	// T070: Integrity status
	status := "✓"
	if b.IsCorrupted {
		status = "⚠️"
	}

	// T068, T069: Display age and size
	age := formatAge(time.Since(b.ModTime))
	size := formatSize(b.Size)
	fmt.Printf("%s %s ago, %s", status, age, size)

	// T075: Verbose mode shows full path
	if verbose {
		fmt.Printf("\n   Path: %s", b.Path)
		fmt.Printf("\n   Modified: %s", b.ModTime.Format("2006-01-02 15:04:05"))
	}
}

// T068: Format age as human-readable duration
func formatAge(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	if days < 7 {
		return fmt.Sprintf("%d days", days)
	}
	weeks := days / 7
	if weeks == 1 {
		return "1 week"
	}
	if weeks < 4 {
		return fmt.Sprintf("%d weeks", weeks)
	}
	months := days / 30
	if months == 1 {
		return "1 month"
	}
	return fmt.Sprintf("%d months", months)
}

// T069: Format size as human-readable bytes
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
