package storage

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	// crypto/rand.Read always returns len(bytes), nil in practice
	// Only returns error if Reader fails (extremely rare, indicates system issue)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based suffix if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%x", bytes)
}

// writeToTempFile writes encrypted data to temporary file with vault permissions (T012)
func (s *StorageService) writeToTempFile(path string, data []byte) error {
	// Create temp file with vault permissions (0600)
	// #nosec G304 -- Temp file path is generated internally with timestamp+random suffix
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, VaultPermissions)
	if err != nil {
		// Check for disk space or permission errors
		if os.IsPermission(err) {
			return fmt.Errorf("%w: %v", ErrPermissionDenied, err)
		}
		// Generic disk space or filesystem error
		return fmt.Errorf("%w: %v", ErrDiskSpaceExhausted, err)
	}
	defer func() {
		// Close is best-effort - file already synced, data is on disk
		_ = file.Close()
	}()

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

// verifyTempFile decrypts and validates temporary file before commit (T020)
func (s *StorageService) verifyTempFile(path string, password string) error {
	// Read temp file
	// #nosec G304 -- Temp file path is generated internally with timestamp+random suffix
	data, err := s.fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%w: cannot read temporary file: %v", ErrVerificationFailed, err)
	}

	// Parse as EncryptedVault structure
	var encryptedVault EncryptedVault
	if err := json.Unmarshal(data, &encryptedVault); err != nil {
		return fmt.Errorf("%w: invalid vault structure: %v", ErrVerificationFailed, err)
	}

	// Derive key from password and salt
	key, err := s.cryptoService.DeriveKey([]byte(password), encryptedVault.Metadata.Salt, encryptedVault.Metadata.Iterations)
	if err != nil {
		return fmt.Errorf("%w: failed to derive key: %v", ErrVerificationFailed, err)
	}
	defer s.cryptoService.ClearKey(key)

	// Decrypt vault data (in-memory verification)
	decryptedData, err := s.cryptoService.Decrypt(encryptedVault.Data, key)
	if err != nil {
		return fmt.Errorf("%w: encrypted data could not be decrypted: %v", ErrVerificationFailed, err)
	}
	// CRITICAL: Clear decrypted memory immediately after validation
	defer s.cryptoService.ClearData(decryptedData)

	// Verification successful - data decrypts correctly
	// Note: JSON structure validation is the responsibility of the vault layer
	// Storage layer only verifies that data can be decrypted successfully
	return nil
}

// atomicRename performs atomic file rename (handles platform differences via os.Rename) (T013)
func (s *StorageService) atomicRename(oldPath, newPath string) error {
	if err := s.fs.Rename(oldPath, newPath); err != nil {
		// Check for cross-device rename error (filesystem not atomic)
		if os.IsPermission(err) {
			return fmt.Errorf("%w: %v", ErrPermissionDenied, err)
		}
		// Cross-device or other filesystem error
		return fmt.Errorf("%w: %v", ErrFilesystemNotAtomic, err)
	}
	return nil
}

// cleanupTempFile removes temporary file (best-effort, logs warning if fails) (T031)
func (s *StorageService) cleanupTempFile(path string) error {
	if err := s.fs.Remove(path); err != nil && !os.IsNotExist(err) {
		// Log warning but don't fail - cleanup is non-critical
		fmt.Fprintf(os.Stderr, "Warning: failed to remove temporary file %s: %v\n", path, err)
		return err
	}
	return nil
}

// cleanupOrphanedTempFiles removes old temp files from crashed previous saves (T032)
func (s *StorageService) cleanupOrphanedTempFiles(currentTempPath string) {
	vaultDir := filepath.Dir(s.vaultPath)
	pattern := filepath.Join(vaultDir, "*.tmp.*")

	matches, err := s.fs.Glob(pattern)
	if err != nil {
		// Best-effort cleanup, ignore glob errors
		return
	}

	for _, orphan := range matches {
		// Don't delete current temp file
		if orphan == currentTempPath {
			continue
		}

		// Remove orphaned file from previous crashed save
		if err := s.fs.Remove(orphan); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove orphaned temp file %s: %v\n", orphan, err)
		}
	}
}
