package health

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupChecker checks backup file status
type BackupChecker struct {
	vaultDir string
}

// NewBackupChecker creates a new backup checker
func NewBackupChecker(vaultDir string) HealthChecker {
	return &BackupChecker{
		vaultDir: vaultDir,
	}
}

// Name returns the check name
func (b *BackupChecker) Name() string {
	return "backup"
}

// Run executes the backup check
func (b *BackupChecker) Run(ctx context.Context) CheckResult {
	details := BackupCheckDetails{
		VaultDir:    b.vaultDir,
		BackupFiles: []BackupFile{},
		OldBackups:  0,
	}

	// Find all *.backup files in vault directory
	pattern := filepath.Join(b.vaultDir, "*.backup")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return CheckResult{
			Name:    b.Name(),
			Status:  CheckPass,
			Message: fmt.Sprintf("Unable to check backups: %v", err),
			Details: details,
		}
	}

	// No backups found - this is OK
	if len(matches) == 0 {
		return CheckResult{
			Name:    b.Name(),
			Status:  CheckPass,
			Message: "No backup files found",
			Details: details,
		}
	}

	// Analyze each backup file
	now := time.Now()
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue // Skip if can't stat
		}

		modTime := info.ModTime()
		ageHours := now.Sub(modTime).Hours()

		status := "recent"
		if ageHours > 168 { // 1 week
			status = "abandoned"
			details.OldBackups++
		} else if ageHours > 24 {
			status = "old"
			details.OldBackups++
		}

		backupFile := BackupFile{
			Path:       path,
			Size:       info.Size(),
			ModifiedAt: modTime,
			AgeHours:   ageHours,
			Status:     status,
		}
		details.BackupFiles = append(details.BackupFiles, backupFile)
	}

	// Determine check status
	if details.OldBackups > 0 {
		message := fmt.Sprintf("%d old backup file(s) found", details.OldBackups)
		recommendation := "Review and clean up old backup files"
		if details.OldBackups == 1 {
			message = "1 old backup file found"
		}

		return CheckResult{
			Name:           b.Name(),
			Status:         CheckWarning,
			Message:        message,
			Recommendation: recommendation,
			Details:        details,
		}
	}

	return CheckResult{
		Name:    b.Name(),
		Status:  CheckPass,
		Message: fmt.Sprintf("%d recent backup file(s)", len(details.BackupFiles)),
		Details: details,
	}
}
