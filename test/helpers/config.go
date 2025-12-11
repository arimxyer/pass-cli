//go:build integration

package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

const (
	// AuditKeyService matches the service name used in internal/security/audit.go
	AuditKeyService = "pass-cli-audit"

	// KeychainService matches the service name used in internal/keychain
	KeychainService = "pass-cli"
)

// SetupTestVaultConfig creates a temporary config file with a custom vault_path
// and returns the config file path and cleanup function.
func SetupTestVaultConfig(t *testing.T, vaultPath string) (configPath string, cleanup func()) {
	t.Helper()

	// Create temporary directory for config
	tempDir := t.TempDir()
	configPath = filepath.Join(tempDir, "config.yml")

	// Write config file with vault_path
	configContent := "vault_path: " + vaultPath + "\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Set environment variable to point to test config
	oldConfig := os.Getenv("PASS_CLI_CONFIG")
	if err := os.Setenv("PASS_CLI_CONFIG", configPath); err != nil {
		t.Fatalf("Failed to set PASS_CLI_CONFIG: %v", err)
	}

	// Return cleanup function to restore environment
	cleanup = func() {
		if oldConfig != "" {
			_ = os.Setenv("PASS_CLI_CONFIG", oldConfig) // Best effort cleanup
		} else {
			_ = os.Unsetenv("PASS_CLI_CONFIG") // Best effort cleanup
		}
	}

	return configPath, cleanup
}

// CleanupVaultDir removes a vault directory and its associated keychain entries.
// This should be used in defer statements to ensure keychain entries don't leak.
// Walks the directory to find all vault.enc files and cleans up ALL keychain entries:
// - Master password entries (vault-specific and legacy global)
// - Audit HMAC key entries
func CleanupVaultDir(t *testing.T, vaultDir string) {
	t.Helper()

	// Find all vault files (.enc) in the directory and delete their keychain entries
	_ = filepath.Walk(vaultDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if there are errors
		}
		if info != nil && !info.IsDir() && strings.HasSuffix(filepath.Base(path), ".enc") && !strings.Contains(filepath.Base(path), ".backup") {
			vaultID := filepath.Base(filepath.Dir(path))

			// Clean master password keychain entries
			vaultSpecificAccount := "master-password-" + vaultID
			_ = keyring.Delete(KeychainService, vaultSpecificAccount)

			// Clean audit key entries (both formats)
			_ = keyring.Delete(AuditKeyService, path)
			_ = keyring.Delete(AuditKeyService, vaultID)
		}
		return nil
	})

	// Also try to delete by the top-level directory name (for single-vault dirs)
	topLevelID := filepath.Base(vaultDir)
	_ = keyring.Delete(AuditKeyService, topLevelID)
	_ = keyring.Delete(KeychainService, "master-password-"+topLevelID)

	// Clean legacy global entry (in case any test used it)
	_ = keyring.Delete(KeychainService, "master-password")

	// Remove the directory
	_ = os.RemoveAll(vaultDir)
}

// CleanupVaultPath removes a specific vault's keychain entries and its parent directory.
// Use this when you know the exact vault.enc path.
// Cleans ALL keychain entries: master password (vault-specific + legacy) and audit keys.
func CleanupVaultPath(t *testing.T, vaultPath string) {
	t.Helper()

	vaultID := filepath.Base(filepath.Dir(vaultPath))

	// Clean master password keychain entries
	vaultSpecificAccount := "master-password-" + vaultID
	_ = keyring.Delete(KeychainService, vaultSpecificAccount)
	_ = keyring.Delete(KeychainService, "master-password") // legacy global

	// Clean audit key entries (both formats)
	_ = keyring.Delete(AuditKeyService, vaultPath)
	_ = keyring.Delete(AuditKeyService, vaultID)

	// Remove the parent directory (which contains the vault)
	_ = os.RemoveAll(filepath.Dir(vaultPath))
}

// CleanupKeychain removes keychain entries for a vault.
// Call this to clean up keychain entries without removing files.
// Cleans both vault-specific entries (new format) and legacy global entry.
func CleanupKeychain(t *testing.T, vaultPath string) {
	t.Helper()

	// Derive vault ID from path
	vaultID := filepath.Base(filepath.Dir(vaultPath))

	// Clean up vault-specific keychain entry (new format: master-password-<vaultID>)
	vaultSpecificAccount := "master-password-" + vaultID
	_ = keyring.Delete(KeychainService, vaultSpecificAccount)

	// Clean up legacy global keychain entry
	_ = keyring.Delete(KeychainService, "master-password")

	// Clean up audit keychain entries
	_ = keyring.Delete(AuditKeyService, vaultPath)
	_ = keyring.Delete(AuditKeyService, vaultID)
}

// CreateTempVaultDir creates a temporary directory for a test vault.
// Returns the vault path (dir/vault.enc) and a cleanup function.
//
// Deprecated: Use SetupTestVault() instead which handles all cleanup automatically
// including keychain entries.
func CreateTempVaultDir(t *testing.T) (vaultPath string, cleanup func()) {
	t.Helper()

	tempDir := t.TempDir()
	vaultPath = filepath.Join(tempDir, "vault.enc")

	cleanup = func() {
		CleanupVaultPath(t, vaultPath)
	}

	return vaultPath, cleanup
}

// SetupTestVault creates a temporary vault directory with automatic cleanup.
// Uses t.Cleanup() to ensure ALL resources are cleaned up:
//   - Master password keychain entry (vault-specific format)
//   - Audit HMAC key keychain entries (both path and vaultID formats)
//   - Temporary files and directories
//
// Usage:
//
//	vaultPath := helpers.SetupTestVault(t)
//	// Use vaultPath in tests - cleanup happens automatically
func SetupTestVault(t *testing.T) string {
	t.Helper()

	// Create temp directory with a named subdirectory for predictable vaultID
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "test-vault")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "vault.enc")
	vaultID := filepath.Base(vaultDir)

	// Register cleanup that handles EVERYTHING
	t.Cleanup(func() {
		// Clean master password keychain entry (vault-specific format)
		vaultSpecificAccount := "master-password-" + vaultID
		_ = keyring.Delete(KeychainService, vaultSpecificAccount)

		// Clean legacy global entry (in case test used it)
		_ = keyring.Delete(KeychainService, "master-password")

		// Clean audit keychain entries (both formats)
		_ = keyring.Delete(AuditKeyService, vaultPath)
		_ = keyring.Delete(AuditKeyService, vaultID)

		// Files are cleaned up by t.TempDir() automatically
	})

	return vaultPath
}

// SetupTestVaultWithName creates a vault with a specific directory name.
// Useful when you need a specific vaultID for testing.
// Cleanup is handled automatically via t.Cleanup().
// Note: vaultDirName can include path separators (e.g., "parent/child/vault-name")
func SetupTestVaultWithName(t *testing.T, vaultDirName string) string {
	t.Helper()

	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, vaultDirName)
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "vault.enc")
	// vaultID is the actual directory name containing vault.enc (last component of path)
	vaultID := filepath.Base(vaultDir)

	t.Cleanup(func() {
		// Clean master password keychain entry
		vaultSpecificAccount := "master-password-" + vaultID
		_ = keyring.Delete(KeychainService, vaultSpecificAccount)
		_ = keyring.Delete(KeychainService, "master-password")

		// Clean audit keychain entries
		_ = keyring.Delete(AuditKeyService, vaultPath)
		_ = keyring.Delete(AuditKeyService, vaultID)
	})

	return vaultPath
}
