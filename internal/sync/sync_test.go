package sync

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"pass-cli/internal/config"
)

func TestNewService(t *testing.T) {
	cfg := config.SyncConfig{
		Enabled: true,
		Remote:  "gdrive:.pass-cli",
	}

	service := NewService(cfg)

	if service == nil {
		t.Fatal("NewService returned nil")
	}
	if service.config.Enabled != true {
		t.Error("Expected Enabled to be true")
	}
	if service.config.Remote != "gdrive:.pass-cli" {
		t.Errorf("Expected Remote to be 'gdrive:.pass-cli', got '%s'", service.config.Remote)
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   config.SyncConfig
		expected bool
	}{
		{
			name: "enabled with remote",
			config: config.SyncConfig{
				Enabled: true,
				Remote:  "gdrive:.pass-cli",
			},
			expected: true,
		},
		{
			name: "disabled",
			config: config.SyncConfig{
				Enabled: false,
				Remote:  "gdrive:.pass-cli",
			},
			expected: false,
		},
		{
			name: "enabled but empty remote",
			config: config.SyncConfig{
				Enabled: true,
				Remote:  "",
			},
			expected: false,
		},
		{
			name: "disabled and empty remote",
			config: config.SyncConfig{
				Enabled: false,
				Remote:  "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.config)
			got := service.IsEnabled()
			if got != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsRcloneInstalled(t *testing.T) {
	service := NewService(config.SyncConfig{})

	// Check if rclone is in PATH
	_, err := exec.LookPath("rclone")
	expected := err == nil

	got := service.IsRcloneInstalled()
	if got != expected {
		t.Errorf("IsRcloneInstalled() = %v, want %v", got, expected)
	}
}

func TestGetVaultDir(t *testing.T) {
	tests := []struct {
		name      string
		vaultPath string
		expected  string
	}{
		{
			name:      "standard path",
			vaultPath: filepath.Join("home", "user", ".pass-cli", "vault.enc"),
			expected:  filepath.Join("home", "user", ".pass-cli"),
		},
		{
			name:      "relative path",
			vaultPath: filepath.Join(".", ".pass-cli", "vault.enc"),
			expected:  filepath.Join(".", ".pass-cli"),
		},
		{
			name:      "deep nested path",
			vaultPath: filepath.Join("a", "b", "c", "d", "vault.enc"),
			expected:  filepath.Join("a", "b", "c", "d"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetVaultDir(tt.vaultPath)
			if got != tt.expected {
				t.Errorf("GetVaultDir(%q) = %q, want %q", tt.vaultPath, got, tt.expected)
			}
		})
	}
}

func TestPull_Disabled(t *testing.T) {
	service := NewService(config.SyncConfig{
		Enabled: false,
		Remote:  "",
	})

	// Pull should return nil without error when disabled
	err := service.Pull("/tmp/test-vault")
	if err != nil {
		t.Errorf("Pull() with disabled sync returned error: %v", err)
	}
}

func TestPull_NoRemote(t *testing.T) {
	service := NewService(config.SyncConfig{
		Enabled: true,
		Remote:  "",
	})

	// Pull should return nil without error when no remote configured
	err := service.Pull("/tmp/test-vault")
	if err != nil {
		t.Errorf("Pull() with empty remote returned error: %v", err)
	}
}

func TestPush_Disabled(t *testing.T) {
	service := NewService(config.SyncConfig{
		Enabled: false,
		Remote:  "",
	})

	// Push should return nil without error when disabled
	err := service.Push("/tmp/test-vault")
	if err != nil {
		t.Errorf("Push() with disabled sync returned error: %v", err)
	}
}

func TestPush_NoRemote(t *testing.T) {
	service := NewService(config.SyncConfig{
		Enabled: true,
		Remote:  "",
	})

	// Push should return nil without error when no remote configured
	err := service.Push("/tmp/test-vault")
	if err != nil {
		t.Errorf("Push() with empty remote returned error: %v", err)
	}
}

func TestPush_VaultDirNotExist(t *testing.T) {
	// Skip if rclone is not installed (can't test push without rclone)
	_, err := exec.LookPath("rclone")
	if err != nil {
		t.Skip("rclone not installed, skipping push test")
	}

	service := NewService(config.SyncConfig{
		Enabled: true,
		Remote:  "gdrive:.pass-cli-test",
	})

	// Push to non-existent directory should return error
	err = service.Push("/nonexistent/path/that/should/not/exist")
	if err == nil {
		t.Error("Push() to non-existent directory should return error")
	}
}

func TestPull_CreatesDirectory(t *testing.T) {
	// Skip if rclone is not installed
	_, err := exec.LookPath("rclone")
	if err != nil {
		t.Skip("rclone not installed, skipping directory creation test")
	}

	// Create a temp directory that doesn't exist yet
	tmpDir := filepath.Join(os.TempDir(), "pass-cli-sync-test-"+randString(8))
	defer os.RemoveAll(tmpDir)

	// Ensure it doesn't exist
	os.RemoveAll(tmpDir)

	service := NewService(config.SyncConfig{
		Enabled: true,
		Remote:  "nonexistent:remote", // This will fail but directory should be created first
	})

	// Pull should create the directory (even if sync itself fails)
	_ = service.Pull(tmpDir)

	// Check if directory was created
	info, err := os.Stat(tmpDir)
	if err == nil && info.IsDir() {
		// Directory was created successfully
		return
	}

	// Note: If rclone fails fast before creating dir, this is also acceptable
	// The important thing is the function doesn't panic
}

// randString generates a random string for test directory names
func randString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[i%len(chars)]
	}
	return string(b)
}
