package storage

import (
	"crypto/rand"
	"fmt"
	"os"
	"time"
)

// generateTempFileName creates unique temp file name with timestamp + random suffix
// Format: vault.enc.tmp.YYYYMMDD-HHMMSS.XXXXXX
func (s *StorageService) generateTempFileName() string {
	timestamp := time.Now().Format("20060102-150405")
	suffix := randomHexSuffix(6)
	return fmt.Sprintf("%s.tmp.%s.%s", s.vaultPath, timestamp, suffix)
}

// randomHexSuffix generates N-character hex suffix from crypto/rand
func randomHexSuffix(length int) string {
	bytes := make([]byte, length/2) // 2 hex chars per byte
	rand.Read(bytes)                 // crypto/rand, not math/rand
	return fmt.Sprintf("%x", bytes)
}

// writeToTempFile writes encrypted data to temporary file with vault permissions (T012)
func (s *StorageService) writeToTempFile(path string, data []byte) error {
	// Create temp file with vault permissions (0600)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, VaultPermissions)
	if err != nil {
		// Check for disk space or permission errors
		if os.IsPermission(err) {
			return fmt.Errorf("%w: %v", ErrPermissionDenied, err)
		}
		// Generic disk space or filesystem error
		return fmt.Errorf("%w: %v", ErrDiskSpaceExhausted, err)
	}
	defer file.Close()

	// Write encrypted vault data
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Force flush to disk before verification (FR-015)
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	return nil
}

// verifyTempFile decrypts and validates temporary file before commit
func (s *StorageService) verifyTempFile(path string, password string) error {
	// TODO: Implement in User Story 2 (T020)
	return nil
}

// atomicRename performs atomic file rename (handles platform differences via os.Rename) (T013)
func (s *StorageService) atomicRename(oldPath, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		// Check for cross-device rename error (filesystem not atomic)
		if os.IsPermission(err) {
			return fmt.Errorf("%w: %v", ErrPermissionDenied, err)
		}
		// Cross-device or other filesystem error
		return fmt.Errorf("%w: %v", ErrFilesystemNotAtomic, err)
	}
	return nil
}

// cleanupTempFile removes temporary file (best-effort, logs warning if fails)
func (s *StorageService) cleanupTempFile(path string) error {
	// TODO: Implement in User Story 4 (T031)
	return nil
}

// cleanupOrphanedTempFiles removes old temp files from crashed previous saves
func (s *StorageService) cleanupOrphanedTempFiles(currentTempPath string) {
	// TODO: Implement in User Story 4 (T032)
}
