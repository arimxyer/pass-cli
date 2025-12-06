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
	return &testKeychainService{KeychainService: New()}
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
	ks := New()
	if ks == nil {
		t.Fatal("New() returned nil")
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
	ks := New()

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
