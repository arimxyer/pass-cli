package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"pass-cli/internal/shared"
)

// Metadata represents vault configuration stored in .meta.json
type Metadata struct {
	Version         string                   `json:"version"`
	CreatedAt       time.Time                `json:"created_at"`
	LastModified    time.Time                `json:"last_modified"`
	KeychainEnabled bool                     `json:"keychain_enabled"`
	AuditEnabled    bool                     `json:"audit_enabled"`
	Recovery        *shared.RecoveryMetadata `json:"recovery,omitempty"` // BIP39 recovery configuration
}

// RecoveryMetadata is an alias for shared.RecoveryMetadata for backward compatibility
type RecoveryMetadata = shared.RecoveryMetadata

// KDFParams is an alias for shared.KDFParams for backward compatibility
type KDFParams = shared.KDFParams

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
