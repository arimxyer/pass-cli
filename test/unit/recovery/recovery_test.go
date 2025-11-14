package recovery_test

import (
	"strings"
	"testing"

	"pass-cli/internal/recovery"
	"pass-cli/internal/vault"

	"github.com/tyler-smith/go-bip39"
)

// T015: Unit test for VerifyBackup()
// Tests: correct/incorrect words

func TestVerifyBackup(t *testing.T) {
	// Test mnemonic (valid BIP39 from test vectors)
	testMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"

	t.Run("succeeds with correct words", func(t *testing.T) {
		// Verify positions 0, 5, 10
		config := &recovery.VerifyConfig{
			Mnemonic:        testMnemonic,
			VerifyPositions: []int{0, 5, 10},
			UserWords:       []string{"abandon", "abandon", "abandon"},
		}

		err := recovery.VerifyBackup(config)
		if err != nil {
			t.Errorf("VerifyBackup() failed with correct words: %v", err)
		}
	})

	t.Run("fails with incorrect words", func(t *testing.T) {
		config := &recovery.VerifyConfig{
			Mnemonic:        testMnemonic,
			VerifyPositions: []int{0, 5, 23},
			UserWords:       []string{"abandon", "abandon", "wrong"}, // Position 23 should be "art"
		}

		err := recovery.VerifyBackup(config)
		if err == nil {
			t.Error("VerifyBackup() should fail with incorrect words")
		}
		if err != recovery.ErrVerificationFailed {
			t.Errorf("Expected ErrVerificationFailed, got: %v", err)
		}
	})

	t.Run("fails when word count mismatch", func(t *testing.T) {
		config := &recovery.VerifyConfig{
			Mnemonic:        testMnemonic,
			VerifyPositions: []int{0, 5, 10},
			UserWords:       []string{"abandon", "abandon"}, // Only 2 words, expected 3
		}

		err := recovery.VerifyBackup(config)
		if err == nil {
			t.Error("VerifyBackup() should fail when word count doesn't match positions")
		}
	})

	t.Run("case-insensitive comparison", func(t *testing.T) {
		config := &recovery.VerifyConfig{
			Mnemonic:        testMnemonic,
			VerifyPositions: []int{0, 5, 10},
			UserWords:       []string{"ABANDON", "Abandon", "aBaNdOn"}, // Mixed case
		}

		err := recovery.VerifyBackup(config)
		if err != nil {
			t.Errorf("VerifyBackup() should be case-insensitive, got error: %v", err)
		}
	})

	t.Run("trims whitespace from user words", func(t *testing.T) {
		config := &recovery.VerifyConfig{
			Mnemonic:        testMnemonic,
			VerifyPositions: []int{0, 5, 10},
			UserWords:       []string{" abandon ", "abandon\t", "\nabandon"},
		}

		err := recovery.VerifyBackup(config)
		if err != nil {
			t.Errorf("VerifyBackup() should trim whitespace, got error: %v", err)
		}
	})

	t.Run("fails with empty user words", func(t *testing.T) {
		config := &recovery.VerifyConfig{
			Mnemonic:        testMnemonic,
			VerifyPositions: []int{0, 5, 10},
			UserWords:       []string{"", "", ""},
		}

		err := recovery.VerifyBackup(config)
		if err == nil {
			t.Error("VerifyBackup() should fail with empty words")
		}
	})

	t.Run("validates positions are within range", func(t *testing.T) {
		config := &recovery.VerifyConfig{
			Mnemonic:        testMnemonic,
			VerifyPositions: []int{0, 24, 30}, // 24 and 30 are out of range
			UserWords:       []string{"abandon", "word", "word"},
		}

		err := recovery.VerifyBackup(config)
		if err == nil {
			t.Error("VerifyBackup() should fail with out-of-range positions")
		}
	})
}

// T016: Unit test for SetupRecovery()
// Tests: mnemonic generation, metadata creation, encryption

func TestSetupRecovery(t *testing.T) {
	t.Run("generates valid 24-word mnemonic", func(t *testing.T) {
		config := &recovery.SetupConfig{
			Passphrase: nil,
			KDFParams:  nil, // Use defaults
		}

		result, err := recovery.SetupRecovery(config)
		if err != nil {
			t.Fatalf("SetupRecovery() failed: %v", err)
		}

		// Verify mnemonic is 24 words
		words := strings.Fields(result.Mnemonic)
		if len(words) != 24 {
			t.Errorf("Expected 24 words in mnemonic, got %d", len(words))
		}

		// Verify BIP39 checksum
		_, err = bip39.EntropyFromMnemonic(result.Mnemonic)
		if err != nil {
			t.Errorf("Mnemonic has invalid checksum: %v", err)
		}
	})

	t.Run("creates valid recovery metadata", func(t *testing.T) {
		config := &recovery.SetupConfig{
			Passphrase: nil,
			KDFParams:  nil,
		}

		result, err := recovery.SetupRecovery(config)
		if err != nil {
			t.Fatalf("SetupRecovery() failed: %v", err)
		}

		metadata := result.Metadata
		if metadata == nil {
			t.Fatal("Metadata is nil")
		}

		// Verify enabled flag
		if !metadata.Enabled {
			t.Error("Metadata.Enabled should be true")
		}

		// Verify challenge positions (should be 6)
		if len(metadata.ChallengePositions) != 6 {
			t.Errorf("Expected 6 challenge positions, got %d", len(metadata.ChallengePositions))
		}

		// Verify positions are unique and within range
		seen := make(map[int]bool)
		for _, pos := range metadata.ChallengePositions {
			if pos < 0 || pos >= 24 {
				t.Errorf("Invalid position %d (must be 0-23)", pos)
			}
			if seen[pos] {
				t.Errorf("Duplicate position: %d", pos)
			}
			seen[pos] = true
		}

		// Verify encrypted data exists
		if len(metadata.EncryptedStoredWords) == 0 {
			t.Error("EncryptedStoredWords should not be empty")
		}
		if len(metadata.NonceStored) != 12 {
			t.Errorf("NonceStored should be 12 bytes, got %d", len(metadata.NonceStored))
		}
		if len(metadata.EncryptedRecoveryKey) == 0 {
			t.Error("EncryptedRecoveryKey should not be empty")
		}
		if len(metadata.NonceRecovery) != 12 {
			t.Errorf("NonceRecovery should be 12 bytes, got %d", len(metadata.NonceRecovery))
		}

		// Verify KDF params
		if metadata.KDFParams.Algorithm != "argon2id" {
			t.Errorf("Expected argon2id, got %s", metadata.KDFParams.Algorithm)
		}
		if len(metadata.KDFParams.SaltChallenge) != 32 {
			t.Errorf("SaltChallenge should be 32 bytes, got %d", len(metadata.KDFParams.SaltChallenge))
		}
		if len(metadata.KDFParams.SaltRecovery) != 32 {
			t.Errorf("SaltRecovery should be 32 bytes, got %d", len(metadata.KDFParams.SaltRecovery))
		}
	})

	t.Run("returns valid vault recovery key", func(t *testing.T) {
		config := &recovery.SetupConfig{
			Passphrase: nil,
			KDFParams:  nil,
		}

		result, err := recovery.SetupRecovery(config)
		if err != nil {
			t.Fatalf("SetupRecovery() failed: %v", err)
		}

		// Vault recovery key should be 32 bytes (AES-256)
		if len(result.VaultRecoveryKey) != 32 {
			t.Errorf("Expected 32-byte vault recovery key, got %d bytes", len(result.VaultRecoveryKey))
		}
	})

	t.Run("supports optional passphrase", func(t *testing.T) {
		config := &recovery.SetupConfig{
			Passphrase: []byte("my-secret-passphrase"),
			KDFParams:  nil,
		}

		result, err := recovery.SetupRecovery(config)
		if err != nil {
			t.Fatalf("SetupRecovery() with passphrase failed: %v", err)
		}

		// Verify PassphraseRequired flag is set
		if !result.Metadata.PassphraseRequired {
			t.Error("Metadata.PassphraseRequired should be true when passphrase provided")
		}
	})

	t.Run("passphrase not required when empty", func(t *testing.T) {
		config := &recovery.SetupConfig{
			Passphrase: []byte{}, // Empty passphrase
			KDFParams:  nil,
		}

		result, err := recovery.SetupRecovery(config)
		if err != nil {
			t.Fatalf("SetupRecovery() failed: %v", err)
		}

		// Verify PassphraseRequired flag is false
		if result.Metadata.PassphraseRequired {
			t.Error("Metadata.PassphraseRequired should be false for empty passphrase")
		}
	})

	t.Run("uses custom KDF parameters", func(t *testing.T) {
		customParams := &vault.KDFParams{
			Algorithm:     "argon2id",
			Time:          2,     // Custom iteration count
			Memory:        32768, // 32 MB
			Threads:       2,
			SaltChallenge: make([]byte, 32),
			SaltRecovery:  make([]byte, 32),
		}

		config := &recovery.SetupConfig{
			Passphrase: nil,
			KDFParams:  customParams,
		}

		result, err := recovery.SetupRecovery(config)
		if err != nil {
			t.Fatalf("SetupRecovery() with custom params failed: %v", err)
		}

		// Verify custom params are used
		if result.Metadata.KDFParams.Time != 2 {
			t.Errorf("Expected Time=2, got %d", result.Metadata.KDFParams.Time)
		}
		if result.Metadata.KDFParams.Memory != 32768 {
			t.Errorf("Expected Memory=32768, got %d", result.Metadata.KDFParams.Memory)
		}
	})

	t.Run("generates unique results on each call", func(t *testing.T) {
		config := &recovery.SetupConfig{
			Passphrase: nil,
			KDFParams:  nil,
		}

		result1, err := recovery.SetupRecovery(config)
		if err != nil {
			t.Fatalf("First SetupRecovery() failed: %v", err)
		}

		result2, err := recovery.SetupRecovery(config)
		if err != nil {
			t.Fatalf("Second SetupRecovery() failed: %v", err)
		}

		// Mnemonics should be different
		if result1.Mnemonic == result2.Mnemonic {
			t.Error("SetupRecovery() produced identical mnemonics (should be random)")
		}

		// Challenge positions should be different (highly likely)
		positionsEqual := true
		for i := range result1.Metadata.ChallengePositions {
			if result1.Metadata.ChallengePositions[i] != result2.Metadata.ChallengePositions[i] {
				positionsEqual = false
				break
			}
		}
		if positionsEqual {
			t.Error("SetupRecovery() produced identical challenge positions (should be random)")
		}
	})
}
