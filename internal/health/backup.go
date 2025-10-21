package health

import "context"

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
	// Placeholder - will be implemented in Phase 3
	return CheckResult{
		Name:    b.Name(),
		Status:  CheckPass,
		Message: "Backup check not yet implemented",
		Details: BackupCheckDetails{
			VaultDir:    b.vaultDir,
			BackupFiles: []BackupFile{},
			OldBackups:  0,
		},
	}
}
