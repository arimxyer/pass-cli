//go:build integration

package test

import (
	"os"
	"path/filepath"
	"testing"

	"pass-cli/internal/keychain"
	"pass-cli/internal/vault"
)

// TestKeychainPersistence_AfterRestart simulates the upgrade scenario:
// 1. User creates vault with keychain enabled
// 2. User updates pass-cli (new binary, same data)
// 3. User runs a command - should auto-unlock without password prompt
//
// This test catches regressions where:
// - Metadata file isn't persisted correctly
// - New VaultService instances can't read existing metadata
// - Keychain retrieval fails after "restart"
func TestKeychainPersistence_AfterRestart(t *testing.T) {
	// Skip if keychain is not available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	testPassword := "PersistenceTest-Pass@123"
	vaultDir := filepath.Join(testDir, "keychain-persistence-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	metadataPath := vaultPath + ".meta.json"

	// Clean up before and after
	defer cleanupKeychain(t, ks)
	defer cleanupVaultDir(t, vaultDir)
	_ = os.RemoveAll(vaultDir)
	_ = ks.Delete() // Clear any existing keychain entry

	// Create vault directory
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// ========================================
	// PHASE 1: Initial setup (like first install)
	// ========================================
	t.Log("Phase 1: Creating vault with keychain enabled...")

	vs1, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize vault
	if err := vs1.Initialize([]byte(testPassword), false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Enable keychain (stores password + sets metadata.KeychainEnabled = true)
	if err := vs1.EnableKeychain([]byte(testPassword), false); err != nil {
		t.Fatalf("Failed to enable keychain: %v", err)
	}

	// Verify metadata file was created with correct content
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Fatal("FAIL: Metadata file was not created after EnableKeychain()")
	}

	meta1, err := vault.LoadMetadata(vaultPath)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	if !meta1.KeychainEnabled {
		t.Fatal("FAIL: Metadata.KeychainEnabled should be true after EnableKeychain()")
	}
	t.Log("  - Metadata file created with KeychainEnabled=true")

	// Verify password is in keychain
	storedPassword, err := ks.Retrieve()
	if err != nil {
		t.Fatalf("FAIL: Password not stored in keychain: %v", err)
	}
	if storedPassword != testPassword {
		t.Fatal("FAIL: Stored password doesn't match original")
	}
	t.Log("  - Password stored in keychain")

	// ========================================
	// PHASE 2: Simulate app restart / binary update
	// ========================================
	t.Log("Phase 2: Simulating restart (new VaultService instance)...")

	// Create completely NEW VaultService instance
	// This simulates what happens after scoop update - new binary, same data files
	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create second vault service: %v", err)
	}

	// ========================================
	// PHASE 3: Verify auto-unlock works
	// ========================================
	t.Log("Phase 3: Verifying keychain auto-unlock...")

	// This is THE critical test - can UnlockWithKeychain() work after "restart"?
	err = vs2.UnlockWithKeychain()
	if err != nil {
		t.Fatalf("FAIL: UnlockWithKeychain() failed after restart: %v\n"+
			"This simulates the bug where users need to run 'keychain enable --force' after updates", err)
	}

	if !vs2.IsUnlocked() {
		t.Fatal("FAIL: Vault should be unlocked after UnlockWithKeychain()")
	}
	t.Log("  - UnlockWithKeychain() succeeded")

	// Verify we can actually access credentials (proves unlock worked)
	creds, err := vs2.ListCredentials()
	if err != nil {
		t.Fatalf("FAIL: ListCredentials() failed after keychain unlock: %v", err)
	}
	t.Logf("  - Successfully accessed vault (credential count: %d)", len(creds))

	t.Log("SUCCESS: Keychain persistence works correctly across restart")
}

// TestKeychainPersistence_MetadataIntegrity verifies that metadata file
// maintains integrity and correct values across multiple operations
func TestKeychainPersistence_MetadataIntegrity(t *testing.T) {
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	testPassword := "MetadataIntegrity-Pass@123"
	vaultDir := filepath.Join(testDir, "metadata-integrity-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	metadataPath := vaultPath + ".meta.json"

	defer cleanupKeychain(t, ks)
	defer cleanupVaultDir(t, vaultDir)
	_ = os.RemoveAll(vaultDir)
	_ = ks.Delete()

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Create and initialize vault with keychain
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	if err := vs.Initialize([]byte(testPassword), false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	if err := vs.EnableKeychain([]byte(testPassword), false); err != nil {
		t.Fatalf("Failed to enable keychain: %v", err)
	}

	// Read metadata file directly to verify actual disk content
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}

	t.Logf("Metadata file content:\n%s", string(metadataBytes))

	// Verify specific fields through multiple loads
	for i := 0; i < 3; i++ {
		meta, err := vault.LoadMetadata(vaultPath)
		if err != nil {
			t.Fatalf("Load %d: Failed to load metadata: %v", i+1, err)
		}

		if !meta.KeychainEnabled {
			t.Fatalf("Load %d: KeychainEnabled should be true, got false", i+1)
		}

		if meta.Version != "1.0" {
			t.Fatalf("Load %d: Version should be '1.0', got '%s'", i+1, meta.Version)
		}
	}

	t.Log("SUCCESS: Metadata maintains integrity across multiple reads")
}

// TestKeychainPersistence_GracefulDegradation verifies that when keychain
// is unavailable or metadata is missing, the system fails gracefully
func TestKeychainPersistence_GracefulDegradation(t *testing.T) {
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	testPassword := "Degradation-Pass@123"
	vaultDir := filepath.Join(testDir, "degradation-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	metadataPath := vaultPath + ".meta.json"

	defer cleanupKeychain(t, ks)
	defer cleanupVaultDir(t, vaultDir)
	_ = os.RemoveAll(vaultDir)
	_ = ks.Delete()

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Setup: Create vault with keychain enabled
	vs1, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	if err := vs1.Initialize([]byte(testPassword), false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	if err := vs1.EnableKeychain([]byte(testPassword), false); err != nil {
		t.Fatalf("Failed to enable keychain: %v", err)
	}

	// Test 1: Delete metadata file - should fail gracefully
	t.Log("Test 1: Simulating deleted metadata file...")
	if err := os.Remove(metadataPath); err != nil {
		t.Fatalf("Failed to delete metadata file: %v", err)
	}

	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	err = vs2.UnlockWithKeychain()
	if err == nil {
		t.Fatal("FAIL: UnlockWithKeychain() should fail when metadata is missing")
	}
	if err != vault.ErrKeychainNotEnabled {
		t.Logf("  - Got expected error type (details: %v)", err)
	} else {
		t.Log("  - Got ErrKeychainNotEnabled as expected")
	}

	// Restore metadata for next test
	meta := &vault.Metadata{
		Version:         "1.0",
		KeychainEnabled: true,
	}
	if err := vault.SaveMetadata(vaultPath, meta); err != nil {
		t.Fatalf("Failed to restore metadata: %v", err)
	}

	// Test 2: Delete keychain entry - should fail gracefully
	t.Log("Test 2: Simulating deleted keychain entry...")
	if err := ks.Delete(); err != nil {
		t.Fatalf("Failed to delete keychain entry: %v", err)
	}

	vs3, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	err = vs3.UnlockWithKeychain()
	if err == nil {
		t.Fatal("FAIL: UnlockWithKeychain() should fail when keychain entry is missing")
	}
	t.Logf("  - Got expected error: %v", err)

	// Test 3: Manual password unlock should still work
	t.Log("Test 3: Verifying manual password unlock still works...")
	if err := vs3.Unlock([]byte(testPassword)); err != nil {
		t.Fatalf("FAIL: Manual unlock should work even when keychain fails: %v", err)
	}
	t.Log("  - Manual password unlock succeeded")

	t.Log("SUCCESS: System degrades gracefully when keychain is unavailable")
}

// TestKeychainPersistence_MultipleRestarts simulates multiple app restarts
// to catch any state accumulation bugs
func TestKeychainPersistence_MultipleRestarts(t *testing.T) {
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	testPassword := "MultiRestart-Pass@123"
	vaultDir := filepath.Join(testDir, "multi-restart-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	defer cleanupKeychain(t, ks)
	defer cleanupVaultDir(t, vaultDir)
	_ = os.RemoveAll(vaultDir)
	_ = ks.Delete()

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Initial setup
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	if err := vs.Initialize([]byte(testPassword), false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	if err := vs.EnableKeychain([]byte(testPassword), false); err != nil {
		t.Fatalf("Failed to enable keychain: %v", err)
	}

	// Simulate 5 restarts (like 5 scoop updates)
	for i := 1; i <= 5; i++ {
		t.Logf("Restart %d: Creating new VaultService instance...", i)

		vsN, err := vault.New(vaultPath)
		if err != nil {
			t.Fatalf("Restart %d: Failed to create vault service: %v", i, err)
		}

		// Verify keychain unlock works
		if err := vsN.UnlockWithKeychain(); err != nil {
			t.Fatalf("Restart %d: UnlockWithKeychain() failed: %v\n"+
				"This indicates a regression in keychain persistence", i, err)
		}

		if !vsN.IsUnlocked() {
			t.Fatalf("Restart %d: Vault not unlocked after UnlockWithKeychain()", i)
		}

		// Access credentials to prove unlock actually worked
		_, err = vsN.ListCredentials()
		if err != nil {
			t.Fatalf("Restart %d: ListCredentials() failed: %v", i, err)
		}

		// Lock before next iteration
		vsN.Lock()
		t.Logf("  - Restart %d: SUCCESS", i)
	}

	t.Log("SUCCESS: Keychain persistence works across multiple restarts")
}
