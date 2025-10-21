package health

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// T018: TestBackupCheck_NoBackups - No *.backup files → Pass status
func TestBackupCheck_NoBackups(t *testing.T) {
	// Create temporary directory with no backup files
	tmpDir := t.TempDir()

	// Create checker
	checker := NewBackupChecker(tmpDir)

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	if result.Status != CheckPass {
		t.Errorf("Expected status %s, got %s", CheckPass, result.Status)
	}
	if result.Name != "backup" {
		t.Errorf("Expected name 'backup', got %s", result.Name)
	}

	details, ok := result.Details.(BackupCheckDetails)
	if !ok {
		t.Fatal("Expected BackupCheckDetails type")
	}
	if details.VaultDir != tmpDir {
		t.Errorf("Expected vault dir %s, got %s", tmpDir, details.VaultDir)
	}
	if len(details.BackupFiles) > 0 {
		t.Errorf("Expected no backup files, got %d", len(details.BackupFiles))
	}
	if details.OldBackups > 0 {
		t.Errorf("Expected 0 old backups, got %d", details.OldBackups)
	}
}

// T019: TestBackupCheck_OldBackup - Backup >24h old → Warning status with age details
func TestBackupCheck_OldBackup(t *testing.T) {
	// Create temporary directory with old backup file
	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "vault.backup")

	// Create backup file
	content := []byte("old backup content")
	if err := os.WriteFile(backupPath, content, 0600); err != nil {
		t.Fatalf("Failed to create test backup: %v", err)
	}

	// Set modification time to 48 hours ago
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(backupPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old modification time: %v", err)
	}

	// Create checker
	checker := NewBackupChecker(tmpDir)

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	if result.Status != CheckWarning {
		t.Errorf("Expected status %s, got %s", CheckWarning, result.Status)
	}
	if result.Message == "" {
		t.Error("Expected message about old backup")
	}
	if result.Recommendation == "" {
		t.Error("Expected recommendation about old backup")
	}

	details, ok := result.Details.(BackupCheckDetails)
	if !ok {
		t.Fatal("Expected BackupCheckDetails type")
	}
	if len(details.BackupFiles) == 0 {
		t.Fatal("Expected at least one backup file")
	}
	if details.OldBackups == 0 {
		t.Error("Expected OldBackups count to be > 0")
	}

	// Verify backup file details
	backup := details.BackupFiles[0]
	if backup.Path != backupPath {
		t.Errorf("Expected path %s, got %s", backupPath, backup.Path)
	}
	if backup.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), backup.Size)
	}
	if backup.AgeHours < 24 {
		t.Errorf("Expected age > 24 hours, got %.2f", backup.AgeHours)
	}
	if backup.Status != "old" && backup.Status != "abandoned" {
		t.Errorf("Expected status 'old' or 'abandoned', got %s", backup.Status)
	}
}
