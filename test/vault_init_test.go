package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"pass-cli/internal/crypto"
	"pass-cli/internal/storage"
	"pass-cli/internal/vault"
)

// T015: Integration test: init with recovery creates v2 vault
func TestInitWithRecoveryCreatesV2Vault(t *testing.T) {
	// Create temp directory for vault
	tempDir, err := os.MkdirTemp("", "vault-init-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Create vault service
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize with recovery enabled
	password := []byte("Test123!@#Password")
	_, err = vs.InitializeWithRecovery(password, false, "", "", nil)
	if err != nil {
		t.Fatalf("InitializeWithRecovery() error = %v", err)
	}

	// Load raw vault file to check version
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault file: %v", err)
	}

	var encryptedVault struct {
		Metadata storage.VaultMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(vaultData, &encryptedVault); err != nil {
		t.Fatalf("Failed to parse vault file: %v", err)
	}

	// Verify version is 2
	if encryptedVault.Metadata.Version != 2 {
		t.Errorf("Vault version = %d, want 2", encryptedVault.Metadata.Version)
	}

	// Verify WrappedDEK fields are present
	if len(encryptedVault.Metadata.WrappedDEK) != 48 {
		t.Errorf("WrappedDEK length = %d, want 48", len(encryptedVault.Metadata.WrappedDEK))
	}
	if len(encryptedVault.Metadata.WrappedDEKNonce) != 12 {
		t.Errorf("WrappedDEKNonce length = %d, want 12", len(encryptedVault.Metadata.WrappedDEKNonce))
	}

	// Verify recovery metadata has version "2"
	meta, err := vault.LoadMetadata(vaultPath)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	if meta.Recovery == nil {
		t.Fatal("Recovery metadata is nil")
	}
	if meta.Recovery.Version != "2" {
		t.Errorf("Recovery.Version = %q, want \"2\"", meta.Recovery.Version)
	}
}

// T016: Integration test: init with --no-recovery creates v1 vault
func TestInitWithNoRecoveryCreatesV1Vault(t *testing.T) {
	// Create temp directory for vault
	tempDir, err := os.MkdirTemp("", "vault-init-norecovery-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Create vault service
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize WITHOUT recovery (standard init)
	password := []byte("Test123!@#Password")
	err = vs.Initialize(password, false, "", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Load raw vault file to check version
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault file: %v", err)
	}

	var encryptedVault struct {
		Metadata storage.VaultMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(vaultData, &encryptedVault); err != nil {
		t.Fatalf("Failed to parse vault file: %v", err)
	}

	// Verify version is 1 (legacy)
	if encryptedVault.Metadata.Version != 1 {
		t.Errorf("Vault version = %d, want 1", encryptedVault.Metadata.Version)
	}

	// Verify WrappedDEK fields are NOT present (v1 vault)
	if len(encryptedVault.Metadata.WrappedDEK) != 0 {
		t.Errorf("WrappedDEK should be empty for v1 vault, got length %d", len(encryptedVault.Metadata.WrappedDEK))
	}
}

// T017: Unit test: VaultMetadata v2 serialization
func TestVaultMetadataV2Serialization(t *testing.T) {
	// Create v2 metadata with wrapped DEK fields
	wrappedDEK := make([]byte, 48) // 32-byte DEK + 16-byte tag
	wrappedDEKNonce := make([]byte, 12)
	salt := make([]byte, 32)

	// Fill with test data
	for i := range wrappedDEK {
		wrappedDEK[i] = byte(i)
	}
	for i := range wrappedDEKNonce {
		wrappedDEKNonce[i] = byte(i + 100)
	}
	for i := range salt {
		salt[i] = byte(i + 200)
	}

	metadata := storage.VaultMetadata{
		Version:         2,
		Salt:            salt,
		Iterations:      crypto.DefaultIterations,
		WrappedDEK:      wrappedDEK,
		WrappedDEKNonce: wrappedDEKNonce,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	// Deserialize back
	var decoded storage.VaultMetadata
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	// Verify fields
	if decoded.Version != 2 {
		t.Errorf("Version = %d, want 2", decoded.Version)
	}
	if len(decoded.WrappedDEK) != 48 {
		t.Errorf("WrappedDEK length = %d, want 48", len(decoded.WrappedDEK))
	}
	if len(decoded.WrappedDEKNonce) != 12 {
		t.Errorf("WrappedDEKNonce length = %d, want 12", len(decoded.WrappedDEKNonce))
	}

	// Verify data integrity
	for i := range wrappedDEK {
		if decoded.WrappedDEK[i] != wrappedDEK[i] {
			t.Errorf("WrappedDEK[%d] = %d, want %d", i, decoded.WrappedDEK[i], wrappedDEK[i])
		}
	}
}

// Test v1 metadata omits wrapped DEK fields
func TestVaultMetadataV1OmitsWrappedDEK(t *testing.T) {
	metadata := storage.VaultMetadata{
		Version:    1,
		Salt:       make([]byte, 32),
		Iterations: crypto.DefaultIterations,
		// WrappedDEK and WrappedDEKNonce intentionally left empty
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	// Check JSON doesn't contain wrapped_dek fields (omitempty)
	jsonStr := string(jsonData)
	if contains(jsonStr, "wrapped_dek") {
		t.Error("V1 metadata JSON should not contain wrapped_dek (omitempty)")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
