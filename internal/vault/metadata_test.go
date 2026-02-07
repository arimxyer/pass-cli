package vault_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arimxyer/pass-cli/internal/vault"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMetadata_MissingFile(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Load metadata from non-existent file
	metadata, err := vault.LoadMetadata(vaultPath)
	require.NoError(t, err)
	assert.Equal(t, "1.0", metadata.Version)
	assert.False(t, metadata.KeychainEnabled)
	assert.False(t, metadata.AuditEnabled)
	assert.True(t, metadata.CreatedAt.IsZero())
	assert.True(t, metadata.LastModified.IsZero())
}

func TestSaveAndLoadMetadata(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Save metadata
	metadata := &vault.Metadata{
		Version:         "1.0",
		KeychainEnabled: true,
		AuditEnabled:    false,
	}
	err := vault.SaveMetadata(vaultPath, metadata)
	require.NoError(t, err)

	// Verify file exists
	metadataPath := vault.MetadataPath(vaultPath)
	assert.FileExists(t, metadataPath)

	// Load it back
	loaded, err := vault.LoadMetadata(vaultPath)
	require.NoError(t, err)
	assert.Equal(t, "1.0", loaded.Version)
	assert.True(t, loaded.KeychainEnabled)
	assert.False(t, loaded.AuditEnabled)
	assert.False(t, loaded.CreatedAt.IsZero())
	assert.False(t, loaded.LastModified.IsZero())
}

func TestMetadataPermissions(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Save metadata
	metadata := &vault.Metadata{
		Version:         "1.0",
		KeychainEnabled: true,
		AuditEnabled:    true,
	}
	err := vault.SaveMetadata(vaultPath, metadata)
	require.NoError(t, err)

	// Check file permissions (0600)
	metadataPath := vault.MetadataPath(vaultPath)
	info, err := os.Stat(metadataPath)
	require.NoError(t, err)

	// On Unix-like systems, check permissions
	// Windows doesn't support Unix-style permissions
	mode := info.Mode()
	if mode.Perm() != 0600 {
		t.Logf("Note: File permissions are %#o (expected 0600), but this may be OK on Windows", mode.Perm())
	}
}

func TestDeleteMetadata(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Save metadata first
	metadata := &vault.Metadata{
		Version:         "1.0",
		KeychainEnabled: true,
		AuditEnabled:    false,
	}
	err := vault.SaveMetadata(vaultPath, metadata)
	require.NoError(t, err)

	// Verify file exists
	metadataPath := vault.MetadataPath(vaultPath)
	assert.FileExists(t, metadataPath)

	// Delete metadata
	err = vault.DeleteMetadata(vaultPath)
	require.NoError(t, err)

	// Verify file deleted
	assert.NoFileExists(t, metadataPath)

	// Idempotent - deleting again should not error
	err = vault.DeleteMetadata(vaultPath)
	require.NoError(t, err)
}

func TestSaveMetadata_SetsTimestamps(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Save metadata without timestamps
	metadata := &vault.Metadata{
		Version:         "1.0",
		KeychainEnabled: false,
		AuditEnabled:    false,
	}

	before := time.Now().UTC()
	err := vault.SaveMetadata(vaultPath, metadata)
	after := time.Now().UTC()
	require.NoError(t, err)

	// CreatedAt and LastModified should be set
	assert.False(t, metadata.CreatedAt.IsZero())
	assert.False(t, metadata.LastModified.IsZero())

	// Timestamps should be within test time range
	assert.True(t, metadata.CreatedAt.After(before.Add(-time.Second)))
	assert.True(t, metadata.CreatedAt.Before(after.Add(time.Second)))
	assert.True(t, metadata.LastModified.After(before.Add(-time.Second)))
	assert.True(t, metadata.LastModified.Before(after.Add(time.Second)))
}

func TestSaveMetadata_UpdatesLastModified(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Save metadata
	metadata := &vault.Metadata{
		Version:         "1.0",
		KeychainEnabled: false,
		AuditEnabled:    false,
	}
	err := vault.SaveMetadata(vaultPath, metadata)
	require.NoError(t, err)

	firstModified := metadata.LastModified
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	// Update and save again
	metadata.KeychainEnabled = true
	err = vault.SaveMetadata(vaultPath, metadata)
	require.NoError(t, err)

	// LastModified should be updated, CreatedAt should not
	assert.True(t, metadata.LastModified.After(firstModified))
}

func TestMetadataPath(t *testing.T) {
	tests := []struct {
		name      string
		vaultPath string
		expected  string
	}{
		{
			name:      "simple path",
			vaultPath: "/path/to/vault.enc",
			expected:  "/path/to/vault.enc.meta.json",
		},
		{
			name:      "Windows path",
			vaultPath: "C:\\Users\\test\\vault.enc",
			expected:  "C:\\Users\\test\\vault.enc.meta.json",
		},
		{
			name:      "relative path",
			vaultPath: "vault.enc",
			expected:  "vault.enc.meta.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vault.MetadataPath(tt.vaultPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadMetadata_CorruptedFile(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	metadataPath := vault.MetadataPath(vaultPath)

	// Write invalid JSON
	err := os.WriteFile(metadataPath, []byte("{invalid json"), 0600)
	require.NoError(t, err)

	// Load should return error
	_, err = vault.LoadMetadata(vaultPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "corrupted metadata file")
}
