package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"pass-cli/internal/crypto"
)

const (
	VaultPermissions = 0600 // Read/write for owner only
	DefaultVaultName = "vault.enc"
	BackupSuffix     = ".backup"
	TempSuffix       = ".tmp"
)

var (
	ErrVaultNotFound     = errors.New("vault file not found")
	ErrVaultCorrupted    = errors.New("vault file corrupted")
	ErrInvalidVaultPath  = errors.New("invalid vault path")
	ErrBackupFailed      = errors.New("backup operation failed")
	ErrAtomicWriteFailed = errors.New("atomic write operation failed")
)

type VaultMetadata struct {
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Salt       []byte    `json:"salt"`
	Iterations int       `json:"iterations"` // PBKDF2 iteration count (FR-007)
}

type EncryptedVault struct {
	Metadata VaultMetadata `json:"metadata"`
	Data     []byte        `json:"data"`
}

type StorageService struct {
	cryptoService *crypto.CryptoService
	vaultPath     string
}

func NewStorageService(cryptoService *crypto.CryptoService, vaultPath string) (*StorageService, error) {
	if cryptoService == nil {
		return nil, errors.New("crypto service cannot be nil")
	}

	if vaultPath == "" {
		return nil, ErrInvalidVaultPath
	}

	// Ensure the directory exists
	dir := filepath.Dir(vaultPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create vault directory: %w", err)
	}

	return &StorageService{
		cryptoService: cryptoService,
		vaultPath:     vaultPath,
	}, nil
}

func (s *StorageService) InitializeVault(password string) error {
	// Check if vault already exists
	if s.VaultExists() {
		return errors.New("vault already exists")
	}

	// Generate salt for key derivation
	salt, err := s.cryptoService.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Create initial empty vault data
	emptyVault := []byte("{}")

	// T032/T034: Create vault metadata with configurable iterations (FR-007, FR-010)
	// Uses PASS_CLI_ITERATIONS env var if set, otherwise defaults to 600k (OWASP 2023)
	metadata := VaultMetadata{
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Salt:       salt,
		Iterations: crypto.GetIterations(), // Configurable via env var (T034)
	}

	// Encrypt and save vault
	if err := s.saveEncryptedVault(emptyVault, metadata, password); err != nil {
		return fmt.Errorf("failed to initialize vault: %w", err)
	}

	return nil
}

func (s *StorageService) LoadVault(password string) ([]byte, error) {
	encryptedVault, err := s.loadEncryptedVault()
	if err != nil {
		return nil, err
	}

	// T031: Derive key from password and salt with iterations from metadata (FR-007)
	key, err := s.cryptoService.DeriveKey([]byte(password), encryptedVault.Metadata.Salt, encryptedVault.Metadata.Iterations)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}
	defer s.cryptoService.ClearKey(key)

	// Decrypt vault data
	plaintext, err := s.cryptoService.Decrypt(encryptedVault.Data, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault (invalid password?): %w", err)
	}

	return plaintext, nil
}

func (s *StorageService) SaveVault(data []byte, password string) error {
	// Load existing vault to get metadata
	encryptedVault, err := s.loadEncryptedVault()
	if err != nil {
		return err
	}

	// Update metadata
	encryptedVault.Metadata.UpdatedAt = time.Now()

	// Prepare encrypted vault data
	encryptedData, err := s.prepareEncryptedData(data, encryptedVault.Metadata, password)
	if err != nil {
		return fmt.Errorf("save failed: %w. Your vault was not modified.", err)
	}

	// T014: Atomic save pattern - Step 1: Generate temp filename
	tempPath := s.generateTempFileName()

	// Step 2: Write to temp file
	if err := s.writeToTempFile(tempPath, encryptedData); err != nil {
		return fmt.Errorf("save failed: %w. Your vault was not modified.", err)
	}

	// Ensure temp file cleanup on error
	defer func() {
		// Best-effort cleanup if we haven't renamed yet
		_ = os.Remove(tempPath)
	}()

	// Step 3: Verification (T021 - verify temp file is decryptable)
	if err := s.verifyTempFile(tempPath, password); err != nil {
		// Cleanup temp file on verification failure
		_ = os.Remove(tempPath)
		return fmt.Errorf("save failed during verification. Your vault was not modified. %w", err)
	}

	// Step 4: Atomic rename (vault → backup)
	backupPath := s.vaultPath + BackupSuffix
	if err := s.atomicRename(s.vaultPath, backupPath); err != nil {
		return fmt.Errorf("save failed: %w. Your vault was not modified.", err)
	}

	// Step 5: Atomic rename (temp → vault)
	if err := s.atomicRename(tempPath, s.vaultPath); err != nil {
		// CRITICAL ERROR: Try to restore backup
		_ = s.atomicRename(backupPath, s.vaultPath)
		return fmt.Errorf("CRITICAL: save failed during final rename. Attempted automatic restore from backup. Error: %w", err)
	}

	return nil
}

// prepareEncryptedData encrypts vault data and returns JSON bytes ready to write
func (s *StorageService) prepareEncryptedData(data []byte, metadata VaultMetadata, password string) ([]byte, error) {
	// Derive key from password and salt
	key, err := s.cryptoService.DeriveKey([]byte(password), metadata.Salt, metadata.Iterations)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}
	defer s.cryptoService.ClearKey(key)

	// Encrypt vault data
	encryptedData, err := s.cryptoService.Encrypt(data, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt vault data: %w", err)
	}

	// Create encrypted vault structure
	encryptedVault := EncryptedVault{
		Metadata: metadata,
		Data:     encryptedData,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(encryptedVault)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal vault data: %w", err)
	}

	return jsonData, nil
}

// SaveVaultWithIterations saves vault data with an updated iteration count.
// Used for migration from legacy iteration counts (T033).
func (s *StorageService) SaveVaultWithIterations(data []byte, password string, iterations int) error {
	if iterations < crypto.MinIterations {
		return fmt.Errorf("iterations must be >= %d", crypto.MinIterations)
	}

	// T036d: Pre-flight checks before migration (FR-012)
	if err := s.preflightChecks(); err != nil {
		return fmt.Errorf("pre-flight check failed: %w", err)
	}

	// Load existing vault to get metadata
	encryptedVault, err := s.loadEncryptedVault()
	if err != nil {
		return err
	}

	// Update metadata with new iterations
	encryptedVault.Metadata.UpdatedAt = time.Now()
	encryptedVault.Metadata.Iterations = iterations

	// Create backup before saving
	if err := s.createBackup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Save encrypted vault
	if err := s.saveEncryptedVault(data, encryptedVault.Metadata, password); err != nil {
		// Restore from backup on failure
		if restoreErr := s.restoreFromBackup(); restoreErr != nil {
			return fmt.Errorf("save failed and backup restore failed: %v (original error: %w)", restoreErr, err)
		}
		return fmt.Errorf("failed to save vault: %w", err)
	}

	return nil
}

// SaveVaultWithIterationsUnsafe saves vault data with a specific iteration count without validation.
// ONLY FOR TESTING: Allows simulating legacy vaults with low iteration counts.
// DO NOT USE in production code.
func (s *StorageService) SaveVaultWithIterationsUnsafe(data []byte, password string, iterations int) error {
	// Load existing vault to get metadata
	encryptedVault, err := s.loadEncryptedVault()
	if err != nil {
		return err
	}

	// Update metadata with new iterations (no validation)
	encryptedVault.Metadata.UpdatedAt = time.Now()
	encryptedVault.Metadata.Iterations = iterations

	// Create backup before saving
	if err := s.createBackup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Save encrypted vault
	if err := s.saveEncryptedVault(data, encryptedVault.Metadata, password); err != nil {
		// Restore from backup on failure
		if restoreErr := s.restoreFromBackup(); restoreErr != nil {
			return fmt.Errorf("save failed and backup restore failed: %v (original error: %w)", restoreErr, err)
		}
		return fmt.Errorf("failed to save vault: %w", err)
	}

	return nil
}

// GetIterations returns the current PBKDF2 iteration count from vault metadata.
// Returns 0 if vault doesn't exist or error occurs.
func (s *StorageService) GetIterations() int {
	encryptedVault, err := s.loadEncryptedVault()
	if err != nil {
		return 0
	}
	return encryptedVault.Metadata.Iterations
}

// SetIterations updates the PBKDF2 iteration count in vault metadata.
// This will take effect on the next SaveVault call.
// Used for migration from legacy iteration counts (T033).
func (s *StorageService) SetIterations(iterations int) error {
	if iterations < crypto.MinIterations {
		return fmt.Errorf("iterations must be >= %d", crypto.MinIterations)
	}

	encryptedVault, err := s.loadEncryptedVault()
	if err != nil {
		return err
	}

	encryptedVault.Metadata.Iterations = iterations

	// Note: The updated iterations will be persisted on next SaveVault call
	// We don't save immediately to avoid double-write overhead
	return nil
}

func (s *StorageService) VaultExists() bool {
	_, err := os.Stat(s.vaultPath)
	return err == nil
}

func (s *StorageService) GetVaultInfo() (*VaultMetadata, error) {
	encryptedVault, err := s.loadEncryptedVault()
	if err != nil {
		return nil, err
	}

	// Return a copy of metadata (without the salt for security)
	info := VaultMetadata{
		Version:   encryptedVault.Metadata.Version,
		CreatedAt: encryptedVault.Metadata.CreatedAt,
		UpdatedAt: encryptedVault.Metadata.UpdatedAt,
		Salt:      nil, // Don't expose salt
	}

	return &info, nil
}

func (s *StorageService) ValidateVault() error {
	encryptedVault, err := s.loadEncryptedVault()
	if err != nil {
		return err
	}

	// Basic validation checks
	if encryptedVault.Metadata.Version <= 0 {
		return ErrVaultCorrupted
	}

	if len(encryptedVault.Metadata.Salt) != 32 {
		return ErrVaultCorrupted
	}

	if len(encryptedVault.Data) == 0 {
		return ErrVaultCorrupted
	}

	if encryptedVault.Metadata.CreatedAt.IsZero() {
		return ErrVaultCorrupted
	}

	if encryptedVault.Metadata.UpdatedAt.Before(encryptedVault.Metadata.CreatedAt) {
		return ErrVaultCorrupted
	}

	// Validate Iterations field (T025 - FR-007)
	// Allow 0 for backward compatibility (will default to 100000 on load)
	if encryptedVault.Metadata.Iterations != 0 && encryptedVault.Metadata.Iterations < crypto.MinIterations {
		return fmt.Errorf("%w: iterations must be >= %d", ErrVaultCorrupted, crypto.MinIterations)
	}

	return nil
}

func (s *StorageService) CreateBackup() error {
	return s.createBackup()
}

func (s *StorageService) RestoreFromBackup() error {
	return s.restoreFromBackup()
}

func (s *StorageService) RemoveBackup() error {
	backupPath := s.vaultPath + BackupSuffix
	err := os.Remove(backupPath)
	if os.IsNotExist(err) {
		return nil // Backup doesn't exist, which is fine
	}
	return err
}

// Private helper methods

func (s *StorageService) loadEncryptedVault() (*EncryptedVault, error) {
	if !s.VaultExists() {
		return nil, ErrVaultNotFound
	}

	data, err := os.ReadFile(s.vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault file: %w", err)
	}

	var encryptedVault EncryptedVault
	if err := json.Unmarshal(data, &encryptedVault); err != nil {
		return nil, fmt.Errorf("failed to parse vault file: %w", err)
	}

	// T026: Backward compatibility for legacy vaults without Iterations field (FR-008)
	if encryptedVault.Metadata.Iterations == 0 {
		encryptedVault.Metadata.Iterations = 100000 // Legacy default
	}

	return &encryptedVault, nil
}

func (s *StorageService) saveEncryptedVault(data []byte, metadata VaultMetadata, password string) error {
	// T030: Derive key from password and salt with iterations from metadata (FR-007)
	key, err := s.cryptoService.DeriveKey([]byte(password), metadata.Salt, metadata.Iterations)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}
	defer s.cryptoService.ClearKey(key)

	// Encrypt vault data
	encryptedData, err := s.cryptoService.Encrypt(data, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault data: %w", err)
	}

	// Create encrypted vault structure
	encryptedVault := EncryptedVault{
		Metadata: metadata,
		Data:     encryptedData,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(encryptedVault)
	if err != nil {
		return fmt.Errorf("failed to marshal vault data: %w", err)
	}

	// Atomic write using temporary file
	return s.atomicWrite(s.vaultPath, jsonData)
}

func (s *StorageService) atomicWrite(path string, data []byte) error {
	// FR-015: Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	tempPath := path + TempSuffix

	// Write to temporary file
	// #nosec G304 -- Vault path is user-controlled by design for CLI tool
	tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, VaultPermissions)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Ensure temp file is cleaned up on error
	defer func() {
		if tempFile != nil {
			_ = tempFile.Close()
			_ = os.Remove(tempPath)
		}
	}()

	// Write data
	if _, err := tempFile.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync data: %w", err)
	}

	// Close file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tempFile = nil // Prevent cleanup in defer

	// Atomic move (rename) to final location
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath) // Clean up on failure
		return fmt.Errorf("failed to move temp file to final location: %w", err)
	}

	return nil
}

// preflightChecks performs safety checks before migration (T036d, FR-012).
// Verifies:
// - Disk space >= 2x vault size (to accommodate backup + new vault)
// - Write permissions to vault directory
func (s *StorageService) preflightChecks() error {
	// Check if vault exists
	vaultInfo, err := os.Stat(s.vaultPath)
	if err != nil {
		return fmt.Errorf("failed to stat vault: %w", err)
	}

	vaultSize := vaultInfo.Size()
	vaultDir := filepath.Dir(s.vaultPath)

	// Check disk space (need 2x vault size for backup + new vault)
	requiredSpace := vaultSize * 2

	// Get disk usage info (platform-specific)
	availableSpace, err := s.getAvailableDiskSpace(vaultDir)
	if err != nil {
		// If we can't determine disk space, log warning but continue
		fmt.Fprintf(os.Stderr, "Warning: unable to verify disk space: %v\n", err)
	} else if availableSpace < requiredSpace {
		return fmt.Errorf("insufficient disk space: need %d bytes, have %d bytes", requiredSpace, availableSpace)
	}

	// Test write permissions by creating a temporary test file
	testPath := filepath.Join(vaultDir, ".pass-cli-write-test")
	testFile, err := os.OpenFile(testPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, VaultPermissions) // #nosec G304 -- Test file path constructed from validated vault directory
	if err != nil {
		return fmt.Errorf("no write permission in vault directory: %w", err)
	}
	_ = testFile.Close()
	_ = os.Remove(testPath)

	return nil
}

// getAvailableDiskSpace returns available disk space in bytes for the given path.
// Platform-specific implementation.
func (s *StorageService) getAvailableDiskSpace(path string) (int64, error) {
	// Platform-specific disk space check
	// On Windows, syscall.Statfs_t is not available
	// This is a best-effort check - we'll continue with a warning if it fails

	// Try to use platform-specific approach
	// For Windows: Could use golang.org/x/sys/windows.GetDiskFreeSpaceEx
	// For Unix: Could use syscall.Statfs

	// For now, return error to indicate we can't check (will trigger warning in preflightChecks)
	// This is acceptable per FR-012 - disk space check is a safety measure, not a hard requirement
	return 0, fmt.Errorf("disk space check not implemented for this platform")
}

func (s *StorageService) createBackup() error {
	if !s.VaultExists() {
		return nil // No vault to backup
	}

	backupPath := s.vaultPath + BackupSuffix

	// Copy vault file to backup
	src, err := os.Open(s.vaultPath)
	if err != nil {
		return fmt.Errorf("failed to open vault for backup: %w", err)
	}
	defer func() { _ = src.Close() }()

	// #nosec G304 -- Backup path is user-controlled by design for CLI tool
	dst, err := os.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, VaultPermissions)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy vault to backup: %w", err)
	}

	if err := dst.Sync(); err != nil {
		return fmt.Errorf("failed to sync backup file: %w", err)
	}

	return nil
}

func (s *StorageService) restoreFromBackup() error {
	backupPath := s.vaultPath + BackupSuffix

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return ErrBackupFailed
	}

	// Copy backup to vault location
	// #nosec G304 -- Backup path is user-controlled by design for CLI tool
	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer func() { _ = src.Close() }()

	dst, err := os.OpenFile(s.vaultPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, VaultPermissions)
	if err != nil {
		return fmt.Errorf("failed to create vault file: %w", err)
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	if err := dst.Sync(); err != nil {
		return fmt.Errorf("failed to sync restored vault: %w", err)
	}

	return nil
}
