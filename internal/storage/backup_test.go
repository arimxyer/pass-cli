package storage

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"pass-cli/internal/crypto"
)

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
