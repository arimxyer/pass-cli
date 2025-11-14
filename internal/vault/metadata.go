package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Metadata represents vault configuration stored in .meta.json
type Metadata struct {
	Version         string            `json:"version"`
	CreatedAt       time.Time         `json:"created_at"`
	LastModified    time.Time         `json:"last_modified"`
	KeychainEnabled bool              `json:"keychain_enabled"`
	AuditEnabled    bool              `json:"audit_enabled"`
	Recovery        *RecoveryMetadata `json:"recovery,omitempty"` // BIP39 recovery configuration
}

// RecoveryMetadata stores BIP39 mnemonic recovery configuration
type RecoveryMetadata struct {
	Enabled              bool       `json:"enabled"`                // Whether recovery is active
	Version              string     `json:"version"`                // Schema version ("1" initially)
	PassphraseRequired   bool       `json:"passphrase_required"`    // Whether 25th word was set
	ChallengePositions   []int      `json:"challenge_positions"`    // Indices of 6 words [0-23]
	KDFParams            KDFParams  `json:"kdf_params"`             // Cryptographic parameters
	EncryptedStoredWords []byte     `json:"encrypted_stored_words"` // 18 words (AES-GCM)
	NonceStored          []byte     `json:"nonce_stored"`           // GCM nonce (12 bytes)
	EncryptedRecoveryKey []byte     `json:"encrypted_recovery_key"` // Vault unlock key (AES-GCM)
	NonceRecovery        []byte     `json:"nonce_recovery"`         // GCM nonce (12 bytes)
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

// MetadataPath returns the metadata file path for a vault
func MetadataPath(vaultPath string) string {
	return vaultPath + ".meta.json"
}

// LoadMetadata loads metadata from disk, returns default if file missing
func LoadMetadata(vaultPath string) (*Metadata, error) {
	path := MetadataPath(vaultPath)
	// #nosec G304 -- vault path is user-controlled by design for CLI tool
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Legacy vault - return default metadata (FR-001)
		return &Metadata{
			Version:         "1.0",
			KeychainEnabled: false,
			AuditEnabled:    false,
			CreatedAt:       time.Time{}, // Zero value
			LastModified:    time.Time{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("corrupted metadata file: %w", err)
	}

	// FR-017: Warn about unknown metadata versions but continue (graceful degradation)
	if metadata.Version != "1.0" && metadata.Version != "1" {
		fmt.Fprintf(os.Stderr, "Warning: Unknown metadata version %q (expected \"1.0\"). Continuing with best-effort compatibility.\n", metadata.Version)
	}

	return &metadata, nil
}

// SaveMetadata writes metadata to disk
func SaveMetadata(vaultPath string, metadata *Metadata) error {
	metadata.LastModified = time.Now().UTC()
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = metadata.LastModified
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	path := MetadataPath(vaultPath)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// DeleteMetadata removes metadata file (used by vault remove)
func DeleteMetadata(vaultPath string) error {
	path := MetadataPath(vaultPath)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}
	return nil
}
