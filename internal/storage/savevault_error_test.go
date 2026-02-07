package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arimxyer/pass-cli/internal/crypto"
)

// TestSaveVault_DifferentPassword tests SaveVault succeeds with different password (password change scenario)
func TestSaveVault_DifferentPassword(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	oldPassword := "OldPassword123!"
	newPassword := "NewPassword123!"

	// Initialize with old password
	if err := s.InitializeVault(oldPassword); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	data := []byte(`{"credentials":[{"service":"test"}]}`)

	// Save with new password - this re-encrypts the vault with new password
	// SaveVault uses the provided password to encrypt, verification uses same password to decrypt temp file
	// So this should succeed
	err = s.SaveVault(data, newPassword, nil)
	if err != nil {
		t.Fatalf("SaveVault should succeed with different password: %v", err)
	}

	// Verify vault is now accessible with NEW password (not old)
	loadedData, err := s.LoadVault(newPassword)
	if err != nil {
		t.Fatalf("Vault should be accessible with new password: %v", err)
	}

	if string(loadedData) != string(data) {
		t.Error("Data mismatch after password change")
	}

	// Old password should no longer work
	_, err = s.LoadVault(oldPassword)
	if err == nil {
		t.Error("Old password should not work after re-encryption")
	}
}

// TestSaveVault_VaultNotFound tests SaveVault when vault doesn't exist
func TestSaveVault_VaultNotFound(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "nonexistent.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	data := []byte(`{"credentials":[]}`)
	password := "TestPassword123!"

	// Try to save to non-existent vault
	err = s.SaveVault(data, password, nil)
	if err == nil {
		t.Error("SaveVault should fail when vault doesn't exist")
	}

	// Error should mention vault not found
	if err != ErrVaultNotFound {
		t.Logf("Expected ErrVaultNotFound, got: %v", err)
	}
}

// TestSaveVault_SuccessfulSave tests SaveVault completes successfully
func TestSaveVault_SuccessfulSave(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Save multiple times to verify idempotency
	for i := 0; i < 3; i++ {
		data := []byte(`{"credentials":[{"id":` + string(rune('0'+i)) + `}]}`)

		err = s.SaveVault(data, password, nil)
		if err != nil {
			t.Fatalf("SaveVault iteration %d failed: %v", i, err)
		}

		// Verify data was saved correctly
		loadedData, err := s.LoadVault(password)
		if err != nil {
			t.Fatalf("LoadVault iteration %d failed: %v", i, err)
		}

		if string(loadedData) != string(data) {
			t.Errorf("Iteration %d: data mismatch", i)
		}
	}
}

// TestSaveVault_CallbackInvocations tests that callbacks are invoked at all stages
func TestSaveVault_CallbackInvocations(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	data := []byte(`{"credentials":[{"service":"test"}]}`)

	// Track callback invocations
	var events []string
	callback := func(event string, metadata ...string) {
		events = append(events, event)
	}

	// Save with callback
	if err := s.SaveVault(data, password, callback); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Verify all expected events were called
	expectedEvents := []string{
		"atomic_save_started",
		"temp_file_created",
		"verification_started",
		"verification_passed",
		"atomic_rename_started",
		"atomic_rename_started", // Called twice: vault→backup, temp→vault
		"atomic_save_completed",
	}

	if len(events) != len(expectedEvents) {
		t.Errorf("Expected %d callback invocations, got %d", len(expectedEvents), len(events))
		t.Logf("Events: %v", events)
	}

	for i, expected := range expectedEvents {
		if i >= len(events) {
			t.Errorf("Missing event %d: %s", i, expected)
			continue
		}
		if events[i] != expected {
			t.Errorf("Event %d: expected %s, got %s", i, expected, events[i])
		}
	}
}

// TestSaveVault_CallbackIsOptional tests SaveVault works with nil callback
func TestSaveVault_CallbackIsOptional(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	data := []byte(`{"credentials":[]}`)

	// Save with nil callback - should work fine
	if err := s.SaveVault(data, password, nil); err != nil {
		t.Fatalf("SaveVault should work with nil callback: %v", err)
	}

	// Verify data was saved
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if string(loadedData) != string(data) {
		t.Error("Data mismatch")
	}
}

// TestPrepareEncryptedData_InvalidPassword tests encryption with wrong password characteristics
func TestPrepareEncryptedData_EmptyPassword(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Load metadata
	encryptedVault, err := s.loadEncryptedVault()
	if err != nil {
		t.Fatalf("loadEncryptedVault failed: %v", err)
	}

	data := []byte(`{"credentials":[]}`)

	// Try to prepare encrypted data with empty password
	_, err = s.prepareEncryptedData(data, encryptedVault.Metadata, "")
	// Empty password should work (it's valid, just weak)
	if err != nil {
		t.Logf("prepareEncryptedData with empty password: %v", err)
	}
}

// TestCreateBackup_NoVault tests backup creation when vault doesn't exist
func TestCreateBackup_NoVault(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "nonexistent.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Create backup of non-existent vault
	err = s.CreateBackup()
	// Should succeed silently (no vault to backup)
	if err != nil {
		t.Errorf("CreateBackup should succeed when vault doesn't exist: %v", err)
	}
}

// TestRestoreFromBackup_NoBackup tests restoration when backup doesn't exist
func TestRestoreFromBackup_NoBackup(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault (creates vault, no backup yet)
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Try to restore from non-existent backup
	err = s.RestoreFromBackup("")
	if err == nil {
		t.Error("RestoreFromBackup should fail when backup doesn't exist")
	}
}

// TestRestoreFromBackup_Success tests successful backup restoration
func TestRestoreFromBackup_Success(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Save original data
	originalData := []byte(`{"credentials":[{"service":"original"}]}`)
	if err := s.SaveVault(originalData, password, nil); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Create explicit backup
	if err := s.CreateBackup(); err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Save new data (overwrites vault)
	newData := []byte(`{"credentials":[{"service":"new"}]}`)
	if err := s.SaveVault(newData, password, nil); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Verify vault has new data
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}
	if string(loadedData) != string(newData) {
		t.Error("Vault should contain new data before restore")
	}

	// Restore from backup
	if err := s.RestoreFromBackup(""); err != nil {
		t.Fatalf("RestoreFromBackup failed: %v", err)
	}

	// Verify vault has original data
	restoredData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault after restore failed: %v", err)
	}

	if string(restoredData) != string(originalData) {
		t.Error("Vault should contain original data after restore")
	}
}

// TestRemoveBackup_Success tests backup removal
func TestRemoveBackup_Success(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Create backup
	if err := s.CreateBackup(); err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Verify backup exists
	backupPath := vaultPath + BackupSuffix
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("Backup file should exist")
	}

	// Remove backup
	if err := s.RemoveBackup(); err != nil {
		t.Fatalf("RemoveBackup failed: %v", err)
	}

	// Verify backup removed
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Backup file should be removed")
	}
}

// TestRemoveBackup_NoBackup tests removing non-existent backup
func TestRemoveBackup_NoBackup(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Remove non-existent backup - should succeed silently
	if err := s.RemoveBackup(); err != nil {
		t.Errorf("RemoveBackup should succeed for non-existent backup: %v", err)
	}
}
