package recovery

import "pass-cli/internal/vault"

// deriveKey performs Argon2id key derivation
// Parameters: seed (BIP39 seed), salt (32 bytes), params (KDF configuration)
// Returns: 32-byte derived key
func deriveKey(seed, salt []byte, params *vault.KDFParams) []byte {
	// TODO: Implement in Phase 3 (T023)
	return nil
}

// encryptStoredWords encrypts 18 stored words with AES-256-GCM
// Parameters: words (18-word array), key (32-byte encryption key)
// Returns: ciphertext, nonce (12 bytes), error
func encryptStoredWords(words []string, key []byte) (ciphertext, nonce []byte, err error) {
	// TODO: Implement in Phase 3 (T024)
	return nil, nil, ErrEncryptionFailed
}

// decryptStoredWords decrypts 18 stored words with AES-256-GCM
// Parameters: ciphertext, nonce (12 bytes), key (32-byte decryption key)
// Returns: 18-word array, error
func decryptStoredWords(ciphertext, nonce, key []byte) ([]string, error) {
	// TODO: Implement in Phase 3 (T025)
	return nil, ErrDecryptionFailed
}
