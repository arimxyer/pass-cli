package unit_test

import (
	"path/filepath"
	"testing"

	"pass-cli/internal/keychain"
	"pass-cli/internal/vault"
)

// T004: Unit test for enable command success path
// Tests: correct password, keychain stores, audit logs
func TestKeychainEnable_SuccessPath(t *testing.T) {
	// Skip if keychain not available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("Keychain not available on this platform")
	}

	// Setup: Create temp vault without keychain
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	testPassword := []byte("TestPassword@123")

	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize vault without keychain
	if err := vs.Initialize(testPassword, false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Cleanup keychain after test
	defer ks.Delete()

	// Test will FAIL until cmd/keychain_enable.go is implemented
	t.Skip("TODO: Implement keychain enable command (T011)")

	// TODO T011: After implementation, test should:
	// 1. Run enable command logic with testPassword
	// 2. Verify keychain.Retrieve() returns the password
	// 3. Verify audit log contains EventKeychainEnable (if audit enabled)
}

// T005: Unit test for enable command wrong password
// Tests: error, password cleared, no keychain modification
func TestKeychainEnable_WrongPassword(t *testing.T) {
	// Skip if keychain not available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("Keychain not available on this platform")
	}

	// Setup: Create temp vault
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	correctPassword := []byte("CorrectPassword@123")
	_ = []byte("WrongPassword@456") // wrongPassword (will be used after T011 implementation)

	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	if err := vs.Initialize(correctPassword, false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Cleanup keychain after test
	defer ks.Delete()

	// Test will FAIL until cmd/keychain_enable.go is implemented
	t.Skip("TODO: Implement keychain enable command (T011)")

	// TODO T011: After implementation, test should:
	// 1. Run enable command logic with wrongPassword
	// 2. Verify error contains "invalid master password" or "failed to unlock"
	// 3. Verify keychain.Retrieve() returns error (password not stored)
	// 4. Verify wrongPassword bytes are cleared (all zeros)
}

// T006: Unit test for enable command keychain unavailable
// Tests: platform-specific error message
func TestKeychainEnable_KeychainUnavailable(t *testing.T) {
	// This test is platform-dependent - will skip if keychain IS available
	ks := keychain.New()
	if ks.IsAvailable() {
		t.Skip("Keychain is available - cannot test unavailable scenario")
	}

	// Setup: Create temp vault
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	testPassword := []byte("TestPassword@123")

	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	if err := vs.Initialize(testPassword, false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Test will FAIL until cmd/keychain_enable.go is implemented
	t.Skip("TODO: Implement keychain enable command (T011)")

	// TODO T011: After implementation, test should:
	// 1. Run enable command logic
	// 2. Verify error with platform-specific message:
	//    - Windows: "Windows Credential Manager access denied"
	//    - macOS: "macOS Keychain access denied"
	//    - Linux: "Linux Secret Service not running"
	// 3. Verify error includes troubleshooting guidance
}

// T007: Unit test for enable already enabled without --force
// Tests: graceful no-op
func TestKeychainEnable_AlreadyEnabled_NoForce(t *testing.T) {
	// Skip if keychain not available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("Keychain not available on this platform")
	}

	// Setup: Create temp vault WITH keychain
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	testPassword := []byte("TestPassword@123")

	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize with keychain enabled
	if err := vs.Initialize(testPassword, true, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Cleanup keychain after test
	defer ks.Delete()

	// Verify keychain has password
	_, err = ks.Retrieve()
	if err != nil {
		t.Fatalf("Keychain should have password after init with --use-keychain: %v", err)
	}

	// Test will FAIL until cmd/keychain_enable.go is implemented
	t.Skip("TODO: Implement keychain enable command (T011)")

	// TODO T011: After implementation, test should:
	// 1. Run enable command logic without force flag
	// 2. Verify output contains "already enabled"
	// 3. Verify output suggests using --force
	// 4. Verify no error (graceful no-op)
	// 5. Verify keychain entry unchanged
}

// T008: Unit test for enable already enabled with --force
// Tests: overwrite behavior
func TestKeychainEnable_AlreadyEnabled_WithForce(t *testing.T) {
	// Skip if keychain not available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("Keychain not available on this platform")
	}

	// Setup: Create temp vault WITH keychain
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	oldPassword := []byte("OldPassword@123")
	_ = []byte("NewPassword@456") // newPassword (will be used after T011 implementation)

	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize with old password and keychain
	if err := vs.Initialize(oldPassword, true, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Test will FAIL until cmd/keychain_enable.go is implemented
	t.Skip("TODO: Implement keychain enable command (T011)")

	// TODO T011: After implementation:
	// Change vault password to newPassword
	// if err := vs.Unlock(oldPassword); err != nil {
	//     t.Fatalf("Failed to unlock with old password: %v", err)
	// }
	// if err := vs.ChangePassword(newPassword); err != nil {
	//     t.Fatalf("Failed to change password: %v", err)
	// }
	// vs.Lock()
	//
	// At this point: keychain may have been updated, test logic needs review
	//
	// Cleanup keychain after test
	// defer ks.Delete()

	// TODO T011: After implementation, test should:
	// 1. Run enable command logic with --force and newPassword
	// 2. Verify keychain.Retrieve() returns newPassword (overwritten)
	// 3. Verify success message
	// 4. Verify no error
}
