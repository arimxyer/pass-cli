package recovery_test

import (
	"strings"
	"testing"

	"github.com/arimxyer/pass-cli/internal/recovery"

	"github.com/tyler-smith/go-bip39"
)

// T013: Unit test for GenerateMnemonic()
// Tests: 256-bit entropy, 24 words, checksum validation

func TestGenerateMnemonic(t *testing.T) {
	t.Run("generates 24-word mnemonic", func(t *testing.T) {
		mnemonic, err := recovery.GenerateMnemonic()
		if err != nil {
			t.Fatalf("GenerateMnemonic() failed: %v", err)
		}

		words := strings.Fields(mnemonic)
		if len(words) != 24 {
			t.Errorf("Expected 24 words, got %d", len(words))
		}
	})

	t.Run("generates valid BIP39 mnemonic with correct checksum", func(t *testing.T) {
		mnemonic, err := recovery.GenerateMnemonic()
		if err != nil {
			t.Fatalf("GenerateMnemonic() failed: %v", err)
		}

		// Verify BIP39 checksum using library
		_, err = bip39.EntropyFromMnemonic(mnemonic)
		if err != nil {
			t.Errorf("Mnemonic has invalid checksum: %v", err)
		}
	})

	t.Run("generates unique mnemonics on each call", func(t *testing.T) {
		mnemonic1, err := recovery.GenerateMnemonic()
		if err != nil {
			t.Fatalf("First GenerateMnemonic() failed: %v", err)
		}

		mnemonic2, err := recovery.GenerateMnemonic()
		if err != nil {
			t.Fatalf("Second GenerateMnemonic() failed: %v", err)
		}

		if mnemonic1 == mnemonic2 {
			t.Error("GenerateMnemonic() produced identical mnemonics (should be random)")
		}
	})

	t.Run("uses 256-bit entropy", func(t *testing.T) {
		mnemonic, err := recovery.GenerateMnemonic()
		if err != nil {
			t.Fatalf("GenerateMnemonic() failed: %v", err)
		}

		// Extract entropy from mnemonic to verify size
		entropy, err := bip39.EntropyFromMnemonic(mnemonic)
		if err != nil {
			t.Fatalf("Failed to extract entropy: %v", err)
		}

		// 256-bit entropy = 32 bytes
		if len(entropy) != 32 {
			t.Errorf("Expected 32 bytes (256 bits) entropy, got %d bytes", len(entropy))
		}
	})

	t.Run("all words are from BIP39 wordlist", func(t *testing.T) {
		mnemonic, err := recovery.GenerateMnemonic()
		if err != nil {
			t.Fatalf("GenerateMnemonic() failed: %v", err)
		}

		words := strings.Fields(mnemonic)
		for i, word := range words {
			if !recovery.ValidateWord(word) {
				t.Errorf("Word %d (%q) is not in BIP39 wordlist", i+1, word)
			}
		}
	})
}
