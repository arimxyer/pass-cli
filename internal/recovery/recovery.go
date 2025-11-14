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

	// 6. Derive challenge key from seed + challenge words
	challengeSeed := append(seed, []byte(strings.Join(challengeWords, " "))...)
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
	// TODO: Implement in Phase 4 (T038)
	return nil, ErrRecoveryDisabled
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
