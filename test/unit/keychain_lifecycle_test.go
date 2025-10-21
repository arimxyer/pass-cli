package unit_test

import (
	"os"
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
	defer func() { _ = ks.Delete() }()

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
	defer func() { _ = ks.Delete() }()

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
	defer func() { _ = ks.Delete() }()

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

// T016: Unit test for status command with keychain enabled
// Tests: displays availability, storage status, backend
func TestKeychainStatus_Enabled(t *testing.T) {
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

	if err := vs.Initialize(testPassword, true, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Cleanup keychain after test
	defer func() { _ = ks.Delete() }()

	// Test will FAIL until cmd/keychain_status.go is implemented
	t.Skip("TODO: Implement keychain status command (T021)")

	// TODO T021: After implementation, test should:
	// 1. Run status command logic
	// 2. Verify output contains "Available"
	// 3. Verify output contains "Password Stored: Yes"
	// 4. Verify output contains backend name (Windows Credential Manager / macOS Keychain / Linux Secret Service)
	// 5. Verify exit code 0
}

// T017: Unit test for status command with keychain available but not enabled
// Tests: actionable suggestion
func TestKeychainStatus_AvailableNotEnabled(t *testing.T) {
	// Skip if keychain not available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("Keychain not available on this platform")
	}

	// Setup: Create temp vault WITHOUT keychain
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

	// Cleanup keychain after test (in case it was set)
	defer func() { _ = ks.Delete() }()

	// Test will FAIL until cmd/keychain_status.go is implemented
	t.Skip("TODO: Implement keychain status command (T021)")

	// TODO T021: After implementation, test should:
	// 1. Run status command logic
	// 2. Verify output contains "Available"
	// 3. Verify output contains "Password Stored: No"
	// 4. Verify output contains actionable suggestion: "pass-cli keychain enable"
	// 5. Verify exit code 0
}

// T018: Unit test for status command with keychain unavailable
// Tests: platform-specific unavailable message
func TestKeychainStatus_Unavailable(t *testing.T) {
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

	// Test will FAIL until cmd/keychain_status.go is implemented
	t.Skip("TODO: Implement keychain status command (T021)")

	// TODO T021: After implementation, test should:
	// 1. Run status command logic
	// 2. Verify output contains platform-specific message:
	//    - Windows: "Windows Credential Manager"
	//    - macOS: "macOS Keychain"
	//    - Linux: "Linux Secret Service"
	// 3. Verify output indicates unavailability
	// 4. Verify exit code 0 (informational, not an error)
}

// T019: Unit test for status command always returns exit code 0
// Tests: informational nature (never fails)
func TestKeychainStatus_AlwaysExitZero(t *testing.T) {
	// Skip if keychain not available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("Keychain not available on this platform")
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

	// Cleanup keychain after test
	defer func() { _ = ks.Delete() }()

	// Test will FAIL until cmd/keychain_status.go is implemented
	t.Skip("TODO: Implement keychain status command (T021)")

	// TODO T021: After implementation, test should:
	// 1. Run status command logic in various scenarios (enabled, not enabled, unavailable)
	// 2. Verify all scenarios return nil error (exit code 0)
	// 3. Status command is informational only, never returns error
}

// T024: Unit test for remove command success - both deleted
// Tests: file + keychain deletion
func TestVaultRemove_BothDeleted(t *testing.T) {
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

	if err := vs.Initialize(testPassword, true, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Verify both vault file and keychain entry exist
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault file should exist")
	}
	if _, err := ks.Retrieve(); err != nil {
		t.Fatal("Keychain entry should exist")
	}

	// Test will FAIL until cmd/vault_remove.go is implemented
	t.Skip("TODO: Implement vault remove command (T030)")

	// TODO T030: After implementation, test should:
	// 1. Run remove command logic with confirmation
	// 2. Verify vault file deleted (os.Stat returns IsNotExist)
	// 3. Verify keychain entry deleted (ks.Retrieve() returns error)
	// 4. Verify exit code 0 (success)
}

// T025: Unit test for remove command - file missing, keychain exists
// Tests: FR-012 orphan cleanup
func TestVaultRemove_FileMissingKeychainExists(t *testing.T) {
	// Skip if keychain not available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("Keychain not available on this platform")
	}

	// Setup: Create temp vault WITH keychain, then manually delete file
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	testPassword := []byte("TestPassword@123")

	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	if err := vs.Initialize(testPassword, true, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Manually delete vault file (simulate orphaned keychain entry)
	if err := os.Remove(vaultPath); err != nil {
		t.Fatalf("Failed to delete vault file: %v", err)
	}

	// Verify keychain entry still exists
	if _, err := ks.Retrieve(); err != nil {
		t.Fatal("Keychain entry should still exist")
	}

	// Cleanup keychain after test
	defer func() { _ = ks.Delete() }()

	// Test will FAIL until cmd/vault_remove.go is implemented
	t.Skip("TODO: Implement vault remove command (T030)")

	// TODO T030: After implementation, test should:
	// 1. Run remove command logic
	// 2. Verify keychain entry deleted (FR-012: orphan cleanup)
	// 3. Verify no error (file missing is OK per FR-012)
	// 4. Verify warning message about missing file
}

// T026: Unit test for remove command - user cancels confirmation
// Tests: no deletion on cancel
func TestVaultRemove_UserCancels(t *testing.T) {
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

	if err := vs.Initialize(testPassword, true, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Cleanup keychain after test
	defer func() { _ = ks.Delete() }()

	// Test will FAIL until cmd/vault_remove.go is implemented
	t.Skip("TODO: Implement vault remove command (T030)")

	// TODO T030: After implementation, test should:
	// 1. Run remove command logic with user input "n" or "no"
	// 2. Verify vault file still exists
	// 3. Verify keychain entry still exists
	// 4. Verify exit code 1 (user cancelled)
}

// T027: Unit test for remove command with --yes flag
// Tests: skip prompt
func TestVaultRemove_WithYesFlag(t *testing.T) {
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

	if err := vs.Initialize(testPassword, true, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Test will FAIL until cmd/vault_remove.go is implemented
	t.Skip("TODO: Implement vault remove command (T030)")

	// TODO T030: After implementation, test should:
	// 1. Run remove command logic with --yes flag (no prompt)
	// 2. Verify vault file deleted
	// 3. Verify keychain entry deleted
	// 4. Verify no confirmation prompt shown
}

// T028: Unit test for remove command - audit log BEFORE deletion
// Tests: FR-015 logging order
func TestVaultRemove_AuditLogBeforeDeletion(t *testing.T) {
	// Skip if keychain not available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("Keychain not available on this platform")
	}

	// Setup: Create temp vault WITH keychain AND audit enabled
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	auditLogPath := filepath.Join(tempDir, "audit.log")
	testPassword := []byte("TestPassword@123")

	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	if err := vs.Initialize(testPassword, true, auditLogPath, vaultPath); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Cleanup keychain after test
	defer func() { _ = ks.Delete() }()

	// Test will FAIL until cmd/vault_remove.go is implemented
	t.Skip("TODO: Implement vault remove command (T030)")

	// TODO T030: After implementation, test should:
	// 1. Run remove command logic
	// 2. Verify audit log contains vault_remove event (logged BEFORE file deletion)
	// 3. This prevents losing audit trail if deleting audit log itself
}
