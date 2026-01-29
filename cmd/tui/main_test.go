package tui

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetDefaultVaultPath_UsesConfigVaultPath verifies that getDefaultVaultPath()
// reads vault_path from config, ensuring TUI and CLI use the same vault file.
// This is the regression test for the bug where TUI always used ~/.pass-cli/vault.enc
// regardless of config, causing credentials added via TUI to be invisible to CLI.
func TestGetDefaultVaultPath_UsesConfigVaultPath(t *testing.T) {
	t.Run("returns config vault_path when set", func(t *testing.T) {
		// Create a temp config file with a custom vault_path
		tmpDir := t.TempDir()
		customVaultPath := filepath.Join(tmpDir, "custom", "vault.enc")

		configPath := filepath.Join(tmpDir, "config.yml")
		configContent := "vault_path: " + customVaultPath + "\n"
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		// Set PASS_CLI_CONFIG so config.Load() finds our test config
		t.Setenv("PASS_CLI_CONFIG", configPath)

		got := getDefaultVaultPath()
		if got != customVaultPath {
			t.Errorf("getDefaultVaultPath() = %q, want %q", got, customVaultPath)
		}
	})

	t.Run("returns default path when no config vault_path", func(t *testing.T) {
		// Create a config file without vault_path
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yml")
		configContent := "theme: dracula\n"
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		t.Setenv("PASS_CLI_CONFIG", configPath)

		got := getDefaultVaultPath()
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".pass-cli", "vault.enc")
		if got != want {
			t.Errorf("getDefaultVaultPath() = %q, want %q", got, want)
		}
	})

	t.Run("expands tilde in config vault_path", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yml")
		configContent := "vault_path: \"~/my-vault/vault.enc\"\n"
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		t.Setenv("PASS_CLI_CONFIG", configPath)

		got := getDefaultVaultPath()
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, "my-vault", "vault.enc")
		if got != want {
			t.Errorf("getDefaultVaultPath() = %q, want %q", got, want)
		}
	})

	t.Run("returns default path when no config file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Point to a non-existent config file
		t.Setenv("PASS_CLI_CONFIG", filepath.Join(tmpDir, "nonexistent.yml"))

		got := getDefaultVaultPath()
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".pass-cli", "vault.enc")
		if got != want {
			t.Errorf("getDefaultVaultPath() = %q, want %q", got, want)
		}
	})
}
