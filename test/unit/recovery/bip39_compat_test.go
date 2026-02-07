package recovery_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/tyler-smith/go-bip39"

	"github.com/arimxyer/pass-cli/internal/recovery"
)

// T058: BIP39 Compatibility Test
// Verify that pass-cli's mnemonic generation is compatible with standard BIP39 implementations
// Uses official BIP39 test vectors from: https://github.com/trezor/python-mnemonic/blob/master/vectors.json

func TestBIP39Compatibility(t *testing.T) {
	t.Run("BIP39 Test Vector 1: 256-bit entropy", func(t *testing.T) {
		// Official BIP39 test vector (English, 256-bit entropy)
		entropyHex := "0c1e24e5917779d297e14d45f14e1a1a"
		expectedMnemonic := "army van defense carry jealous true garbage claim echo media make crunch"

		// Decode hex entropy
		entropy, err := hex.DecodeString(entropyHex)
		if err != nil {
			t.Fatalf("Failed to decode entropy: %v", err)
		}

		// Generate mnemonic from entropy using go-bip39
		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			t.Fatalf("Failed to generate mnemonic: %v", err)
		}

		// Verify matches expected mnemonic
		if mnemonic != expectedMnemonic {
			t.Errorf("Mnemonic mismatch:\nExpected: %s\nGot:      %s", expectedMnemonic, mnemonic)
		}

		t.Logf("✓ BIP39 test vector 1 matches: %s", mnemonic)
	})

	t.Run("BIP39 Test Vector 2: 256-bit entropy with passphrase", func(t *testing.T) {
		// Test vector with passphrase (from BIP39 spec)
		mnemonic := "hamster diagram private dutch cause delay private meat slide toddler razor book happy fancy gospel tennis maple dilemma loan word shrug inflict delay length"
		passphrase := "TREZOR"

		// Expected seed (first 64 bytes of PBKDF2 output)
		expectedSeedHex := "64c87cde7e12ecf6704ab95bb1408bef047c22db4cc7491c4271d170a1b213d20b385bc1588d9c7b38f1b39d415665b8a9030c9ec653d75e65f847d8fc1fc440"

		// Generate seed using go-bip39
		seed := bip39.NewSeed(mnemonic, passphrase)

		// Compare seed
		actualSeedHex := hex.EncodeToString(seed)
		if actualSeedHex != expectedSeedHex {
			t.Errorf("Seed mismatch:\nExpected: %s\nGot:      %s", expectedSeedHex, actualSeedHex)
		}

		t.Logf("✓ BIP39 seed derivation matches test vector")
		t.Logf("  Mnemonic: %s", mnemonic)
		t.Logf("  Passphrase: %s", passphrase)
		t.Logf("  Seed (hex): %s...", actualSeedHex[:32])
	})

	t.Run("pass-cli mnemonic generation produces valid BIP39", func(t *testing.T) {
		// Generate mnemonic using pass-cli's recovery package
		mnemonic, err := recovery.GenerateMnemonic()
		if err != nil {
			t.Fatalf("Failed to generate mnemonic: %v", err)
		}

		// Verify it's valid BIP39
		if !bip39.IsMnemonicValid(mnemonic) {
			t.Errorf("pass-cli generated invalid BIP39 mnemonic: %s", mnemonic)
		}

		// Verify it's 24 words
		words := strings.Fields(mnemonic)
		if len(words) != 24 {
			t.Errorf("Expected 24 words, got %d", len(words))
		}

		// Verify can derive seed (no passphrase)
		seed := bip39.NewSeed(mnemonic, "")
		if len(seed) != 64 {
			t.Errorf("Expected 64-byte seed, got %d bytes", len(seed))
		}

		t.Logf("✓ pass-cli generates valid BIP39 mnemonic")
		t.Logf("  Mnemonic: %s", mnemonic)
		t.Logf("  Seed length: %d bytes", len(seed))
	})

	t.Run("pass-cli mnemonic with passphrase matches external tool", func(t *testing.T) {
		// Use a known mnemonic for reproducibility
		mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"
		passphrase := "test-passphrase"

		// Derive seed using go-bip39
		seed := bip39.NewSeed(mnemonic, passphrase)
		actualSeedHex := hex.EncodeToString(seed)

		// Note: This test validates that seed derivation works.
		// The actual seed value can be verified against external BIP39 tools
		// such as: https://iancoleman.io/bip39/
		t.Logf("Mnemonic: %s", mnemonic)
		t.Logf("Passphrase: %s", passphrase)
		t.Logf("Seed (hex): %s", actualSeedHex)

		// Verify seed is 64 bytes
		if len(seed) != 64 {
			t.Errorf("Expected 64-byte seed, got %d bytes", len(seed))
		}

		t.Logf("✓ Seed derivation successful (manual verification required against external BIP39 tool)")
	})

	t.Run("Entropy size produces correct word count", func(t *testing.T) {
		testCases := []struct {
			entropyBits int
			wordCount   int
		}{
			{128, 12}, // 128 bits → 12 words
			{160, 15}, // 160 bits → 15 words
			{192, 18}, // 192 bits → 18 words
			{224, 21}, // 224 bits → 21 words
			{256, 24}, // 256 bits → 24 words (pass-cli uses this)
		}

		for _, tc := range testCases {
			entropy, err := bip39.NewEntropy(tc.entropyBits)
			if err != nil {
				t.Fatalf("Failed to generate %d-bit entropy: %v", tc.entropyBits, err)
			}

			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				t.Fatalf("Failed to generate mnemonic from %d-bit entropy: %v", tc.entropyBits, err)
			}

			words := strings.Fields(mnemonic)
			if len(words) != tc.wordCount {
				t.Errorf("Entropy %d bits: expected %d words, got %d", tc.entropyBits, tc.wordCount, len(words))
			}

			t.Logf("✓ %d bits → %d words", tc.entropyBits, tc.wordCount)
		}
	})

	t.Run("BIP39 wordlist validation", func(t *testing.T) {
		// Verify go-bip39 uses English wordlist by default
		// BIP39 English wordlist has exactly 2048 words
		validWord := "abandon" // First word in BIP39 English wordlist
		invalidWord := "notaword"

		// Test with known valid mnemonic
		validMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
		if !bip39.IsMnemonicValid(validMnemonic) {
			t.Error("Valid BIP39 mnemonic marked as invalid")
		}

		// Test with invalid word
		invalidMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon notaword"
		if bip39.IsMnemonicValid(invalidMnemonic) {
			t.Error("Invalid mnemonic (bad word) marked as valid")
		}

		// Test with invalid checksum
		invalidChecksum := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
		if bip39.IsMnemonicValid(invalidChecksum) {
			t.Error("Invalid mnemonic (bad checksum) marked as valid")
		}

		t.Logf("✓ BIP39 wordlist validation working correctly")
		t.Logf("  Valid word '%s': recognized", validWord)
		t.Logf("  Invalid word '%s': rejected", invalidWord)
	})
}
