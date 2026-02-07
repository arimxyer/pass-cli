package recovery

import (
	"strings"

	"github.com/arimxyer/pass-cli/internal/crypto"

	"github.com/tyler-smith/go-bip39"
)

// GenerateMnemonic generates a 24-word BIP39 mnemonic phrase
// Returns: mnemonic string (space-separated 24 words), error
func GenerateMnemonic() (string, error) {
	// Generate 256-bit entropy (32 bytes) for 24-word mnemonic
	entropy, err := bip39.NewEntropy(EntropyBits)
	if err != nil {
		return "", ErrEntropyGeneration
	}
	defer crypto.ClearBytes(entropy) // Clear entropy from memory after use

	// Generate mnemonic from entropy
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", ErrMnemonicGeneration
	}

	return mnemonic, nil
}

// ValidateWord checks if a word is in the BIP39 English wordlist
// Parameters: word to validate (case-insensitive)
// Returns: true if word in wordlist, false otherwise
func ValidateWord(word string) bool {
	// BIP39 wordlist is case-insensitive, normalize to lowercase
	word = strings.ToLower(strings.TrimSpace(word))

	// Get the BIP39 English wordlist
	wordlist := bip39.GetWordList()

	// Check if word exists in wordlist
	for _, validWord := range wordlist {
		if word == validWord {
			return true
		}
	}

	return false
}

// ValidateMnemonic validates a mnemonic's BIP39 checksum
// Parameters: mnemonic string to validate
// Returns: true if checksum valid, false otherwise
func ValidateMnemonic(mnemonic string) bool {
	// Use BIP39 library to validate checksum
	// This verifies the mnemonic is well-formed and has a valid checksum
	return bip39.IsMnemonicValid(mnemonic)
}
