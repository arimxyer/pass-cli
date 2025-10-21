//go:build integration
// +build integration

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

// T020: Integration test for status command
// Tests: creates vault, checks status before/after enable, verifies output format
func TestIntegration_KeychainStatus(t *testing.T) {
	// Check if keychain is available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping keychain status integration test")
	}

	testPassword := "StatusTest-Pass@123"
	vaultPath := filepath.Join(testDir, "status-test-vault", "vault.enc")

	// Ensure clean state
	defer cleanupKeychain(t, ks)
	defer os.RemoveAll(filepath.Dir(vaultPath))

	// Step 1: Initialize vault WITHOUT keychain
	t.Run("1_Init_Without_Keychain", func(t *testing.T) {
		input := testPassword + "\n" + testPassword + "\n"
		cmd := exec.Command(binaryPath, "--vault", vaultPath, "init")
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
	})

	// Step 2: Check status BEFORE enabling keychain
	t.Run("2_Status_Before_Enable", func(t *testing.T) {
		// This test will FAIL until cmd/keychain_status.go is implemented (T021)
		t.Skip("TODO: Implement keychain status command (T021)")

		// TODO T021: After implementation, test should:
		// cmd := exec.Command(binaryPath, "--vault", vaultPath, "keychain", "status")
		//
		// var stdout, stderr bytes.Buffer
		// cmd.Stdout = &stdout
		// cmd.Stderr = &stderr
		//
		// err := cmd.Run()
		// if err != nil {
		//     t.Errorf("Status command should not error (exit 0): %v\nStderr: %s", err, stderr.String())
		// }
		//
		// output := stdout.String()
		// if !strings.Contains(output, "Available") {
		//     t.Errorf("Expected output to contain 'Available', got: %s", output)
		// }
		// if !strings.Contains(output, "No") || !strings.Contains(output, "not enabled") {
		//     t.Errorf("Expected output to indicate password not stored, got: %s", output)
		// }
		// if !strings.Contains(output, "pass-cli keychain enable") {
		//     t.Errorf("Expected actionable suggestion to enable keychain, got: %s", output)
		// }
	})

	// Step 3: Enable keychain
	t.Run("3_Enable_Keychain", func(t *testing.T) {
		t.Skip("TODO: Implement after T011 (depends on enable command)")

		// TODO T011: After enable command implementation:
		// input := testPassword + "\n"
		// cmd := exec.Command(binaryPath, "--vault", vaultPath, "keychain", "enable")
		// cmd.Stdin = strings.NewReader(input)
		//
		// var stdout, stderr bytes.Buffer
		// cmd.Stdout = &stdout
		// cmd.Stderr = &stderr
		//
		// err := cmd.Run()
		// if err != nil {
		//     t.Fatalf("Keychain enable failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		// }
	})

	// Step 4: Check status AFTER enabling keychain
	t.Run("4_Status_After_Enable", func(t *testing.T) {
		t.Skip("TODO: Implement after T011 and T021")

		// TODO T021: After status command implementation:
		// cmd := exec.Command(binaryPath, "--vault", vaultPath, "keychain", "status")
		//
		// var stdout, stderr bytes.Buffer
		// cmd.Stdout = &stdout
		// cmd.Stderr = &stderr
		//
		// err := cmd.Run()
		// if err != nil {
		//     t.Errorf("Status command should not error (exit 0): %v\nStderr: %s", err, stderr.String())
		// }
		//
		// output := stdout.String()
		// if !strings.Contains(output, "Available") {
		//     t.Errorf("Expected output to contain 'Available', got: %s", output)
		// }
		// if !strings.Contains(output, "Yes") || !strings.Contains(output, "enabled") {
		//     t.Errorf("Expected output to indicate password is stored, got: %s", output)
		// }
		//
		// // Verify backend name is displayed (platform-specific)
		// hasBackend := strings.Contains(output, "Windows Credential Manager") ||
		//               strings.Contains(output, "macOS Keychain") ||
		//               strings.Contains(output, "Linux Secret Service")
		// if !hasBackend {
		//     t.Errorf("Expected output to contain backend name, got: %s", output)
		// }
	})
}
