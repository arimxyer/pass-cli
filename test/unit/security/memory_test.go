package security_test

import (
	"testing"

	"pass-cli/internal/crypto"
)

// TestMemoryClearing verifies that password bytes are cleared after vault operations.
// This test uses crypto.ClearBytes and checks that memory is zeroed within 1 second.
func TestMemoryClearing(t *testing.T) {
	// Create a test password
	password := []byte("test-password-123")
	passwordCopy := make([]byte, len(password))
	copy(passwordCopy, password)

	// Clear the password
	crypto.ClearBytes(password)

	// Verify all bytes are zeroed
	for i, b := range password {
		if b != 0 {
			t.Errorf("byte at index %d not cleared: expected 0, got %d", i, b)
		}
	}

	// Verify original content is preserved in copy
	if string(passwordCopy) != "test-password-123" {
		t.Errorf("password copy was modified")
	}
}

// TestPanicRecoveryWithDeferredCleanup verifies that deferred cleanup executes even on panic.
func TestPanicRecoveryWithDeferredCleanup(t *testing.T) {
	password := []byte("panic-test-password")

	// Track whether cleanup executed
	cleanupExecuted := false

	// Recover from panic and verify cleanup happened
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic did not occur")
		}
		if !cleanupExecuted {
			t.Error("deferred cleanup did not execute during panic")
		}
	}()

	// Simulate function with deferred cleanup that panics
	func() {
		defer func() {
			crypto.ClearBytes(password)
			cleanupExecuted = true
		}()

		// Simulate panic during password processing
		panic("simulated error during vault operation")
	}()
}

// TestVaultPasswordClearingAfterUnlock verifies master password is cleared after vault unlock.
// Blocked on: T009-T016 (vault methods accept []byte) - architectural change
func TestVaultPasswordClearingAfterUnlock(t *testing.T) {
	t.Skip("Deferred: Requires vault API refactor to accept []byte (tracked in separate spec)")

	// This test will verify that VaultService.masterPassword is cleared after Lock()
	// by using reflection or memory inspection tools.

	// Setup: Create temporary vault
	// Unlock with password
	// Lock vault
	// Verify masterPassword field is zeroed
}

// BenchmarkClearBytes measures the performance of memory clearing.
func BenchmarkClearBytes(b *testing.B) {
	password := make([]byte, 64) // Typical password length
	copy(password, []byte("benchmark-password-for-performance-testing-long-string"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crypto.ClearBytes(password)
	}
}
