package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

const (
	// auditKeyService matches the service name used in internal/security/audit.go
	auditKeyService = "pass-cli-audit"
)

// setupTestVaultConfig creates a temporary config file with a custom vault_path
// and returns the config file path and cleanup function.
//
//nolint:unused // Used in integration tests (build tag: integration)
func setupTestVaultConfig(t *testing.T, vaultPath string) (configPath string, cleanup func()) {
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

// cleanupVaultDir removes a vault directory and its associated keychain entries.
// This should be used in defer statements to ensure keychain entries don't leak.
// Walks the directory to find all vault.enc files and cleans up their audit keys.
//
//nolint:unused // Used in integration tests (build tag: integration)
func cleanupVaultDir(t *testing.T, vaultDir string) {
	t.Helper()

	// Find all vault.enc files in the directory and delete their keychain entries
	_ = filepath.Walk(vaultDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if there are errors
		}
		if info != nil && !info.IsDir() && filepath.Base(path) == "vault.enc" {
			// Delete the audit key for this vault path (full path format)
			_ = keyring.Delete(auditKeyService, path)
			// Also delete by vault ID (directory name format - used by pass-cli)
			vaultID := filepath.Base(filepath.Dir(path))
			_ = keyring.Delete(auditKeyService, vaultID)
		}
		return nil
	})

	// Also try to delete by the top-level directory name (for single-vault dirs)
	_ = keyring.Delete(auditKeyService, filepath.Base(vaultDir))

	// Remove the directory
	_ = os.RemoveAll(vaultDir)
}

// cleanupVaultPath removes a specific vault's keychain entry and its parent directory.
// Use this when you know the exact vault.enc path.
//
//nolint:unused // Used in integration tests (build tag: integration)
func cleanupVaultPath(t *testing.T, vaultPath string) {
	t.Helper()

	// Delete the audit key for this vault (full path format)
	_ = keyring.Delete(auditKeyService, vaultPath)
	// Also delete by vault ID (directory name format - used by pass-cli)
	vaultID := filepath.Base(filepath.Dir(vaultPath))
	_ = keyring.Delete(auditKeyService, vaultID)

	// Remove the parent directory (which contains the vault)
	_ = os.RemoveAll(filepath.Dir(vaultPath))
}

// cleanupKeychain removes keychain entries using KeychainService.
// This is a compatibility wrapper for old tests.
// DEPRECATED: Use helpers.CleanupKeychain instead.
//
//nolint:unused // Used in integration tests (build tag: integration)
func cleanupKeychain(t *testing.T, ks interface{}) {
	t.Helper()

	// Use the keychain service to delete all entries
	// This is compatible with the old signature cleanupKeychain(t, ks)
	if ksSvc, ok := ks.(interface{ Clear() error }); ok {
		_ = ksSvc.Clear() // Best effort cleanup
	}
}
