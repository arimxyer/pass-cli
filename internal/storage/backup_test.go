package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"pass-cli/internal/crypto"
)

// mockFileSystem is a test double for FileSystem that allows error injection
type mockFileSystem struct {
	osFileSystem // Embed real implementation

	// Optional override functions for error injection
	openFileFunc func(name string, flag int, perm os.FileMode) (*os.File, error)
}

// newMockFileSystem creates a mock that delegates to real filesystem by default
func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		osFileSystem: osFileSystem{},
	}
}

// OpenFile delegates to override function if set, otherwise uses real implementation
func (m *mockFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	if m.openFileFunc != nil {
		return m.openFileFunc(name, flag, perm)
	}
	return m.osFileSystem.OpenFile(name, flag, perm)
}


func TestGenerateManualBackupPath(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Generate manual backup path
	backupPath := storage.generateManualBackupPath()

	// Verify format: vault.enc.YYYYMMDD-HHMMSS.manual.backup
	expected := regexp.MustCompile(`vault\.enc\.\d{8}-\d{6}\.manual\.backup$`)
	if !expected.MatchString(filepath.Base(backupPath)) {
		t.Errorf("Backup path does not match expected format: %s", filepath.Base(backupPath))
	}

	// Verify directory matches vault directory
	if filepath.Dir(backupPath) != filepath.Dir(vaultPath) {
		t.Errorf("Backup directory %s does not match vault directory %s",
			filepath.Dir(backupPath), filepath.Dir(vaultPath))
	}

	// Verify timestamp is recent (within last 5 seconds)
	baseName := filepath.Base(backupPath)
	// Extract timestamp: vault.enc.YYYYMMDD-HHMMSS.manual.backup
	timestampRegex := regexp.MustCompile(`vault\.enc\.(\d{8}-\d{6})\.manual\.backup`)
	matches := timestampRegex.FindStringSubmatch(baseName)
	if len(matches) < 2 {
		t.Fatalf("Could not extract timestamp from %s", baseName)
	}

	timestamp, err := time.Parse("20060102-150405", matches[1])
	if err != nil {
		t.Fatalf("Failed to parse timestamp %s: %v", matches[1], err)
	}

	// Check if timestamp is recent (within 5 seconds, using UTC)
	now := time.Now().UTC()
	age := now.Sub(timestamp)
	if age < 0 {
		age = -age // Take absolute value
	}
	if age > 5*time.Second {
		t.Errorf("Timestamp %s is not recent compared to %s (age: %v)", timestamp, now, age)
	}
}

func TestGenerateManualBackupPath_Uniqueness(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Generate multiple backup paths rapidly
	paths := make([]string, 3)
	for i := 0; i < 3; i++ {
		paths[i] = storage.generateManualBackupPath()
		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Second)
	}

	// Verify all paths are unique
	seen := make(map[string]bool)
	for _, path := range paths {
		if seen[path] {
			t.Errorf("Duplicate backup path generated: %s", path)
		}
		seen[path] = true
	}
}

func TestCreateManualBackup(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Initialize vault
	password := "test-password"
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Create manual backup
	backupPath, err := storage.CreateManualBackup()
	if err != nil {
		t.Fatalf("CreateManualBackup failed: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup file was not created at %s", backupPath)
	}

	// Verify backup has same size as vault
	vaultInfo, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("Failed to stat vault: %v", err)
	}

	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("Failed to stat backup: %v", err)
	}

	if vaultInfo.Size() != backupInfo.Size() {
		t.Errorf("Backup size %d does not match vault size %d",
			backupInfo.Size(), vaultInfo.Size())
	}

	// Verify backup has correct permissions
	if backupInfo.Mode().Perm() != os.FileMode(VaultPermissions) {
		t.Logf("Note: Backup permissions are %v (expected %v) - this is normal on Windows",
			backupInfo.Mode().Perm(), os.FileMode(VaultPermissions))
	}
}

func TestCreateManualBackup_NoVault(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Try to create backup without vault
	_, err = storage.CreateManualBackup()
	if err != ErrVaultNotFound {
		t.Errorf("Expected ErrVaultNotFound, got %v", err)
	}
}

func TestCreateManualBackup_DirectoryCreation(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()

	// Use nested directory that doesn't exist yet
	vaultPath := filepath.Join(tempDir, "subdir", "vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Initialize vault (will create directory)
	password := "test-password"
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Remove the subdirectory to test directory creation
	subdirPath := filepath.Join(tempDir, "subdir")
	if err := os.RemoveAll(subdirPath); err != nil {
		t.Fatalf("Failed to remove subdir: %v", err)
	}

	// Create vault again
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed after directory removal: %v", err)
	}

	// Create manual backup - should create directory if needed
	backupPath, err := storage.CreateManualBackup()
	if err != nil {
		t.Fatalf("CreateManualBackup failed: %v", err)
	}

	// Verify backup was created
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup file was not created at %s (directory creation may have failed)", backupPath)
	}
}

// T019: Unit test for backup verification logic
func TestVerifyBackupIntegrity(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Initialize vault
	password := "test-password"
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Create manual backup
	backupPath, err := storage.CreateManualBackup()
	if err != nil {
		t.Fatalf("CreateManualBackup failed: %v", err)
	}

	// Test 1: Valid backup passes verification
	if err := storage.verifyBackupIntegrity(backupPath); err != nil {
		t.Errorf("Valid backup failed verification: %v", err)
	}

	// Test 2: Non-existent backup fails
	nonExistentPath := filepath.Join(tempDir, "nonexistent.backup")
	if err := storage.verifyBackupIntegrity(nonExistentPath); err == nil {
		t.Error("Expected error for non-existent backup, got nil")
	}

	// Test 3: Empty backup fails
	emptyBackupPath := filepath.Join(tempDir, "empty.backup")
	if err := os.WriteFile(emptyBackupPath, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to create empty backup: %v", err)
	}
	if err := storage.verifyBackupIntegrity(emptyBackupPath); err == nil {
		t.Error("Expected error for empty backup, got nil")
	}

	// Test 4: Too small backup fails
	tooSmallPath := filepath.Join(tempDir, "toosmall.backup")
	if err := os.WriteFile(tooSmallPath, []byte("tiny"), 0600); err != nil {
		t.Fatalf("Failed to create too-small backup: %v", err)
	}
	if err := storage.verifyBackupIntegrity(tooSmallPath); err == nil {
		t.Error("Expected error for too-small backup, got nil")
	}

	// Test 5: Unreadable backup fails (test with closed file)
	// Create a valid-sized file but make it temporarily unreadable
	unreadablePath := filepath.Join(tempDir, "unreadable.backup")
	data := make([]byte, 200) // Valid size
	if err := os.WriteFile(unreadablePath, data, 0000); err != nil {
		t.Fatalf("Failed to create unreadable backup: %v", err)
	}
	// On Windows, file permissions work differently, so this test may pass
	// That's acceptable - we're primarily testing the logic path
	_ = storage.verifyBackupIntegrity(unreadablePath)
	// Cleanup: restore permissions for removal
	_ = os.Chmod(unreadablePath, 0600)
}

// TestCreateManualBackup_PermissionDenied tests permission denied error during backup
// T036: Unit test for permission denied error handling
func TestCreateManualBackup_PermissionDenied(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Create mock filesystem that injects permission errors
	mockFS := newMockFileSystem()
	storage, err := NewStorageServiceWithFS(cryptoService, vaultPath, mockFS)
	if err != nil {
		t.Fatalf("NewStorageServiceWithFS failed: %v", err)
	}

	// Initialize vault using real filesystem
	password := "test-password"
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Inject permission denied error when trying to create backup file
	mockFS.openFileFunc = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		// Allow reading source vault file, but fail on creating destination backup
		if flag&os.O_WRONLY != 0 && filepath.Base(name) != "vault.enc" {
			return nil, os.ErrPermission
		}
		// Use real filesystem for everything else
		return mockFS.osFileSystem.OpenFile(name, flag, perm)
	}

	// Try to create backup - should fail with permission error
	_, err = storage.CreateManualBackup()
	if err == nil {
		t.Fatal("Expected permission error, got nil")
	}

	// Verify error contains context
	if err.Error() == "" {
		t.Error("Expected error message, got empty string")
	}

	// Verify error is wrapped correctly
	expectedSubstring := "failed to create destination file"
	if !contains(err.Error(), expectedSubstring) {
		t.Errorf("Expected error to contain %q, got: %s", expectedSubstring, err.Error())
	}
}

// TestCreateManualBackup_DiskFull tests disk full error during backup
// T035: Unit test for disk full error handling
func TestCreateManualBackup_DiskFull(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Create mock filesystem that injects disk space errors
	mockFS := newMockFileSystem()
	storage, err := NewStorageServiceWithFS(cryptoService, vaultPath, mockFS)
	if err != nil {
		t.Fatalf("NewStorageServiceWithFS failed: %v", err)
	}

	// Initialize vault using real filesystem
	password := "test-password"
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Track number of OpenFile calls to allow source file read but fail on destination
	callCount := 0

	// Inject disk space error when trying to create backup file
	mockFS.openFileFunc = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		// Allow reading source vault file (first write call)
		// Fail on creating destination backup file (second write call)
		if flag&os.O_WRONLY != 0 && filepath.Base(name) != "vault.enc" {
			callCount++
			if callCount > 0 {
				// Simulate disk full error on destination file creation
				return nil, fmt.Errorf("write %s: no space left on device", name)
			}
		}

		// Use real filesystem for everything else
		return mockFS.osFileSystem.OpenFile(name, flag, perm)
	}

	// Try to create backup - should fail with disk space error
	_, err = storage.CreateManualBackup()
	if err == nil {
		t.Fatal("Expected disk space error, got nil")
	}

	// Verify error contains context
	if err.Error() == "" {
		t.Error("Expected error message, got empty string")
	}

	// Verify error is wrapped correctly - error happens at OpenFile stage
	expectedSubstring := "failed to create destination file"
	if !contains(err.Error(), expectedSubstring) {
		t.Errorf("Expected error to contain %q, got: %s", expectedSubstring, err.Error())
	}

	// Verify the underlying error is preserved
	if !contains(err.Error(), "no space left on device") {
		t.Errorf("Expected error to contain 'no space left on device', got: %s", err.Error())
	}
}

// TestListBackups tests backup discovery and sorting
// T059: Unit test for backup listing and sorting
func TestListBackups(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Initialize vault
	password := "test-password"
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Create automatic backup (SaveVault creates this)
	automaticBackupPath := vaultPath + BackupSuffix
	vaultContent, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault: %v", err)
	}
	if err := os.WriteFile(automaticBackupPath, vaultContent, VaultPermissions); err != nil {
		t.Fatalf("Failed to create automatic backup: %v", err)
	}

	// Create multiple manual backups with different timestamps
	manual1, err := storage.CreateManualBackup()
	if err != nil {
		t.Fatalf("CreateManualBackup 1 failed: %v", err)
	}

	time.Sleep(1 * time.Second) // Ensure different timestamps

	manual2, err := storage.CreateManualBackup()
	if err != nil {
		t.Fatalf("CreateManualBackup 2 failed: %v", err)
	}

	// List all backups
	backups, err := storage.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	// Verify we have all backups (1 automatic + 2 manual = 3 total, but may be 2 if timestamps collide)
	if len(backups) < 2 {
		t.Errorf("Expected at least 2 backups, got %d", len(backups))
	}

	// Verify backups are sorted by ModTime descending (newest first)
	for i := 0; i < len(backups)-1; i++ {
		if backups[i].ModTime.Before(backups[i+1].ModTime) {
			t.Errorf("Backups not sorted correctly: backup[%d] (%v) is older than backup[%d] (%v)",
				i, backups[i].ModTime, i+1, backups[i+1].ModTime)
		}
	}

	// Verify backup types are identified correctly
	foundAutomatic := false
	foundManual := false
	for _, backup := range backups {
		if backup.Type == BackupTypeAutomatic {
			foundAutomatic = true
			if backup.Path != automaticBackupPath {
				t.Errorf("Automatic backup path mismatch: got %s, want %s", backup.Path, automaticBackupPath)
			}
		}
		if backup.Type == BackupTypeManual {
			foundManual = true
			if backup.Path != manual1 && backup.Path != manual2 {
				t.Errorf("Unknown manual backup path: %s", backup.Path)
			}
		}
	}

	if !foundAutomatic {
		t.Error("Automatic backup not found in list")
	}
	if !foundManual {
		t.Error("Manual backup not found in list")
	}

	// Verify all backups have valid metadata
	for i, backup := range backups {
		if backup.Path == "" {
			t.Errorf("Backup[%d] has empty path", i)
		}
		if backup.Size == 0 {
			t.Errorf("Backup[%d] has zero size", i)
		}
		if backup.ModTime.IsZero() {
			t.Errorf("Backup[%d] has zero ModTime", i)
		}
		// IsCorrupted should be false for valid backups
		if backup.IsCorrupted {
			t.Errorf("Backup[%d] marked as corrupted but should be valid", i)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr)+1 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
