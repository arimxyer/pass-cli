//go:build integration

package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"pass-cli/internal/keychain"
)

// T009: Integration test for enable command end-to-end
// Tests: Create vault without keychain, run enable, verify subsequent commands don't prompt
func TestIntegration_KeychainEnable(t *testing.T) {
	// Check if keychain is available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping keychain enable integration test")
	}

	testPassword := "EnableTest-Pass@123"
	vaultPath := filepath.Join(testDir, "enable-test-vault", "vault.enc")

	// Ensure clean state
	defer cleanupKeychain(t, ks)
	defer os.RemoveAll(filepath.Dir(vaultPath))

	// Step 1: Initialize vault WITHOUT --use-keychain
	t.Run("1_Init_Without_Keychain", func(t *testing.T) {
		// Setup config with vault_path
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		input := testPassword + "\n" + testPassword + "\n"
		cmd := exec.Command(binaryPath, "init")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		cmd.Stdin = strings.NewReader(input)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Fatalf("Init failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		}

		// Verify vault file was created
		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			t.Error("Vault file was not created")
		}

		// Verify password is NOT in keychain
		_, err = ks.Retrieve()
		if err == nil {
			t.Error("Password should NOT be in keychain after init without --use-keychain")
		}
	})

	// Step 2: Run keychain enable command
	t.Run("2_Enable_Keychain", func(t *testing.T) {
		// T012: Unskipped - This test will FAIL until cmd/keychain_enable.go is implemented
		input := testPassword + "\n"
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()
		cmd := exec.Command(binaryPath, "keychain", "enable")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		cmd.Stdin = strings.NewReader(input)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Fatalf("Keychain enable failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "enabled") {
			t.Errorf("Expected success message containing 'enabled', got: %s", output)
		}

		// Verify password is NOW in keychain
		retrievedPassword, err := ks.Retrieve()
		if err != nil {
			t.Fatalf("Password was not stored in keychain: %v", err)
		}

		if retrievedPassword != testPassword {
			t.Errorf("Keychain password = %q, want %q", retrievedPassword, testPassword)
		}
	})

	// Step 3: Verify subsequent commands don't prompt for password
	t.Run("3_Add_Without_Password_Prompt", func(t *testing.T) {
		// T014: Unskipped - Test idempotent keychain usage
		// Add credential - should NOT prompt for master password (uses keychain)
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()
		input := "testuser\n" + "testpass123\n" // Only username and credential password
		cmd := exec.Command(binaryPath, "add", "github.com")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		cmd.Stdin = strings.NewReader(input)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Fatalf("Add failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		}

		// Should NOT contain "Master password:" prompt
		allOutput := stdout.String() + stderr.String()
		if strings.Contains(allOutput, "Master password:") {
			t.Error("Unexpected master password prompt - keychain should have been used")
		}
	})

	// Step 4: Test --force flag (overwrite existing keychain entry)
	t.Run("4_Enable_With_Force", func(t *testing.T) {
		// T015: Unskipped - Test --force flag
		input := testPassword + "\n"
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()
		cmd := exec.Command(binaryPath, "keychain", "enable", "--force")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		cmd.Stdin = strings.NewReader(input)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Fatalf("Keychain enable --force failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		}

		// Verify password still in keychain
		retrievedPassword, err := ks.Retrieve()
		if err != nil {
			t.Fatalf("Password was not in keychain after --force: %v", err)
		}

		if retrievedPassword != testPassword {
			t.Errorf("Keychain password = %q, want %q", retrievedPassword, testPassword)
		}
	})
}
