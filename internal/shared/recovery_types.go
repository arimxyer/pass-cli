// Package shared contains types shared across multiple internal packages
// to avoid import cycles.
package shared

// RecoveryMetadata stores BIP39 mnemonic recovery configuration
type RecoveryMetadata struct {
	Enabled              bool      `json:"enabled"`                // Whether recovery is active
	Version              string    `json:"version"`                // Schema version ("1" for v1, "2" for v2 with challenge)
	PassphraseRequired   bool      `json:"passphrase_required"`    // Whether 25th word was set
	ChallengePositions   []int     `json:"challenge_positions"`    // Indices of 6 words [0-23]
	KDFParams            KDFParams `json:"kdf_params"`             // Cryptographic parameters
	EncryptedStoredWords []byte    `json:"encrypted_stored_words"` // 18 words (AES-GCM)
	NonceStored          []byte    `json:"nonce_stored"`           // GCM nonce (12 bytes)
	EncryptedRecoveryKey []byte    `json:"encrypted_recovery_key"` // Vault unlock key (AES-GCM)
	NonceRecovery        []byte    `json:"nonce_recovery"`         // GCM nonce (12 bytes)
}

// KDFParams stores Argon2id key derivation function parameters
type KDFParams struct {
	Algorithm     string `json:"algorithm"`      // "argon2id" (fixed)
	Time          uint32 `json:"time"`           // Iteration count (1)
	Memory        uint32 `json:"memory"`         // Memory cost in KiB (65536 = 64 MB)
	Threads       uint8  `json:"threads"`        // Parallelism (4)
	SaltChallenge []byte `json:"salt_challenge"` // 32-byte salt for challenge KDF
	SaltRecovery  []byte `json:"salt_recovery"`  // 32-byte salt for recovery KDF
}
