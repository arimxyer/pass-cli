package test

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestVaultConfig creates a temporary config file with a custom vault_path
// and returns the config file path and cleanup function.
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
