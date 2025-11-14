//go:build integration

package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"pass-cli/internal/vault"
)

// T050: Integration test for --no-recovery flag
// Tests: Init with --no-recovery, verify no recovery metadata, attempt recovery returns error

func TestIntegration_NoRecovery(t *testing.T) {
	t.Run("1. Init with --no-recovery flag", func(t *testing.T) {
		// Setup test vault directory
		vaultDir := t.TempDir()
		vaultPath := filepath.Join(vaultDir, "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault WITHOUT recovery
		testPassword := "Test@Password123"

		initCmd := exec.Command(binaryPath, "init", "--no-recovery")
		initCmd.Env = append(os.Environ(), "PASS_CONFIG_PATH="+configPath, "PASS_CLI_TEST=1")

		// Input: master password, confirm password
		initStdin := strings.NewReader(
			testPassword + "\n" + // master password
				testPassword + "\n", // confirm password
		)
		initCmd.Stdin = initStdin

		output, err := initCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Init command failed: %v\nOutput: %s", err, output)
		}

		// Verify no recovery phrase in output
		outputStr := string(output)
		if strings.Contains(outputStr, "Recovery Phrase Setup") {
			t.Error("Recovery phrase should not be displayed with --no-recovery flag")
		}

		if strings.Contains(outputStr, "Verify your backup?") {
			t.Error("Verification prompt should not appear with --no-recovery flag")
		}

		// Verify vault file created
		if _, err := os.Stat(vaultPath); err != nil {
			t.Fatalf("Vault file not created: %v", err)
		}

		// Check metadata
		metadataPath := vault.MetadataPath(vaultPath)
		if _, err := os.Stat(metadataPath); err == nil {
			// Metadata exists, verify recovery is disabled or nil
			metadataBytes, err := os.ReadFile(metadataPath)
			if err != nil {
				t.Fatalf("Failed to read metadata: %v", err)
			}

			var metadata vault.Metadata
			if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
				t.Fatalf("Failed to parse metadata: %v", err)
			}

			// Recovery should be nil or disabled
			if metadata.Recovery != nil && metadata.Recovery.Enabled {
				t.Error("Recovery should not be enabled with --no-recovery flag")
			}
		}

		t.Log("✓ Vault initialized without recovery")
	})

	t.Run("2. Attempt recovery on vault without recovery enabled", func(t *testing.T) {
		// Setup test vault directory
		vaultDir := t.TempDir()
		vaultPath := filepath.Join(vaultDir, "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault WITHOUT recovery
		testPassword := "Test@Password123"

		initCmd := exec.Command(binaryPath, "init", "--no-recovery")
		initCmd.Env = append(os.Environ(), "PASS_CONFIG_PATH="+configPath, "PASS_CLI_TEST=1")
		initStdin := strings.NewReader(testPassword + "\n" + testPassword + "\n")
		initCmd.Stdin = initStdin

		if output, err := initCmd.CombinedOutput(); err != nil {
			t.Fatalf("Init failed: %v\nOutput: %s", err, output)
		}

		// Attempt to use recovery on this vault
		changeCmd := exec.Command(binaryPath, "change-password", "--recover")
		changeCmd.Env = append(os.Environ(), "PASS_CONFIG_PATH="+configPath, "PASS_CLI_TEST=1")

		output, err := changeCmd.CombinedOutput()
		if err == nil {
			t.Error("change-password --recover should fail when recovery not enabled")
		}

		// Verify error message mentions recovery not enabled
		outputStr := string(output)
		if !strings.Contains(outputStr, "recovery not enabled") && !strings.Contains(outputStr, "not available") {
			t.Errorf("Expected 'recovery not enabled' error, got: %s", outputStr)
		}

		t.Log("✓ Recovery attempt correctly rejected")
	})

	t.Run("3. Verify vault still works normally", func(t *testing.T) {
		// Setup test vault directory
		vaultDir := t.TempDir()
		vaultPath := filepath.Join(vaultDir, "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault WITHOUT recovery
		testPassword := "Test@Password123"

		initCmd := exec.Command(binaryPath, "init", "--no-recovery")
		initCmd.Env = append(os.Environ(), "PASS_CONFIG_PATH="+configPath, "PASS_CLI_TEST=1")
		initStdin := strings.NewReader(testPassword + "\n" + testPassword + "\n")
		initCmd.Stdin = initStdin

		if output, err := initCmd.CombinedOutput(); err != nil {
			t.Fatalf("Init failed: %v\nOutput: %s", err, output)
		}

		// Verify normal password change still works (without --recover flag)
		newPassword := "NewTest@Password456"

		changeCmd := exec.Command(binaryPath, "change-password")
		changeCmd.Env = append(os.Environ(), "PASS_CONFIG_PATH="+configPath, "PASS_CLI_TEST=1")
		changeStdin := strings.NewReader(
			testPassword + "\n" + // current password
				newPassword + "\n" + // new password
				newPassword + "\n", // confirm new password
		)
		changeCmd.Stdin = changeStdin

		output, err := changeCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Password change failed: %v\nOutput: %s", err, output)
		}

		// Verify success message
		outputStr := string(output)
		if !strings.Contains(outputStr, "successfully") {
			t.Errorf("Expected success message, got: %s", outputStr)
		}

		t.Log("✓ Normal password change works without recovery")
	})
}
