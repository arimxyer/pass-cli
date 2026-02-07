//go:build integration

package integration

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/arimxyer/pass-cli/test/helpers"
)

// TestSyncEnable tests the sync enable command
func TestSyncEnable(t *testing.T) {
	testPassword := "SyncTest-Pass@123"

	t.Run("Error_No_Vault", func(t *testing.T) {
		// Create config pointing to non-existent vault
		tempDir := t.TempDir()
		nonExistentVault := tempDir + "/nonexistent/vault.enc"
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, nonExistentVault)
		defer cleanup()

		// Run sync enable without initializing vault
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, "", "sync", "enable")

		// Should fail because vault doesn't exist
		if err == nil {
			t.Error("Expected error when vault doesn't exist")
		}

		allOutput := stdout + stderr
		if !strings.Contains(allOutput, "vault not found") && !strings.Contains(allOutput, "not found") {
			t.Errorf("Expected 'vault not found' error, got: %s", allOutput)
		}
	})

	t.Run("Error_Rclone_Not_Installed", func(t *testing.T) {
		// Skip if rclone is actually installed
		if _, err := exec.LookPath("rclone"); err == nil {
			t.Skip("rclone is installed - skipping 'not installed' test")
		}

		// Setup vault
		vaultPath := helpers.SetupTestVaultWithName(t, "sync-enable-vault")
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault first (with --no-sync to skip sync prompts)
		initInput := helpers.BuildInitStdin(helpers.InitOptions{
			Password:   testPassword,
			NoSync:     true,
			SkipVerify: true,
		})
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, initInput, "init", "--no-sync")
		if err != nil {
			t.Fatalf("Init failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Run sync enable - should fail because rclone not installed
		stdout, stderr, err = helpers.RunCmd(t, binaryPath, testConfigPath, "", "sync", "enable")
		if err == nil {
			t.Error("Expected error when rclone not installed")
		}

		allOutput := stdout + stderr
		if !strings.Contains(allOutput, "rclone") {
			t.Errorf("Expected rclone-related error, got: %s", allOutput)
		}
	})

	t.Run("Shows_Help", func(t *testing.T) {
		tempDir := t.TempDir()
		dummyVault := tempDir + "/dummy/vault.enc"
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, dummyVault)
		defer cleanup()

		stdout, _, err := helpers.RunCmd(t, binaryPath, testConfigPath, "", "sync", "enable", "--help")
		if err != nil {
			t.Fatalf("Help command failed: %v", err)
		}

		// Verify help output contains expected content
		if !strings.Contains(stdout, "Enable cloud sync") {
			t.Errorf("Expected help to mention 'Enable cloud sync', got: %s", stdout)
		}
		if !strings.Contains(stdout, "rclone") {
			t.Errorf("Expected help to mention 'rclone', got: %s", stdout)
		}
		if !strings.Contains(stdout, "--force") {
			t.Errorf("Expected help to mention '--force' flag, got: %s", stdout)
		}
	})

	t.Run("Parent_Command_Shows_Subcommands", func(t *testing.T) {
		tempDir := t.TempDir()
		dummyVault := tempDir + "/dummy/vault.enc"
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, dummyVault)
		defer cleanup()

		stdout, _, err := helpers.RunCmd(t, binaryPath, testConfigPath, "", "sync", "--help")
		if err != nil {
			t.Fatalf("Sync help failed: %v", err)
		}

		// Verify parent command shows enable subcommand
		if !strings.Contains(stdout, "enable") {
			t.Errorf("Expected sync help to show 'enable' subcommand, got: %s", stdout)
		}
		if !strings.Contains(stdout, "cloud sync") || !strings.Contains(stdout, "cloud synchronization") {
			t.Errorf("Expected sync help to describe cloud sync, got: %s", stdout)
		}
	})
}
