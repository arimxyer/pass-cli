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

// T034: Integration test for recovery
// Tests: Init with known mnemonic, recover with correct 6 words, verify vault unlocks, verify password changes

func TestIntegration_ChangePasswordWithRecovery(t *testing.T) {
	// This test requires the change-password command with --recover flag
	// which will be implemented in T039-T042
	// For now, we test the internal recovery flow

	t.Run("1. Init vault with recovery, then change password using recovery", func(t *testing.T) {
		// Setup test vault directory
		vaultDir := t.TempDir()
		vaultPath := filepath.Join(vaultDir, "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Step 1: Initialize vault with recovery (skip verification)
		testPassword := "Test@Password123"
		initCmd := exec.Command(binaryPath, "init")
		initCmd.Env = append(os.Environ(), "PASS_CONFIG_PATH="+configPath, "PASS_CLI_TEST=1")
		initStdin := strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n") // password, confirm, no passphrase, skip verification
		initCmd.Stdin = initStdin

		output, err := initCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Init command failed: %v\nOutput: %s", err, output)
		}

		// Verify vault file created
		if _, err := os.Stat(vaultPath); err != nil {
			t.Fatalf("Vault file not created: %v", err)
		}

		// Step 2: Load metadata to get recovery configuration
		metadataPath := vault.MetadataPath(vaultPath)
		metadataBytes, err := os.ReadFile(metadataPath)
		if err != nil {
			t.Fatalf("Failed to read metadata: %v", err)
		}

		var metadata vault.Metadata
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			t.Fatalf("Failed to parse metadata: %v", err)
		}

		// Verify recovery is enabled
		if metadata.Recovery == nil || !metadata.Recovery.Enabled {
			t.Fatal("Recovery not enabled in metadata")
		}

		// Step 3: Extract mnemonic from init output
		// (In real scenario, user would have written down the 24 words)
		// For testing, we parse it from the output
		outputStr := string(output)
		if !strings.Contains(outputStr, "Recovery Phrase Setup") {
			t.Fatal("Recovery phrase not displayed in init output")
		}

		// TODO: When change-password --recover is implemented (T039-T042):
		// - Extract the 24-word mnemonic from output
		// - Extract the 6 challenge positions from metadata
		// - Extract the 6 challenge words
		// - Run: pass-cli change-password --recover
		// - Provide the 6 challenge words
		// - Provide new password
		// - Verify vault can be unlocked with new password
		// - Verify vault cannot be unlocked with old password

		t.Log("✓ Vault initialized with recovery (change-password --recover not yet implemented)")
	})

	t.Run("2. Verify metadata contains valid recovery configuration", func(t *testing.T) {
		// Setup test vault directory
		vaultDir := t.TempDir()
		vaultPath := filepath.Join(vaultDir, "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault with recovery
		testPassword := "Test@Password123"
		initCmd := exec.Command(binaryPath, "init")
		initCmd.Env = append(os.Environ(), "PASS_CONFIG_PATH="+configPath, "PASS_CLI_TEST=1")
		initStdin := strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n") // password, confirm, no passphrase, skip verification
		initCmd.Stdin = initStdin

		output, err := initCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Init command failed: %v\nOutput: %s", err, output)
		}

		// Load and verify metadata
		metadataPath := vault.MetadataPath(vaultPath)
		metadataBytes, err := os.ReadFile(metadataPath)
		if err != nil {
			t.Fatalf("Failed to read metadata: %v", err)
		}

		var metadata vault.Metadata
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			t.Fatalf("Failed to parse metadata: %v", err)
		}

		// Verify recovery metadata
		if metadata.Recovery == nil {
			t.Fatal("Recovery metadata is nil")
		}

		if !metadata.Recovery.Enabled {
			t.Error("Recovery should be enabled")
		}

		// V2 recovery with 6-word challenge
		if metadata.Recovery.Version != "2" {
			t.Errorf("Expected version '2', got '%s'", metadata.Recovery.Version)
		}

		// Verify 6-word challenge data exists
		if len(metadata.Recovery.ChallengePositions) != 6 {
			t.Errorf("Expected 6 challenge positions, got %d", len(metadata.Recovery.ChallengePositions))
		}

		// Verify encrypted stored words (18 words)
		if len(metadata.Recovery.EncryptedStoredWords) == 0 {
			t.Error("EncryptedStoredWords should not be empty")
		}

		if len(metadata.Recovery.NonceStored) != 12 {
			t.Errorf("NonceStored should be 12 bytes, got %d", len(metadata.Recovery.NonceStored))
		}

		// Verify encrypted recovery key (wrapped DEK) exists
		if len(metadata.Recovery.EncryptedRecoveryKey) == 0 {
			t.Error("EncryptedRecoveryKey should not be empty")
		}

		if len(metadata.Recovery.NonceRecovery) != 12 {
			t.Errorf("NonceRecovery should be 12 bytes, got %d", len(metadata.Recovery.NonceRecovery))
		}

		// Verify KDF params
		if metadata.Recovery.KDFParams.Algorithm != "argon2id" {
			t.Errorf("Expected argon2id, got %s", metadata.Recovery.KDFParams.Algorithm)
		}

		if len(metadata.Recovery.KDFParams.SaltChallenge) != 32 {
			t.Errorf("SaltChallenge should be 32 bytes, got %d", len(metadata.Recovery.KDFParams.SaltChallenge))
		}

		if len(metadata.Recovery.KDFParams.SaltRecovery) != 32 {
			t.Errorf("SaltRecovery should be 32 bytes, got %d", len(metadata.Recovery.KDFParams.SaltRecovery))
		}

		t.Log("✓ Metadata contains valid v2 recovery configuration with 6-word challenge")
	})

	t.Run("3. Init with --no-recovery flag skips recovery setup", func(t *testing.T) {
		// Setup test vault directory
		vaultDir := t.TempDir()
		vaultPath := filepath.Join(vaultDir, "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault WITHOUT recovery (--no-recovery flag)
		testPassword := "Test@Password123"
		initCmd := exec.Command(binaryPath, "init", "--no-recovery")
		initCmd.Env = append(os.Environ(), "PASS_CONFIG_PATH="+configPath, "PASS_CLI_TEST=1")
		initStdin := strings.NewReader(testPassword + "\n" + testPassword + "\n")
		initCmd.Stdin = initStdin

		output, err := initCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Init command failed: %v\nOutput: %s", err, output)
		}

		// Verify vault file created
		if _, err := os.Stat(vaultPath); err != nil {
			t.Fatalf("Vault file not created: %v", err)
		}

		// Verify no recovery phrase in output
		outputStr := string(output)
		if strings.Contains(outputStr, "Recovery Phrase Setup") {
			t.Error("Recovery phrase should not be displayed with --no-recovery flag")
		}

		// Verify metadata file doesn't exist OR recovery is not enabled
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

			if metadata.Recovery != nil && metadata.Recovery.Enabled {
				t.Error("Recovery should not be enabled with --no-recovery flag")
			}
		}

		t.Log("✓ --no-recovery flag skips recovery setup")
	})
}
