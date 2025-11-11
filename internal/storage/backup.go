package storage

import (
	"fmt"
	"path/filepath"
	"time"
)

// backup.go contains backup-related functionality for manual vault backups.
// This includes creating timestamped manual backups, listing available backups,
// finding the newest backup, and verifying backup integrity.

// ManualBackupSuffix is the file extension for manual backups
const ManualBackupSuffix = ".manual.backup"

// BackupInfo represents metadata about a single backup file (automatic or manual).
// Purpose: Provide structured information about backup files for listing, sorting,
// and restore priority determination.
type BackupInfo struct {
	Path        string    // Absolute file path to backup file
	ModTime     time.Time // File modification timestamp (used for restore priority)
	Size        int64     // File size in bytes
	Type        string    // "automatic" or "manual"
	IsCorrupted bool      // Integrity check result
}

// generateManualBackupPath generates a timestamped filename for manual backups.
// Format: vault.enc.YYYYMMDD-HHMMSS.manual.backup
// Example: vault.enc.20251111-143022.manual.backup
func (s *StorageService) generateManualBackupPath() string {
	timestamp := time.Now().Format("20060102-150405")
	baseDir := filepath.Dir(s.vaultPath)
	baseName := filepath.Base(s.vaultPath)
	return filepath.Join(baseDir, fmt.Sprintf("%s.%s%s", baseName, timestamp, ManualBackupSuffix))
}

// CreateManualBackup creates a timestamped manual backup of the vault file.
// Returns the absolute path to the created backup file.
// Creates backup directory if missing (FR-018).
func (s *StorageService) CreateManualBackup() (string, error) {
	// Implementation in Phase 2: Foundational (T007)
	return "", fmt.Errorf("not yet implemented")
}

// ListBackups discovers and returns all backup files (automatic + manual).
// Returns BackupInfo slice sorted by modification time (newest first).
func (s *StorageService) ListBackups() ([]BackupInfo, error) {
	// Implementation in Phase 2: Foundational (T008)
	return nil, fmt.Errorf("not yet implemented")
}

// FindNewestBackup returns the most recent backup (automatic or manual).
// Performs integrity check before returning.
// Returns nil if no valid backup exists.
func (s *StorageService) FindNewestBackup() (*BackupInfo, error) {
	// Implementation in Phase 2: Foundational (T009)
	return nil, fmt.Errorf("not yet implemented")
}

// verifyBackupIntegrity performs lightweight integrity check on backup file.
// Checks file existence, non-zero size, and readable header.
// Does not decrypt entire file (performance optimization).
func (s *StorageService) verifyBackupIntegrity(backupPath string) error {
	// Implementation in Phase 2: Foundational (T011)
	return fmt.Errorf("not yet implemented")
}
