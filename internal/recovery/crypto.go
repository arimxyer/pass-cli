package recovery

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"strings"

	"pass-cli/internal/vault"

	"golang.org/x/crypto/argon2"
)

// deriveKey performs Argon2id key derivation
// Parameters: seed (BIP39 seed), salt (32 bytes), params (KDF configuration)
// Returns: 32-byte derived key
func deriveKey(seed, salt []byte, params *vault.KDFParams) []byte {
	// Use Argon2id with parameters from config
	key := argon2.IDKey(
		seed,
		salt,
		params.Time,
		params.Memory,
		params.Threads,
		DefaultKeyLen,
	)

	return key
}

// encryptStoredWords encrypts 18 stored words with AES-256-GCM
// Parameters: words (18-word array), key (32-byte encryption key)
// Returns: ciphertext, nonce (12 bytes), error
func encryptStoredWords(words []string, key []byte) (ciphertext, nonce []byte, err error) {
	// Serialize words to JSON
	plaintext, err := json.Marshal(words)
	if err != nil {
		return nil, nil, ErrEncryptionFailed
	}

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
	// #nosec G407 - False positive: nonce is cryptographically generated above, not hardcoded
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nonce, nil
}

// decryptStoredWords decrypts 18 stored words with AES-256-GCM
// Parameters: ciphertext, nonce (12 bytes), key (32-byte decryption key)
// Returns: 18-word array, error
func decryptStoredWords(ciphertext, nonce, key []byte) ([]string, error) {
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

	// Deserialize JSON to words array
	var words []string
	if err := json.Unmarshal(plaintext, &words); err != nil {
		return nil, ErrDecryptionFailed
	}

	// Validate word count
	if len(words) != 18 {
		return nil, ErrDecryptionFailed
	}

	// Validate each word against BIP39 wordlist
	for _, word := range words {
		if !ValidateWord(strings.TrimSpace(word)) {
			return nil, ErrInvalidWord
		}
	}

	return words, nil
}
