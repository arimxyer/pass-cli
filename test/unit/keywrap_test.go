package unit

import (
	"bytes"
	"testing"

	"pass-cli/internal/crypto"
)

// T003: Unit test for GenerateDEK()
func TestGenerateDEK(t *testing.T) {
	t.Run("generates 32-byte key", func(t *testing.T) {
		dek, err := crypto.GenerateDEK()
		if err != nil {
			t.Fatalf("GenerateDEK() error = %v", err)
		}
		defer crypto.ClearBytes(dek)

		if len(dek) != 32 {
			t.Errorf("GenerateDEK() length = %d, want 32", len(dek))
		}
	})

	t.Run("generates unique keys", func(t *testing.T) {
		dek1, err := crypto.GenerateDEK()
		if err != nil {
			t.Fatalf("GenerateDEK() error = %v", err)
		}
		defer crypto.ClearBytes(dek1)

		dek2, err := crypto.GenerateDEK()
		if err != nil {
			t.Fatalf("GenerateDEK() error = %v", err)
		}
		defer crypto.ClearBytes(dek2)

		if bytes.Equal(dek1, dek2) {
			t.Error("GenerateDEK() generated identical keys")
		}
	})
}

// T004: Unit test for WrapKey() round-trip
func TestWrapKeyRoundTrip(t *testing.T) {
	// Generate test keys
	dek, err := crypto.GenerateDEK()
	if err != nil {
		t.Fatalf("GenerateDEK() error = %v", err)
	}
	defer crypto.ClearBytes(dek)

	kek := make([]byte, 32)
	copy(kek, "test-kek-for-wrapping-12345678") // 32 bytes
	defer crypto.ClearBytes(kek)

	// Wrap the DEK
	wrapped, err := crypto.WrapKey(dek, kek)
	if err != nil {
		t.Fatalf("WrapKey() error = %v", err)
	}

	// Verify wrapped structure
	if len(wrapped.Ciphertext) != 48 { // 32 bytes + 16 byte tag
		t.Errorf("WrapKey() ciphertext length = %d, want 48", len(wrapped.Ciphertext))
	}
	if len(wrapped.Nonce) != 12 {
		t.Errorf("WrapKey() nonce length = %d, want 12", len(wrapped.Nonce))
	}

	// Unwrap and verify round-trip
	unwrapped, err := crypto.UnwrapKey(wrapped, kek)
	if err != nil {
		t.Fatalf("UnwrapKey() error = %v", err)
	}
	defer crypto.ClearBytes(unwrapped)

	if !bytes.Equal(dek, unwrapped) {
		t.Error("WrapKey/UnwrapKey round-trip failed: keys don't match")
	}
}

// T005: Unit test for UnwrapKey() with wrong KEK
func TestUnwrapKeyWithWrongKEK(t *testing.T) {
	// Generate and wrap DEK with correct KEK
	dek, _ := crypto.GenerateDEK()
	defer crypto.ClearBytes(dek)

	correctKEK := make([]byte, 32)
	copy(correctKEK, "correct-kek-for-test-12345678")
	defer crypto.ClearBytes(correctKEK)

	wrapped, err := crypto.WrapKey(dek, correctKEK)
	if err != nil {
		t.Fatalf("WrapKey() error = %v", err)
	}

	// Try to unwrap with wrong KEK
	wrongKEK := make([]byte, 32)
	copy(wrongKEK, "wrong---kek-for-test-12345678")
	defer crypto.ClearBytes(wrongKEK)

	_, err = crypto.UnwrapKey(wrapped, wrongKEK)
	if err == nil {
		t.Error("UnwrapKey() with wrong KEK should fail")
	}
	if err != crypto.ErrDecryptionFailed {
		t.Errorf("UnwrapKey() error = %v, want ErrDecryptionFailed", err)
	}
}

// T006: Unit test for nonce uniqueness
func TestNonceUniqueness(t *testing.T) {
	dek, _ := crypto.GenerateDEK()
	defer crypto.ClearBytes(dek)

	kek := make([]byte, 32)
	copy(kek, "test-kek-for-nonce-test-1234")
	defer crypto.ClearBytes(kek)

	// Wrap the same DEK multiple times
	nonces := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		wrapped, err := crypto.WrapKey(dek, kek)
		if err != nil {
			t.Fatalf("WrapKey() iteration %d error = %v", i, err)
		}
		nonces[i] = wrapped.Nonce
	}

	// Verify all nonces are unique
	for i := 0; i < len(nonces); i++ {
		for j := i + 1; j < len(nonces); j++ {
			if bytes.Equal(nonces[i], nonces[j]) {
				t.Errorf("Nonces at index %d and %d are identical", i, j)
			}
		}
	}
}

// T007: Unit test for GenerateAndWrapDEK()
func TestGenerateAndWrapDEK(t *testing.T) {
	passwordKEK := make([]byte, 32)
	copy(passwordKEK, "password-kek-for-test-12345678")
	defer crypto.ClearBytes(passwordKEK)

	recoveryKEK := make([]byte, 32)
	copy(recoveryKEK, "recovery-kek-for-test-12345678")
	defer crypto.ClearBytes(recoveryKEK)

	result, err := crypto.GenerateAndWrapDEK(passwordKEK, recoveryKEK)
	if err != nil {
		t.Fatalf("GenerateAndWrapDEK() error = %v", err)
	}
	defer crypto.ClearBytes(result.DEK)

	// Verify DEK
	if len(result.DEK) != 32 {
		t.Errorf("GenerateAndWrapDEK() DEK length = %d, want 32", len(result.DEK))
	}

	// Verify password-wrapped DEK
	if len(result.PasswordWrapped.Ciphertext) != 48 {
		t.Errorf("PasswordWrapped ciphertext length = %d, want 48", len(result.PasswordWrapped.Ciphertext))
	}
	if len(result.PasswordWrapped.Nonce) != 12 {
		t.Errorf("PasswordWrapped nonce length = %d, want 12", len(result.PasswordWrapped.Nonce))
	}

	// Verify recovery-wrapped DEK
	if len(result.RecoveryWrapped.Ciphertext) != 48 {
		t.Errorf("RecoveryWrapped ciphertext length = %d, want 48", len(result.RecoveryWrapped.Ciphertext))
	}
	if len(result.RecoveryWrapped.Nonce) != 12 {
		t.Errorf("RecoveryWrapped nonce length = %d, want 12", len(result.RecoveryWrapped.Nonce))
	}

	// Verify both wrapped versions use different nonces
	if bytes.Equal(result.PasswordWrapped.Nonce, result.RecoveryWrapped.Nonce) {
		t.Error("Password and recovery wrapped DEKs use same nonce")
	}

	// Verify both can unwrap to same DEK
	unwrappedPassword, err := crypto.UnwrapKey(result.PasswordWrapped, passwordKEK)
	if err != nil {
		t.Fatalf("UnwrapKey(PasswordWrapped) error = %v", err)
	}
	defer crypto.ClearBytes(unwrappedPassword)

	unwrappedRecovery, err := crypto.UnwrapKey(result.RecoveryWrapped, recoveryKEK)
	if err != nil {
		t.Fatalf("UnwrapKey(RecoveryWrapped) error = %v", err)
	}
	defer crypto.ClearBytes(unwrappedRecovery)

	if !bytes.Equal(result.DEK, unwrappedPassword) {
		t.Error("Password unwrapped DEK doesn't match original")
	}
	if !bytes.Equal(result.DEK, unwrappedRecovery) {
		t.Error("Recovery unwrapped DEK doesn't match original")
	}
}

// T007.1: Contract test - verify WrapKey preconditions (32-byte dek/kek)
func TestWrapKeyPreconditions(t *testing.T) {
	validDEK := make([]byte, 32)
	validKEK := make([]byte, 32)

	t.Run("rejects short DEK", func(t *testing.T) {
		shortDEK := make([]byte, 16)
		_, err := crypto.WrapKey(shortDEK, validKEK)
		if err == nil {
			t.Error("WrapKey() should reject DEK shorter than 32 bytes")
		}
		if err != crypto.ErrInvalidKeyLength {
			t.Errorf("WrapKey() error = %v, want ErrInvalidKeyLength", err)
		}
	})

	t.Run("rejects long DEK", func(t *testing.T) {
		longDEK := make([]byte, 64)
		_, err := crypto.WrapKey(longDEK, validKEK)
		if err == nil {
			t.Error("WrapKey() should reject DEK longer than 32 bytes")
		}
		if err != crypto.ErrInvalidKeyLength {
			t.Errorf("WrapKey() error = %v, want ErrInvalidKeyLength", err)
		}
	})

	t.Run("rejects short KEK", func(t *testing.T) {
		shortKEK := make([]byte, 16)
		_, err := crypto.WrapKey(validDEK, shortKEK)
		if err == nil {
			t.Error("WrapKey() should reject KEK shorter than 32 bytes")
		}
		if err != crypto.ErrInvalidKeyLength {
			t.Errorf("WrapKey() error = %v, want ErrInvalidKeyLength", err)
		}
	})

	t.Run("rejects long KEK", func(t *testing.T) {
		longKEK := make([]byte, 64)
		_, err := crypto.WrapKey(validDEK, longKEK)
		if err == nil {
			t.Error("WrapKey() should reject KEK longer than 32 bytes")
		}
		if err != crypto.ErrInvalidKeyLength {
			t.Errorf("WrapKey() error = %v, want ErrInvalidKeyLength", err)
		}
	})

	t.Run("accepts valid 32-byte keys", func(t *testing.T) {
		_, err := crypto.WrapKey(validDEK, validKEK)
		if err != nil {
			t.Errorf("WrapKey() with valid keys error = %v", err)
		}
	})
}

// T007.2: Contract test - verify UnwrapKey postconditions (32-byte output)
func TestUnwrapKeyPostconditions(t *testing.T) {
	dek, _ := crypto.GenerateDEK()
	defer crypto.ClearBytes(dek)

	kek := make([]byte, 32)
	copy(kek, "test-kek-postcondition-test-123")
	defer crypto.ClearBytes(kek)

	wrapped, _ := crypto.WrapKey(dek, kek)

	t.Run("returns exactly 32-byte DEK", func(t *testing.T) {
		unwrapped, err := crypto.UnwrapKey(wrapped, kek)
		if err != nil {
			t.Fatalf("UnwrapKey() error = %v", err)
		}
		defer crypto.ClearBytes(unwrapped)

		if len(unwrapped) != 32 {
			t.Errorf("UnwrapKey() output length = %d, want 32", len(unwrapped))
		}
	})

	t.Run("rejects invalid ciphertext length", func(t *testing.T) {
		invalidWrapped := crypto.WrappedKey{
			Ciphertext: make([]byte, 20), // Too short (should be 48)
			Nonce:      make([]byte, 12),
		}
		_, err := crypto.UnwrapKey(invalidWrapped, kek)
		if err == nil {
			t.Error("UnwrapKey() should reject invalid ciphertext length")
		}
	})

	t.Run("rejects invalid nonce length", func(t *testing.T) {
		invalidWrapped := crypto.WrappedKey{
			Ciphertext: wrapped.Ciphertext,
			Nonce:      make([]byte, 8), // Too short (should be 12)
		}
		_, err := crypto.UnwrapKey(invalidWrapped, kek)
		if err == nil {
			t.Error("UnwrapKey() should reject invalid nonce length")
		}
	})
}
