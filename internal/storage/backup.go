package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// backup.go contains backup-related functionality for manual vault backups.
// This includes creating timestamped manual backups, listing available backups,
// finding the newest backup, and verifying backup integrity.

// ManualBackupSuffix is the file extension for manual backups
const ManualBackupSuffix = ".manual.backup"

// Backup type constants
const (
	BackupTypeAutomatic = "automatic"
	BackupTypeManual    = "manual"
)

// minValidVaultSize is the minimum size in bytes for a valid encrypted vault file.
// A valid vault must contain at minimum:
//   - Salt: 32 bytes (for key derivation)
//   - Nonce: 12 bytes (for AES-GCM)
//   - Auth tag: 16 bytes (for AES-GCM authentication)
//   - Encrypted data: minimal JSON structure
//
// Total: ~100 bytes minimum for a structurally valid (but potentially empty) vault
const minValidVaultSize = 100

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
// Format: vault.enc.YYYYMMDD-HHMMSS.manual.backup (timestamp in UTC)
// Example: vault.enc.20251111-143022.manual.backup
func (s *StorageService) generateManualBackupPath() string {
	timestamp := time.Now().UTC().Format("20060102-150405")
	baseDir := filepath.Dir(s.vaultPath)
	baseName := filepath.Base(s.vaultPath)
	return filepath.Join(baseDir, fmt.Sprintf("%s.%s%s", baseName, timestamp, ManualBackupSuffix))
}

// CreateManualBackup creates a timestamped manual backup of the vault file.
// Returns the absolute path to the created backup file.
// Creates backup directory if missing (FR-018).
func (s *StorageService) CreateManualBackup() (string, error) {
	if !s.VaultExists() {
		return "", ErrVaultNotFound
	}

	// Generate timestamped backup path
	backupPath := s.generateManualBackupPath()

	// Ensure backup directory exists (FR-018)
	backupDir := filepath.Dir(backupPath)
	if err := s.fs.MkdirAll(backupDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy vault file to backup using atomic operation
	if err := s.copyFile(s.vaultPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to create manual backup: %w", err)
	}

	return backupPath, nil
}

// ListBackups discovers and returns all backup files (automatic + manual).
// Returns BackupInfo slice sorted by modification time (newest first).
func (s *StorageService) ListBackups() ([]BackupInfo, error) {
	vaultDir := filepath.Dir(s.vaultPath)
	baseName := filepath.Base(s.vaultPath)

	var backups []BackupInfo

	// Pattern 1: Automatic backup (vault.enc.backup)
	automaticPath := filepath.Join(vaultDir, baseName+BackupSuffix)
	if info, err := os.Stat(automaticPath); err == nil {
		backups = append(backups, BackupInfo{
			Path:        automaticPath,
			ModTime:     info.ModTime(),
			Size:        info.Size(),
			Type:        BackupTypeAutomatic,
			IsCorrupted: s.verifyBackupIntegrity(automaticPath) != nil,
		})
	}

	// Pattern 2: Manual backups (vault.enc.*.manual.backup)
	manualPattern := filepath.Join(vaultDir, baseName+".*"+ManualBackupSuffix)
	matches, err := filepath.Glob(manualPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find manual backups: %w", err)
	}

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue // Skip files we can't stat
		}

		backups = append(backups, BackupInfo{
			Path:        match,
			ModTime:     info.ModTime(),
			Size:        info.Size(),
			Type:        BackupTypeManual,
			IsCorrupted: s.verifyBackupIntegrity(match) != nil,
		})
	}

	// Sort by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.After(backups[j].ModTime)
	})

	return backups, nil
}

// FindNewestBackup returns the most recent backup (automatic or manual).
// Performs integrity check before returning.
// Returns nil if no valid backup exists.
func (s *StorageService) FindNewestBackup() (*BackupInfo, error) {
	backups, err := s.ListBackups()
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	// ListBackups already sorts by newest first and checks integrity
	// Find first non-corrupted backup
	for i := range backups {
		if !backups[i].IsCorrupted {
			return &backups[i], nil
		}
	}

	// No valid backups found
	return nil, nil
}

// verifyBackupIntegrity performs structural integrity check on backup file.
// Checks file existence, non-zero size, and validates encrypted JSON structure.
// Does not decrypt data (no password required), but verifies expected fields exist.
func (s *StorageService) verifyBackupIntegrity(backupPath string) error {
	// Check file existence
	info, err := os.Stat(backupPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("backup file not found: %w", err)
		}
		return fmt.Errorf("failed to stat backup file: %w", err)
	}

	// Check non-zero size
	if info.Size() == 0 {
		return fmt.Errorf("backup file is empty")
	}

	// Check minimum size for valid encrypted vault
	if info.Size() < minValidVaultSize {
		return fmt.Errorf("backup file too small (%d bytes, minimum %d bytes required)", info.Size(), minValidVaultSize)
	}

	// Read entire file for structural validation
	data, err := s.fs.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("cannot read backup file: %w", err)
	}

	// Validate vault JSON structure
	if err := validateVaultStructure(data); err != nil {
		return fmt.Errorf("backup file validation failed: %w", err)
	}

	return nil
}

// validateVaultStructure validates that data contains a properly structured encrypted vault.
// Checks JSON format and required fields without decrypting the vault.
func validateVaultStructure(data []byte) error {
	// Parse as JSON to verify structure
	// Expected format: {"metadata": {...}, "data": "..."}
	var vault EncryptedVault
	if err := json.Unmarshal(data, &vault); err != nil {
		return fmt.Errorf("not valid vault JSON: %w", err)
	}

	// Verify required top-level fields exist
	if len(vault.Data) == 0 {
		return fmt.Errorf("missing or empty encrypted data")
	}

	// Verify metadata structure
	if len(vault.Metadata.Salt) == 0 {
		return fmt.Errorf("missing salt in metadata")
	}

	if vault.Metadata.Version == 0 {
		return fmt.Errorf("invalid version in metadata")
	}

	return nil
}

// copyFile copies a file from src to dst with proper permissions.
// Uses the FileSystem abstraction for testability.
func (s *StorageService) copyFile(src, dst string) error {
	// Open source file (read-only)
	srcFile, err := s.fs.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	// Create destination file with vault permissions
	// #nosec G304 -- Backup path is user-controlled by design for CLI tool
	dstFile, err := s.fs.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, VaultPermissions)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	// Copy data
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}
