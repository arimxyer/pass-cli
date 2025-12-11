package keychain

import (
	"testing"

	"github.com/zalando/go-keyring"
)

// Test-specific constants to avoid conflicts with real keychain entries
const (
	testServiceName = "pass-cli-test"
	testAccountName = "test-master-password"
)

// testKeychainService wraps KeychainService for testing with isolated keychain entries
type testKeychainService struct {
	*KeychainService
}

func newTestKeychainService() *testKeychainService {
	return &testKeychainService{KeychainService: New("")}
}

func (tks *testKeychainService) Store(password string) error {
	return keyring.Set(testServiceName, testAccountName, password)
}

func (tks *testKeychainService) Retrieve() (string, error) {
	password, err := keyring.Get(testServiceName, testAccountName)
	if err == keyring.ErrNotFound {
		return "", ErrPasswordNotFound
	}
	return password, err
}

func (tks *testKeychainService) Delete() error {
	err := keyring.Delete(testServiceName, testAccountName)
	if err == keyring.ErrNotFound {
		return nil
	}
	return err
}

func TestNew(t *testing.T) {
	// Test with empty vaultID (legacy/global behavior)
	ks := New("")
	if ks == nil {
		t.Fatal("New(\"\") returned nil")
	}
	if ks.vaultID != "" {
		t.Errorf("vaultID = %q, want empty string", ks.vaultID)
	}

	// Test with vault ID
	ksVault := New("test-vault")
	if ksVault == nil {
		t.Fatal("New(\"test-vault\") returned nil")
	}
	if ksVault.vaultID != "test-vault" {
		t.Errorf("vaultID = %q, want %q", ksVault.vaultID, "test-vault")
	}

	// Availability depends on the test environment
	// Just verify the field is set (true or false)
	t.Logf("Keychain available: %v", ks.IsAvailable())
}

func TestStoreAndRetrieve(t *testing.T) {
	ks := newTestKeychainService()

	if !ks.IsAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	// Clean up before test - ensure we start with a clean slate
	_ = ks.Delete()

	testPassword := "test-master-password-12345"

	// Test Store
	err := ks.Store(testPassword)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Test Retrieve
	retrieved, err := ks.Retrieve()
	if err != nil {
		t.Fatalf("Retrieve() failed: %v", err)
	}

	if retrieved != testPassword {
		t.Errorf("Retrieved password = %q, want %q", retrieved, testPassword)
	}

	// Clean up after test
	if err := ks.Delete(); err != nil {
		t.Logf("Warning: cleanup delete failed: %v", err)
	}
}

func TestRetrieveNonExistent(t *testing.T) {
	ks := newTestKeychainService()

	if !ks.IsAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	// Ensure password doesn't exist
	_ = ks.Delete()

	// Try to retrieve non-existent password
	_, err := ks.Retrieve()
	if err == nil {
		t.Fatal("Retrieve() should fail for non-existent password")
	}

	if err != ErrPasswordNotFound {
		t.Errorf("Retrieve() error = %v, want %v", err, ErrPasswordNotFound)
	}
}

func TestDelete(t *testing.T) {
	ks := newTestKeychainService()

	if !ks.IsAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	// Clean up before test
	_ = ks.Delete()

	// Store a password first
	testPassword := "test-password-to-delete"
	err := ks.Store(testPassword)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Delete it
	err = ks.Delete()
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify it's gone
	_, err = ks.Retrieve()
	if err != ErrPasswordNotFound {
		t.Errorf("After Delete(), Retrieve() error = %v, want %v", err, ErrPasswordNotFound)
	}
}

func TestDeleteNonExistent(t *testing.T) {
	ks := newTestKeychainService()

	if !ks.IsAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	// Ensure password doesn't exist
	_ = ks.Delete()

	// Delete should not error for non-existent password
	err := ks.Delete()
	if err != nil {
		t.Errorf("Delete() on non-existent password failed: %v", err)
	}
}

func TestClear(t *testing.T) {
	ks := newTestKeychainService()

	if !ks.IsAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	// Clean up before test - ensure we start with a clean slate
	_ = ks.Delete()

	// Store a password
	testPassword := "test-password-to-clear"
	err := ks.Store(testPassword)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Clear it (using Delete since testKeychainService doesn't wrap Clear)
	err = ks.Delete()
	if err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	// Verify it's gone
	_, err = ks.Retrieve()
	if err != ErrPasswordNotFound {
		t.Errorf("After Clear(), Retrieve() error = %v, want %v", err, ErrPasswordNotFound)
	}
}

func TestUnavailableKeychain(t *testing.T) {
	// After removing proactive availability checks (for macOS CI fix),
	// operations now attempt to access keychain directly regardless of 'available' flag.
	// The 'available' flag is now only set by Ping() and is not checked before operations.
	// This test verifies operations complete (successfully or with error) without panicking.

	ks := &KeychainService{available: false}

	// Test Store - may succeed or fail depending on actual system keychain availability
	err := ks.Store("test-password-unavailable-check")
	t.Logf("Store() returned: %v", err)

	// Test Retrieve - may succeed (if Store succeeded) or fail
	_, err = ks.Retrieve()
	t.Logf("Retrieve() returned: %v", err)

	// Test Delete - should complete without panic
	err = ks.Delete()
	t.Logf("Delete() returned: %v", err)

	// Test Clear - should behave same as Delete
	err = ks.Clear()
	t.Logf("Clear() returned: %v", err)

	// Success if we get here without panicking
	t.Log("✓ All operations completed without panic (expected behavior after lazy initialization changes)")
}

func TestStoreEmptyPassword(t *testing.T) {
	ks := newTestKeychainService()

	if !ks.IsAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	// Clean up before test
	_ = ks.Delete()

	// Store empty password (should be allowed)
	err := ks.Store("")
	if err != nil {
		t.Fatalf("Store() with empty password failed: %v", err)
	}

	// Retrieve it
	retrieved, err := ks.Retrieve()
	if err != nil {
		t.Fatalf("Retrieve() failed: %v", err)
	}

	if retrieved != "" {
		t.Errorf("Retrieved password = %q, want empty string", retrieved)
	}

	// Clean up
	_ = ks.Delete()
}

func TestMultipleStoreOverwrites(t *testing.T) {
	ks := newTestKeychainService()

	if !ks.IsAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	// Clean up before test
	_ = ks.Delete()

	// Store first password
	password1 := "first-password"
	err := ks.Store(password1)
	if err != nil {
		t.Fatalf("First Store() failed: %v", err)
	}

	// Store second password (should overwrite)
	password2 := "second-password"
	err = ks.Store(password2)
	if err != nil {
		t.Fatalf("Second Store() failed: %v", err)
	}

	// Retrieve should get the second password
	retrieved, err := ks.Retrieve()
	if err != nil {
		t.Fatalf("Retrieve() failed: %v", err)
	}

	if retrieved != password2 {
		t.Errorf("Retrieved password = %q, want %q", retrieved, password2)
	}

	// Clean up
	_ = ks.Delete()
}

// TestCheckAvailability verifies the lazy initialization behavior
func TestCheckAvailability(t *testing.T) {
	ks := New("")

	// IsAvailable() now checks availability on demand by calling Ping()
	// So it should return the actual availability status
	available := ks.IsAvailable()

	// Verify consistent behavior by calling IsAvailable() again
	available2 := ks.IsAvailable()
	if available != available2 {
		t.Error("IsAvailable() should return consistent results")
	}

	if available {
		t.Log("✓ Keychain available on this system")
	} else {
		t.Log("✓ Keychain unavailable on this system")
	}

	// Ping() should return consistent results with IsAvailable()
	err := ks.Ping()
	if err == nil {
		if !ks.IsAvailable() {
			t.Error("After successful Ping(), IsAvailable() should return true")
		}
	} else {
		if ks.IsAvailable() {
			t.Error("After failed Ping(), IsAvailable() should return false")
		}
	}
}

// TestSanitizeVaultID tests the vault ID sanitization function
func TestSanitizeVaultID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{".", ""},
		{"my-vault", "my-vault"},
		{"my_vault", "my_vault"},
		{"MyVault123", "MyVault123"},
		{"my vault", "my_vault"},
		{"my/vault", "my_vault"},
		{"my\\vault", "my_vault"},
		{"my:vault", "my_vault"},
		{"vault-with-dashes", "vault-with-dashes"},
		{"vault_with_underscores", "vault_with_underscores"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := sanitizeVaultID(tc.input)
			if result != tc.expected {
				t.Errorf("sanitizeVaultID(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestAccountName tests the account name generation
func TestAccountName(t *testing.T) {
	tests := []struct {
		vaultID  string
		expected string
	}{
		{"", "master-password"},
		{"my-vault", "master-password-my-vault"},
		{"test_vault", "master-password-test_vault"},
	}

	for _, tc := range tests {
		t.Run(tc.vaultID, func(t *testing.T) {
			ks := New(tc.vaultID)
			result := ks.accountName()
			if result != tc.expected {
				t.Errorf("accountName() = %q, want %q", result, tc.expected)
			}
		})
	}
}

// TestVaultIsolation verifies that different vaults have isolated keychain entries
func TestVaultIsolation(t *testing.T) {
	ks1 := New("vault1")
	ks2 := New("vault2")

	if !ks1.IsAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	// Clean up before test
	_ = ks1.Delete()
	_ = ks2.Delete()

	// Store different passwords in different vaults
	password1 := "password-for-vault1"
	password2 := "password-for-vault2"

	if err := ks1.Store(password1); err != nil {
		t.Fatalf("Failed to store password1: %v", err)
	}
	if err := ks2.Store(password2); err != nil {
		t.Fatalf("Failed to store password2: %v", err)
	}

	// Retrieve and verify isolation
	retrieved1, err := ks1.Retrieve()
	if err != nil {
		t.Fatalf("Failed to retrieve password1: %v", err)
	}
	if retrieved1 != password1 {
		t.Errorf("Vault1 password = %q, want %q", retrieved1, password1)
	}

	retrieved2, err := ks2.Retrieve()
	if err != nil {
		t.Fatalf("Failed to retrieve password2: %v", err)
	}
	if retrieved2 != password2 {
		t.Errorf("Vault2 password = %q, want %q", retrieved2, password2)
	}

	// Clean up
	_ = ks1.Delete()
	_ = ks2.Delete()
}

// TestMigrationFromGlobal tests the migration from global to vault-specific entry
func TestMigrationFromGlobal(t *testing.T) {
	ksGlobal := New("")
	ksVault := New("test-migration-vault")

	if !ksGlobal.IsAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	// Clean up before test
	_ = ksGlobal.Delete()
	_ = ksVault.Delete()

	// Store password in global entry
	globalPassword := "global-password-to-migrate"
	if err := ksGlobal.Store(globalPassword); err != nil {
		t.Fatalf("Failed to store global password: %v", err)
	}

	// Verify global entry exists
	if !ksVault.HasGlobalEntry() {
		t.Fatal("HasGlobalEntry() should return true after storing global password")
	}

	// Migrate to vault-specific entry
	migrated, err := ksVault.MigrateFromGlobal()
	if err != nil {
		t.Fatalf("MigrateFromGlobal() failed: %v", err)
	}
	if !migrated {
		t.Fatal("MigrateFromGlobal() should return true when global entry exists")
	}

	// Verify vault-specific entry has the password
	retrieved, err := ksVault.Retrieve()
	if err != nil {
		t.Fatalf("Failed to retrieve migrated password: %v", err)
	}
	if retrieved != globalPassword {
		t.Errorf("Migrated password = %q, want %q", retrieved, globalPassword)
	}

	// Global entry should still exist (not deleted by MigrateFromGlobal)
	if !ksVault.HasGlobalEntry() {
		t.Error("HasGlobalEntry() should still return true after migration")
	}

	// Delete global entry
	if err := ksVault.DeleteGlobal(); err != nil {
		t.Fatalf("DeleteGlobal() failed: %v", err)
	}

	// Verify global entry is gone
	if ksVault.HasGlobalEntry() {
		t.Error("HasGlobalEntry() should return false after DeleteGlobal()")
	}

	// Vault-specific entry should still exist
	retrieved, err = ksVault.Retrieve()
	if err != nil {
		t.Fatalf("Vault-specific entry should still exist: %v", err)
	}
	if retrieved != globalPassword {
		t.Errorf("Vault-specific password = %q, want %q", retrieved, globalPassword)
	}

	// Clean up
	_ = ksVault.Delete()
}

// TestMigrationNoGlobalEntry tests migration when no global entry exists
func TestMigrationNoGlobalEntry(t *testing.T) {
	ksGlobal := New("")
	ksVault := New("test-no-global-vault")

	if !ksGlobal.IsAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	// Clean up - ensure no global entry exists
	_ = ksGlobal.Delete()
	_ = ksVault.Delete()

	// Attempt migration when no global entry exists
	migrated, err := ksVault.MigrateFromGlobal()
	if err != nil {
		t.Fatalf("MigrateFromGlobal() failed: %v", err)
	}
	if migrated {
		t.Error("MigrateFromGlobal() should return false when no global entry exists")
	}
}

// TestMigrationWithEmptyVaultID tests that migration is skipped for global service
func TestMigrationWithEmptyVaultID(t *testing.T) {
	ksGlobal := New("")

	// Migration should be skipped for global service
	migrated, err := ksGlobal.MigrateFromGlobal()
	if err != nil {
		t.Fatalf("MigrateFromGlobal() failed: %v", err)
	}
	if migrated {
		t.Error("MigrateFromGlobal() should return false for global service")
	}
}
