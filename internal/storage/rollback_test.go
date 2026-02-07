package storage

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arimxyer/pass-cli/internal/crypto"
)

// TestSaveVault_RollbackOnSecondRenameFail tests the critical rollback path
// when the second atomic rename (temp→vault) fails
func TestSaveVault_RollbackOnSecondRenameFail(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	// Create spy filesystem (delegates to real OS but can fail on demand)
	spyFS := NewSpyFileSystem()

	s, err := NewStorageServiceWithFS(cryptoService, vaultPath, spyFS)
	if err != nil {
		t.Fatalf("NewStorageServiceWithFS failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault (this creates the initial vault file)
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Save some initial data
	initialData := []byte(`{"credentials":[{"service":"initial","username":"user1"}]}`)
	if err := s.SaveVault(initialData, password, nil); err != nil {
		t.Fatalf("First SaveVault failed: %v", err)
	}

	// Verify initial data was saved
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault after initial save failed: %v", err)
	}
	if !bytes.Equal(loadedData, initialData) {
		t.Fatal("Initial data mismatch")
	}

	// Now configure spy to fail on the SECOND rename
	// SaveVault performs two renames:
	// 1. vault.enc → vault.enc.backup (should succeed)
	// 2. vault.enc.tmp.XXX → vault.enc (should fail)
	spyFS.ResetRenameCount()
	spyFS.failRenameAt = 2 // Fail on second rename call

	// Try to save new data - second rename should fail, triggering rollback
	newData := []byte(`{"credentials":[{"service":"new","username":"user2"}]}`)
	err = s.SaveVault(newData, password, nil)

	// SaveVault should return an error indicating the critical failure
	if err == nil {
		t.Fatal("SaveVault should have failed when second rename fails")
	}

	if !strings.Contains(err.Error(), "CRITICAL") {
		t.Errorf("Error should indicate CRITICAL failure, got: %v", err)
	}

	// CRITICAL VERIFICATION: Vault should be restored to original state
	// The rollback logic should have attempted: backup → vault
	restoredData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault after rollback failed: %v", err)
	}

	if !bytes.Equal(restoredData, initialData) {
		t.Errorf("Rollback failed: vault should contain original data")
		t.Logf("Expected: %s", initialData)
		t.Logf("Got: %s", restoredData)
	}

	// Verify vault is still functional after rollback
	anotherData := []byte(`{"credentials":[{"service":"test","username":"test"}]}`)
	spyFS.failRenameAt = 0 // Reset failure

	if err := s.SaveVault(anotherData, password, nil); err != nil {
		t.Fatalf("SaveVault should work after rollback: %v", err)
	}

	finalData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault after recovery failed: %v", err)
	}

	if !bytes.Equal(finalData, anotherData) {
		t.Error("Vault should be functional after rollback")
	}
}

// TestAtomicRename_FirstRenameFail tests when first rename (vault→backup) fails
func TestAtomicRename_FirstRenameFail(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	// Create spy filesystem
	spyFS := NewSpyFileSystem()

	s, err := NewStorageServiceWithFS(cryptoService, vaultPath, spyFS)
	if err != nil {
		t.Fatalf("NewStorageServiceWithFS failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Save initial data
	initialData := []byte(`{"credentials":[{"service":"test"}]}`)
	if err := s.SaveVault(initialData, password, nil); err != nil {
		t.Fatalf("Initial SaveVault failed: %v", err)
	}

	// Configure mock to fail on FIRST rename
	spyFS.ResetRenameCount()
	spyFS.failRenameAt = 1

	// Try to save - should fail on first rename, vault should remain unchanged
	newData := []byte(`{"credentials":[{"service":"new"}]}`)
	err = s.SaveVault(newData, password, nil)

	if err == nil {
		t.Fatal("SaveVault should fail when first rename fails")
	}

	// Vault should NOT be modified (error before any atomic operation completed)
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if !bytes.Equal(loadedData, initialData) {
		t.Error("Vault should be unchanged when first rename fails")
	}
}

// TestAtomicRename_PermissionError tests permission denied during rename
func TestAtomicRename_PermissionError(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	spyFS := NewSpyFileSystem()

	s, err := NewStorageServiceWithFS(cryptoService, vaultPath, spyFS)
	if err != nil {
		t.Fatalf("NewStorageServiceWithFS failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault (must succeed before we enable failures)
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Now set rename to always fail with permission error
	spyFS.failRename = true

	// Try to save data - should fail with permission error
	data := []byte(`{"credentials":[]}`)
	err = s.SaveVault(data, password, nil)

	if err == nil {
		t.Fatal("SaveVault should fail with permission error")
	}

	// Error should be related to filesystem or permission
	if !strings.Contains(err.Error(), "save failed") {
		t.Errorf("Expected save failed error, got: %v", err)
	}
}

// TestSaveVault_CallbackDuringRollback verifies callbacks are invoked during rollback
func TestSaveVault_CallbackDuringRollback(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	spyFS := NewSpyFileSystem()

	s, err := NewStorageServiceWithFS(cryptoService, vaultPath, spyFS)
	if err != nil {
		t.Fatalf("NewStorageServiceWithFS failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Save initial data
	initialData := []byte(`{"credentials":[{"service":"test"}]}`)
	if err := s.SaveVault(initialData, password, nil); err != nil {
		t.Fatalf("Initial SaveVault failed: %v", err)
	}

	// Configure mock to fail on second rename
	spyFS.ResetRenameCount()
	spyFS.failRenameAt = 2

	// Track callback events
	var events []string
	callback := func(event string, metadata ...string) {
		events = append(events, event)
	}

	// Try to save - should fail and trigger rollback
	newData := []byte(`{"credentials":[{"service":"new"}]}`)
	_ = s.SaveVault(newData, password, callback)

	// Verify rollback callbacks were invoked
	hasRollbackStarted := false
	hasRollbackCompleted := false

	for _, event := range events {
		if event == "rollback_started" {
			hasRollbackStarted = true
		}
		if event == "rollback_completed" {
			hasRollbackCompleted = true
		}
	}

	if !hasRollbackStarted {
		t.Error("Expected rollback_started callback")
		t.Logf("Events: %v", events)
	}

	if !hasRollbackCompleted {
		t.Error("Expected rollback_completed callback")
		t.Logf("Events: %v", events)
	}

	// Verify "atomic_save_completed" is NOT called (save failed)
	for _, event := range events {
		if event == "atomic_save_completed" {
			t.Error("atomic_save_completed should not be called on failure")
		}
	}
}

// TestAtomicRename_NonexistentSource tests renaming a file that doesn't exist
func TestAtomicRename_NonexistentSource(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	spyFS := NewSpyFileSystem()

	s, err := NewStorageServiceWithFS(cryptoService, vaultPath, spyFS)
	if err != nil {
		t.Fatalf("NewStorageServiceWithFS failed: %v", err)
	}

	// Try to rename a file that doesn't exist
	err = s.atomicRename("/nonexistent/path", "/some/dest")

	if err == nil {
		t.Fatal("atomicRename should fail for nonexistent source")
	}

	// Should be a filesystem error
	if !strings.Contains(err.Error(), "filesystem not atomic") && !strings.Contains(err.Error(), "not exist") {
		t.Logf("Got error: %v", err)
	}
}

// Note: Removed TestSaveVault_VerificationFailTriggersCleanup
// SaveVault allows password changes (that's how ChangePassword works)
// The verification step decrypts with the SAME password used for encryption,
// so using a different password is not a verification failure
