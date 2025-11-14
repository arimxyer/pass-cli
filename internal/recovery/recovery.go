package recovery

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"strings"

	"pass-cli/internal/crypto"
	"pass-cli/internal/vault"

	"github.com/tyler-smith/go-bip39"
)

// SetupConfig configures recovery setup during vault initialization
type SetupConfig struct {
	Passphrase []byte            // Optional passphrase (25th word). Empty = no passphrase.
	KDFParams  *vault.KDFParams  // Custom Argon2 parameters (optional, uses defaults if nil)
}

// SetupResult contains the result of recovery setup operation
type SetupResult struct {
	Mnemonic         string                 // 24-word mnemonic to display to user
	Metadata         *vault.RecoveryMetadata // Recovery metadata to store in vault
	VaultRecoveryKey []byte                 // Vault recovery key (32 bytes) to integrate with vault unlocking
}

// RecoveryConfig configures recovery execution
type RecoveryConfig struct {
	ChallengeWords []string                // Words entered by user (from challenge positions)
	Passphrase     []byte                  // Passphrase (if required). Empty = no passphrase.
	Metadata       *vault.RecoveryMetadata // Recovery metadata from vault
}

// VerifyConfig configures backup verification during init
type VerifyConfig struct {
	Mnemonic        string // Full 24-word mnemonic (generated during setup)
	VerifyPositions []int  // Positions to prompt for (randomly selected)
	UserWords       []string // Words entered by user
}

// SetupRecovery generates BIP39 mnemonic and prepares recovery metadata
// Parameters: config (setup configuration)
// Returns: SetupResult (mnemonic, metadata, vault recovery key), error
func SetupRecovery(config *SetupConfig) (*SetupResult, error) {
	// 1. Generate 24-word mnemonic
	mnemonic, err := GenerateMnemonic()
	if err != nil {
		return nil, err
	}

	// 2. Generate BIP39 seed from mnemonic (with optional passphrase)
	seed := bip39.NewSeed(mnemonic, string(config.Passphrase))
	defer crypto.ClearBytes(seed) // Clear seed from memory

	// 3. Select 6 challenge positions (crypto-secure random)
	challengePositions, err := selectChallengePositions(MnemonicWords, ChallengeCount)
	if err != nil {
		return nil, err
	}

	// 4. Split mnemonic into challenge words (6) and stored words (18)
	challengeWords, storedWords := splitWords(mnemonic, challengePositions)

	// 5. Setup KDF parameters (use defaults if not provided)
	kdfParams := config.KDFParams
	if kdfParams == nil {
		// Generate default KDF parameters with random salts
		saltChallenge := make([]byte, DefaultSaltLen)
		saltRecovery := make([]byte, DefaultSaltLen)
		if _, err := rand.Read(saltChallenge); err != nil {
			return nil, ErrRandomGeneration
		}
		if _, err := rand.Read(saltRecovery); err != nil {
			return nil, ErrRandomGeneration
		}

		kdfParams = &vault.KDFParams{
			Algorithm:     "argon2id",
			Time:          DefaultTime,
			Memory:        DefaultMemory,
			Threads:       DefaultThreads,
			SaltChallenge: saltChallenge,
			SaltRecovery:  saltRecovery,
		}
	}

	// 6. Derive challenge key from challenge words only (as mnemonic)
	// This allows recovery to derive the same key with just the 6 words
	challengeMnemonic := strings.Join(challengeWords, " ")
	challengeSeed := bip39.NewSeed(challengeMnemonic, string(config.Passphrase))
	defer crypto.ClearBytes(challengeSeed)
	challengeKey := deriveKey(challengeSeed, kdfParams.SaltChallenge, kdfParams)
	defer crypto.ClearBytes(challengeKey)

	// 7. Encrypt stored words (18) with challenge key
	encryptedStoredWords, nonceStored, err := encryptStoredWords(storedWords, challengeKey)
	if err != nil {
		return nil, err
	}

	// 8. Generate vault recovery key (32 bytes random)
	vaultRecoveryKey := make([]byte, DefaultKeyLen)
	if _, err := rand.Read(vaultRecoveryKey); err != nil {
		return nil, ErrRandomGeneration
	}

	// 9. Derive recovery key from full seed (for encrypting vault recovery key)
	recoveryKey := deriveKey(seed, kdfParams.SaltRecovery, kdfParams)
	defer crypto.ClearBytes(recoveryKey)

	// 10. Encrypt vault recovery key with recovery key
	encryptedRecoveryKey, nonceRecovery, err := encryptData(vaultRecoveryKey, recoveryKey)
	if err != nil {
		return nil, err
	}

	// 11. Build recovery metadata
	metadata := &vault.RecoveryMetadata{
		Enabled:              true,
		Version:              "1",
		PassphraseRequired:   len(config.Passphrase) > 0,
		ChallengePositions:   challengePositions,
		KDFParams:            *kdfParams,
		EncryptedStoredWords: encryptedStoredWords,
		NonceStored:          nonceStored,
		EncryptedRecoveryKey: encryptedRecoveryKey,
		NonceRecovery:        nonceRecovery,
	}

	// 12. Return result
	result := &SetupResult{
		Mnemonic:         mnemonic,
		Metadata:         metadata,
		VaultRecoveryKey: vaultRecoveryKey,
	}

	return result, nil
}

// PerformRecovery recovers vault access using challenge words
// Parameters: config (recovery configuration)
// Returns: vault recovery key (32 bytes), error
func PerformRecovery(config *RecoveryConfig) ([]byte, error) {
	// 1. Validate recovery is enabled
	if config.Metadata == nil || !config.Metadata.Enabled {
		return nil, ErrRecoveryDisabled
	}

	// 2. Validate word count (must be 6 challenge words)
	if len(config.ChallengeWords) != ChallengeCount {
		return nil, ErrInvalidCount
	}

	// 3. Validate all words are in BIP39 wordlist
	for _, word := range config.ChallengeWords {
		if !ValidateWord(word) {
			return nil, ErrInvalidWord
		}
	}

	// 4. Detect metadata corruption (FR-033)
	if len(config.Metadata.NonceStored) != GCMNonceSize {
		return nil, ErrMetadataCorrupted
	}
	if len(config.Metadata.NonceRecovery) != GCMNonceSize {
		return nil, ErrMetadataCorrupted
	}
	if len(config.Metadata.EncryptedStoredWords) == 0 {
		return nil, ErrMetadataCorrupted
	}
	if len(config.Metadata.EncryptedRecoveryKey) == 0 {
		return nil, ErrMetadataCorrupted
	}
	if len(config.Metadata.ChallengePositions) != ChallengeCount {
		return nil, ErrMetadataCorrupted
	}
	for _, pos := range config.Metadata.ChallengePositions {
		if pos < 0 || pos >= MnemonicWords {
			return nil, ErrMetadataCorrupted
		}
	}

	// 5. Normalize challenge words (lowercase, trim)
	normalizedWords := make([]string, len(config.ChallengeWords))
	for i, word := range config.ChallengeWords {
		normalizedWords[i] = strings.ToLower(strings.TrimSpace(word))
	}

	// 6. Generate BIP39 seed from challenge words (with optional passphrase)
	// Note: We need to reconstruct the partial seed from challenge words first
	challengeMnemonic := strings.Join(normalizedWords, " ")
	challengeSeed := bip39.NewSeed(challengeMnemonic, string(config.Passphrase))
	defer crypto.ClearBytes(challengeSeed)

	// 7. Derive challenge key from seed
	challengeKey := deriveKey(
		challengeSeed,
		config.Metadata.KDFParams.SaltChallenge,
		&config.Metadata.KDFParams,
	)
	defer crypto.ClearBytes(challengeKey)

	// 8. Decrypt stored words (18) with challenge key
	storedWords, err := decryptStoredWords(
		config.Metadata.EncryptedStoredWords,
		config.Metadata.NonceStored,
		challengeKey,
	)
	if err != nil {
		// Decryption failed = wrong challenge words or wrong passphrase
		return nil, ErrDecryptionFailed
	}

	// 9. Reconstruct full 24-word mnemonic
	fullMnemonic, err := reconstructMnemonic(
		normalizedWords,
		config.Metadata.ChallengePositions,
		storedWords,
	)
	if err != nil {
		return nil, err
	}

	// 10. Validate mnemonic checksum
	if !ValidateMnemonic(fullMnemonic) {
		return nil, ErrInvalidMnemonic
	}

	// 11. Generate full BIP39 seed from complete mnemonic (with passphrase)
	fullSeed := bip39.NewSeed(fullMnemonic, string(config.Passphrase))
	defer crypto.ClearBytes(fullSeed)

	// 12. Derive recovery key from full seed
	recoveryKey := deriveKey(
		fullSeed,
		config.Metadata.KDFParams.SaltRecovery,
		&config.Metadata.KDFParams,
	)
	defer crypto.ClearBytes(recoveryKey)

	// 13. Decrypt vault recovery key
	vaultRecoveryKey, err := decryptData(
		config.Metadata.EncryptedRecoveryKey,
		config.Metadata.NonceRecovery,
		recoveryKey,
	)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// 14. Return vault recovery key (caller is responsible for clearing)
	return vaultRecoveryKey, nil
}

// decryptData is a helper to decrypt arbitrary data with AES-256-GCM
func decryptData(ciphertext, nonce, key []byte) ([]byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// Decrypt and verify authentication
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// VerifyBackup verifies user wrote down mnemonic correctly
// Parameters: config (verification configuration)
// Returns: error (nil if verification passes, ErrVerificationFailed if words mismatch)
func VerifyBackup(config *VerifyConfig) error {
	// Split mnemonic into words
	words := strings.Fields(config.Mnemonic)

	// Validate we have 24 words
	if len(words) != MnemonicWords {
		return ErrInvalidMnemonic
	}

	// Validate position count matches user words count
	if len(config.VerifyPositions) != len(config.UserWords) {
		return ErrVerificationFailed
	}

	// Extract expected words at verify positions
	for i, pos := range config.VerifyPositions {
		// Validate position is in range
		if pos < 0 || pos >= len(words) {
			return ErrInvalidPositions
		}

		// Get expected word at this position
		expectedWord := strings.ToLower(strings.TrimSpace(words[pos]))
		userWord := strings.ToLower(strings.TrimSpace(config.UserWords[i]))

		// Compare (case-insensitive, trimmed)
		if expectedWord != userWord {
			return ErrVerificationFailed
		}

		// Validate user word is empty
		if userWord == "" {
			return ErrVerificationFailed
		}
	}

	return nil
}

// encryptData is a helper to encrypt arbitrary data with AES-256-GCM
func encryptData(plaintext, key []byte) (ciphertext, nonce []byte, err error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, ErrEncryptionFailed
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, ErrEncryptionFailed
	}

	// Generate random nonce
	nonce = make([]byte, GCMNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, ErrEncryptionFailed
	}

	// Encrypt and authenticate
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nonce, nil
}
