package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// VaultMetadata contains plaintext audit configuration for a vault.
// Stored in vault.meta file alongside vault.enc to enable audit logging
// for operations that don't unlock the vault (keychain status, vault remove).
type VaultMetadata struct {
	VaultID      string    `json:"vault_id"`        // Absolute path to vault.enc
	AuditEnabled bool      `json:"audit_enabled"`   // Whether audit logging is enabled
	AuditLogPath string    `json:"audit_log_path"`  // Absolute path to audit.log
	CreatedAt    time.Time `json:"created_at"`      // Metadata creation timestamp
	Version      int       `json:"version"`         // Metadata format version (currently 1)
}

// LoadMetadata reads vault.meta file and returns VaultMetadata.
// Returns (nil, nil) if file doesn't exist (not an error - backward compatibility).
// Returns error if file exists but is corrupted/invalid.
//
// T006: Implementation for LoadMetadata
func LoadMetadata(vaultPath string) (*VaultMetadata, error) {
	metaPath := MetadataPath(vaultPath)

	// Check if file exists
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Not an error - backward compatibility
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// Parse JSON
	var meta VaultMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("invalid metadata JSON: %w", err)
	}

	// Validate version
	if meta.Version < 1 {
		return nil, fmt.Errorf("invalid metadata version: %d", meta.Version)
	}
	if meta.Version > 1 {
		// Forward compatibility: log warning but attempt best-effort parsing
		fmt.Fprintf(os.Stderr, "Warning: Unknown metadata version %d (expected 1), attempting best-effort parsing\n", meta.Version)
	}

	// Validate required fields
	if meta.VaultID == "" || meta.CreatedAt.IsZero() {
		return nil, fmt.Errorf("metadata missing required fields")
	}

	// Validate conditional fields (FR-002)
	if meta.AuditEnabled && meta.AuditLogPath == "" {
		return nil, fmt.Errorf("audit enabled but audit_log_path missing")
	}

	return &meta, nil
}

// SaveMetadata writes VaultMetadata to vault.meta file atomically.
// Uses temp file + rename pattern for atomic write.
// Creates parent directory if needed.
//
// T007: Implementation for SaveMetadata
func SaveMetadata(meta *VaultMetadata, vaultPath string) error {
	metaPath := MetadataPath(vaultPath)
	dir := filepath.Dir(metaPath)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Serialize to JSON (indented for readability)
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write to temp file
	tmpPath := filepath.Join(dir, ".vault.meta.tmp")
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp metadata: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, metaPath); err != nil {
		os.Remove(tmpPath) // Cleanup on failure
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// DeleteMetadata removes vault.meta file.
// Returns nil if file doesn't exist (idempotent).
//
// T008: Implementation for DeleteMetadata
func DeleteMetadata(vaultPath string) error {
	metaPath := MetadataPath(vaultPath)

	err := os.Remove(metaPath)
	if err != nil && os.IsNotExist(err) {
		return nil // Idempotent - not an error
	}
	return err
}

// MetadataPath returns the path to vault.meta for given vault.enc path.
// Example: "/path/to/vault.enc" → "/path/to/vault.meta"
//
// T009: Implementation for MetadataPath
func MetadataPath(vaultPath string) string {
	dir := filepath.Dir(vaultPath)
	return filepath.Join(dir, "vault.meta")
}
