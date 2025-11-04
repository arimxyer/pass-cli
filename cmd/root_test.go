package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// T018: Unit test for custom vault_path resolution
func TestGetVaultPath_CustomPath(t *testing.T) {
	// Create temporary config directory
	tmpDir, err := os.MkdirTemp("", "pass-cli-config-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Set XDG_CONFIG_HOME to temp directory
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", origXDG) }()
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}

	tests := []struct {
		name         string
		configYAML   string
		expectSuffix string
	}{
		{
			name:         "custom absolute path",
			configYAML:   "vault_path: " + getTestAbsolutePath() + "\n",
			expectSuffix: filepath.Base(getTestAbsolutePath()),
		},
		{
			name:         "empty config uses default",
			configYAML:   "",
			expectSuffix: filepath.Join(".pass-cli", "vault.enc"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config file
			configDir := filepath.Join(tmpDir, "pass-cli")
			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatalf("Failed to create config dir: %v", err)
			}
			configPath := filepath.Join(configDir, "config.yml")

			if tt.configYAML != "" {
				if err := os.WriteFile(configPath, []byte(tt.configYAML), 0644); err != nil {
					t.Fatalf("Failed to write config: %v", err)
				}
			} else {
				_ = os.Remove(configPath) // Best effort: ensure no config exists
			}

			result := GetVaultPath()
			if !strings.HasSuffix(result, tt.expectSuffix) {
				t.Errorf("Expected path to end with %s, got: %s", tt.expectSuffix, result)
			}
		})
	}
}

// T019: Unit test for ~ expansion in vault path
func TestGetVaultPath_TildeExpansion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pass-cli-config-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", origXDG) }()
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot determine home directory")
	}

	tests := []struct {
		name         string
		configPath   string
		expectPrefix string
	}{
		{
			name:         "tilde expands to home",
			configPath:   "~/.pass-cli/custom.enc",
			expectPrefix: filepath.Join(home, ".pass-cli"),
		},
		{
			name:         "tilde only",
			configPath:   "~/vault.enc",
			expectPrefix: home,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configDir := filepath.Join(tmpDir, "pass-cli")
			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatalf("Failed to create config dir: %v", err)
			}
			configPath := filepath.Join(configDir, "config.yml")
			
			configYAML := "vault_path: " + tt.configPath + "\n"
			if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			result := GetVaultPath()
			if !strings.HasPrefix(result, tt.expectPrefix) {
				t.Errorf("Expected path to start with %s, got: %s", tt.expectPrefix, result)
			}
		})
	}
}

// T020: Unit test for $HOME / %USERPROFILE% expansion
func TestGetVaultPath_EnvVarExpansion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pass-cli-config-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", origXDG) }()
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot determine home directory")
	}

	var envVar string
	if runtime.GOOS == "windows" {
		envVar = "%USERPROFILE%\\.pass-cli\\vault.enc"
	} else {
		envVar = "$HOME/.pass-cli/vault.enc"
	}

	configDir := filepath.Join(tmpDir, "pass-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yml")

	configYAML := "vault_path: " + envVar + "\n"
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	result := GetVaultPath()
	expected := filepath.Join(home, ".pass-cli", "vault.enc")

	if result != expected {
		t.Errorf("Environment variable expansion failed.\nExpected: %s\nGot: %s", expected, result)
	}
}

// T021: Unit test for relative path → absolute conversion
func TestGetVaultPath_RelativeToAbsolute(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pass-cli-config-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", origXDG) }()
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot determine home directory")
	}

	tests := []struct {
		name         string
		configPath   string
		expectPrefix string
	}{
		{
			name:         "relative path converts to absolute",
			configPath:   "custom/vault.enc",
			expectPrefix: home,
		},
		{
			name:         "single file relative path",
			configPath:   "vault.enc",
			expectPrefix: home,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configDir := filepath.Join(tmpDir, "pass-cli")
			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatalf("Failed to create config dir: %v", err)
			}
			configPath := filepath.Join(configDir, "config.yml")
			
			configYAML := "vault_path: " + tt.configPath + "\n"
			if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			result := GetVaultPath()
			if !filepath.IsAbs(result) {
				t.Errorf("Expected absolute path, got relative: %s", result)
			}
			if !strings.HasPrefix(result, tt.expectPrefix) {
				t.Errorf("Expected path to start with %s, got: %s", tt.expectPrefix, result)
			}
		})
	}
}


// T035: Unit test for --vault flag error handler
func TestVaultFlagError(t *testing.T) {
	// Simulate attempting to use --vault flag
	// The flag should not exist, and attempting to use it should produce an error
	
	// We can't easily test cobra's flag parsing in a unit test without running the actual command,
	// so this test verifies the flag is not registered
	cmd := rootCmd
	
	// Try to lookup --vault flag
	flag := cmd.PersistentFlags().Lookup("vault")
	
	if flag != nil {
		t.Errorf("--vault flag should not be registered, but found: %v", flag)
	}
	
	t.Log("✓ --vault flag is not registered (correctly removed)")
}

// Helper function for cross-platform absolute path
func getTestAbsolutePath() string {
	if runtime.GOOS == "windows" {
		return "C:\\custom\\test\\vault.enc"
	}
	return "/custom/test/vault.enc"
}
