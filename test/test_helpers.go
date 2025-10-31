package test

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestVaultConfig creates a temporary config file with a custom vault_path
// and returns the config file path and cleanup function.
// This replaces the need for --vault flag in tests.
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
	os.Setenv("PASS_CLI_CONFIG", configPath)

	// Return cleanup function to restore environment
	cleanup = func() {
		if oldConfig != "" {
			os.Setenv("PASS_CLI_CONFIG", oldConfig)
		} else {
			os.Unsetenv("PASS_CLI_CONFIG")
		}
	}

	return configPath, cleanup
}

// setupTestVault creates a temporary vault file and config pointing to it.
// Returns the vault path, config path, and cleanup function.
func setupTestVault(t *testing.T) (vaultPath string, configPath string, cleanup func()) {
	t.Helper()

	// Create temporary vault file
	tempDir := t.TempDir()
	vaultPath = filepath.Join(tempDir, "test-vault.enc")

	// Create config pointing to this vault
	configPath, configCleanup := setupTestVaultConfig(t, vaultPath)

	cleanup = func() {
		configCleanup()
	}

	return vaultPath, configPath, cleanup
}
