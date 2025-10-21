package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// T001: Unit tests for LoadMetadata
func TestLoadMetadata_Success(t *testing.T) {
	// Create temp directory for test vault
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")
	metaPath := MetadataPath(vaultPath)

	// Create valid metadata file
	meta := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: true,
		AuditLogPath: filepath.Join(tmpDir, "audit.log"),
		CreatedAt:    time.Now().UTC(),
		Version:      1,
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test metadata: %v", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test metadata: %v", err)
	}

	// Test LoadMetadata
	loaded, err := LoadMetadata(vaultPath)
	if err != nil {
		t.Fatalf("LoadMetadata failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("LoadMetadata returned nil metadata")
	}

	if loaded.VaultID != meta.VaultID {
		t.Errorf("Expected VaultID %q, got %q", meta.VaultID, loaded.VaultID)
	}

	if loaded.AuditEnabled != meta.AuditEnabled {
		t.Errorf("Expected AuditEnabled %v, got %v", meta.AuditEnabled, loaded.AuditEnabled)
	}

	if loaded.AuditLogPath != meta.AuditLogPath {
		t.Errorf("Expected AuditLogPath %q, got %q", meta.AuditLogPath, loaded.AuditLogPath)
	}

	if loaded.Version != meta.Version {
		t.Errorf("Expected Version %d, got %d", meta.Version, loaded.Version)
	}
}

func TestLoadMetadata_NotFound(t *testing.T) {
	// Use non-existent vault path
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "nonexistent.enc")

	// Test LoadMetadata on missing file
	loaded, err := LoadMetadata(vaultPath)
	if err != nil {
		t.Fatalf("LoadMetadata should not error on missing file: %v", err)
	}

	if loaded != nil {
		t.Fatal("LoadMetadata should return nil for missing file")
	}
}

func TestLoadMetadata_Corrupted(t *testing.T) {
	// Create temp directory for test vault
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")
	metaPath := MetadataPath(vaultPath)

	// Write invalid JSON
	if err := os.WriteFile(metaPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to write corrupted metadata: %v", err)
	}

	// Test LoadMetadata on corrupted file
	loaded, err := LoadMetadata(vaultPath)
	if err == nil {
		t.Fatal("LoadMetadata should return error for corrupted file")
	}

	if loaded != nil {
		t.Fatal("LoadMetadata should return nil for corrupted file")
	}
}

func TestLoadMetadata_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")
	metaPath := MetadataPath(vaultPath)

	// Create metadata missing required fields
	invalidMeta := map[string]interface{}{
		"vault_id": vaultPath,
		// Missing audit_enabled, created_at, version
	}

	data, _ := json.MarshalIndent(invalidMeta, "", "  ")
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test metadata: %v", err)
	}

	// Test LoadMetadata
	loaded, err := LoadMetadata(vaultPath)
	if err == nil {
		t.Fatal("LoadMetadata should return error for missing required fields")
	}

	if loaded != nil {
		t.Fatal("LoadMetadata should return nil for invalid metadata")
	}
}

func TestLoadMetadata_AuditEnabledWithoutLogPath(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")
	metaPath := MetadataPath(vaultPath)

	// Create metadata with audit_enabled=true but no audit_log_path
	meta := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: true,
		AuditLogPath: "", // Missing when audit enabled
		CreatedAt:    time.Now().UTC(),
		Version:      1,
	}

	data, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test metadata: %v", err)
	}

	// Test LoadMetadata
	loaded, err := LoadMetadata(vaultPath)
	if err == nil {
		t.Fatal("LoadMetadata should return error when audit enabled but no log path")
	}

	if loaded != nil {
		t.Fatal("LoadMetadata should return nil for invalid metadata")
	}
}

func TestLoadMetadata_UnknownVersion(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")
	metaPath := MetadataPath(vaultPath)

	// Create metadata with unknown version
	meta := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: false,
		AuditLogPath: "",
		CreatedAt:    time.Now().UTC(),
		Version:      99, // Unknown version
	}

	data, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test metadata: %v", err)
	}

	// Test LoadMetadata - should log warning but attempt parsing
	loaded, err := LoadMetadata(vaultPath)
	if err != nil {
		t.Fatalf("LoadMetadata should attempt best-effort parsing for unknown version: %v", err)
	}

	if loaded == nil {
		t.Fatal("LoadMetadata should return metadata for unknown version")
	}

	if loaded.Version != 99 {
		t.Errorf("Expected Version 99, got %d", loaded.Version)
	}
}

// T002: Unit tests for SaveMetadata
func TestSaveMetadata_Success(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// Create metadata
	meta := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: true,
		AuditLogPath: filepath.Join(tmpDir, "audit.log"),
		CreatedAt:    time.Now().UTC(),
		Version:      1,
	}

	// Test SaveMetadata
	if err := SaveMetadata(meta, vaultPath); err != nil {
		t.Fatalf("SaveMetadata failed: %v", err)
	}

	// Verify file exists
	metaPath := MetadataPath(vaultPath)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata file was not created")
	}

	// Verify content
	loaded, err := LoadMetadata(vaultPath)
	if err != nil {
		t.Fatalf("Failed to load saved metadata: %v", err)
	}

	if loaded.VaultID != meta.VaultID {
		t.Errorf("Expected VaultID %q, got %q", meta.VaultID, loaded.VaultID)
	}
}

func TestSaveMetadata_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Nested directory that doesn't exist yet
	nestedDir := filepath.Join(tmpDir, "subdir", "nested")
	vaultPath := filepath.Join(nestedDir, "vault.enc")

	meta := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: false,
		AuditLogPath: "",
		CreatedAt:    time.Now().UTC(),
		Version:      1,
	}

	// Test SaveMetadata creates parent directory
	if err := SaveMetadata(meta, vaultPath); err != nil {
		t.Fatalf("SaveMetadata failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Fatal("SaveMetadata did not create parent directory")
	}

	// Verify file exists
	metaPath := MetadataPath(vaultPath)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata file was not created")
	}
}

func TestSaveMetadata_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	meta := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: true,
		AuditLogPath: filepath.Join(tmpDir, "audit.log"),
		CreatedAt:    time.Now().UTC(),
		Version:      1,
	}

	// Save metadata
	if err := SaveMetadata(meta, vaultPath); err != nil {
		t.Fatalf("SaveMetadata failed: %v", err)
	}

	// Verify temp file doesn't exist (atomic rename completed)
	tmpPath := filepath.Join(tmpDir, ".vault.meta.tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("Temp file should be removed after atomic rename")
	}

	// Verify final file exists
	metaPath := MetadataPath(vaultPath)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata file was not created")
	}
}

func TestSaveMetadata_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// Save initial metadata
	meta1 := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: false,
		AuditLogPath: "",
		CreatedAt:    time.Now().UTC(),
		Version:      1,
	}

	if err := SaveMetadata(meta1, vaultPath); err != nil {
		t.Fatalf("Initial SaveMetadata failed: %v", err)
	}

	// Overwrite with updated metadata
	meta2 := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: true,
		AuditLogPath: filepath.Join(tmpDir, "audit.log"),
		CreatedAt:    meta1.CreatedAt, // Preserve original timestamp
		Version:      1,
	}

	if err := SaveMetadata(meta2, vaultPath); err != nil {
		t.Fatalf("SaveMetadata overwrite failed: %v", err)
	}

	// Verify updated content
	loaded, err := LoadMetadata(vaultPath)
	if err != nil {
		t.Fatalf("Failed to load updated metadata: %v", err)
	}

	if !loaded.AuditEnabled {
		t.Error("Metadata was not updated")
	}
}

// T003: Unit tests for DeleteMetadata
func TestDeleteMetadata_Success(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// Create metadata file
	meta := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: true,
		AuditLogPath: filepath.Join(tmpDir, "audit.log"),
		CreatedAt:    time.Now().UTC(),
		Version:      1,
	}

	if err := SaveMetadata(meta, vaultPath); err != nil {
		t.Fatalf("SaveMetadata failed: %v", err)
	}

	// Verify file exists
	metaPath := MetadataPath(vaultPath)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata file was not created")
	}

	// Test DeleteMetadata
	if err := DeleteMetadata(vaultPath); err != nil {
		t.Fatalf("DeleteMetadata failed: %v", err)
	}

	// Verify file deleted
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Error("Metadata file was not deleted")
	}
}

func TestDeleteMetadata_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "nonexistent.enc")

	// Test DeleteMetadata on non-existent file (should be idempotent)
	if err := DeleteMetadata(vaultPath); err != nil {
		t.Errorf("DeleteMetadata should be idempotent, got error: %v", err)
	}
}

// T004: Unit tests for MetadataPath
func TestMetadataPath(t *testing.T) {
	tests := []struct {
		name      string
		vaultPath string
		expected  string
	}{
		{
			name:      "Unix path",
			vaultPath: "/home/user/.pass-cli/vault.enc",
			expected:  "/home/user/.pass-cli/vault.meta",
		},
		{
			name:      "Windows path",
			vaultPath: "C:\\Users\\user\\.pass-cli\\vault.enc",
			expected:  "C:\\Users\\user\\.pass-cli\\vault.meta",
		},
		{
			name:      "Relative path",
			vaultPath: "vault.enc",
			expected:  "vault.meta",
		},
		{
			name:      "Nested directory",
			vaultPath: "/path/to/nested/vault.enc",
			expected:  "/path/to/nested/vault.meta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MetadataPath(tt.vaultPath)
			// Normalize paths for comparison
			expectedNorm := filepath.FromSlash(tt.expected)
			resultNorm := filepath.FromSlash(result)

			if resultNorm != expectedNorm {
				t.Errorf("Expected %q, got %q", expectedNorm, resultNorm)
			}
		})
	}
}

// T004a: Benchmark tests for SC-003 validation (<50ms metadata operations)
func BenchmarkLoadMetadata(b *testing.B) {
	tmpDir := b.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// Create test metadata
	meta := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: true,
		AuditLogPath: filepath.Join(tmpDir, "audit.log"),
		CreatedAt:    time.Now().UTC(),
		Version:      1,
	}

	if err := SaveMetadata(meta, vaultPath); err != nil {
		b.Fatalf("Failed to create test metadata: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadMetadata(vaultPath)
		if err != nil {
			b.Fatalf("LoadMetadata failed: %v", err)
		}
	}
}

func BenchmarkSaveMetadata(b *testing.B) {
	tmpDir := b.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	meta := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: true,
		AuditLogPath: filepath.Join(tmpDir, "audit.log"),
		CreatedAt:    time.Now().UTC(),
		Version:      1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := SaveMetadata(meta, vaultPath)
		if err != nil {
			b.Fatalf("SaveMetadata failed: %v", err)
		}
	}
}

func BenchmarkDeleteMetadata(b *testing.B) {
	tmpDir := b.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// Create test metadata
	meta := &VaultMetadata{
		VaultID:      vaultPath,
		AuditEnabled: true,
		AuditLogPath: filepath.Join(tmpDir, "audit.log"),
		CreatedAt:    time.Now().UTC(),
		Version:      1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create metadata file before each deletion
		if err := SaveMetadata(meta, vaultPath); err != nil {
			b.Fatalf("Failed to create test metadata: %v", err)
		}
		b.StartTimer()

		err := DeleteMetadata(vaultPath)
		if err != nil {
			b.Fatalf("DeleteMetadata failed: %v", err)
		}
	}
}
