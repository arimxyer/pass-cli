// +build integration

package test

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Integration tests for first-run guided initialization
// These tests verify the CLI behavior when detecting and handling first-run scenarios

// T049: TestFirstRun_InteractiveFlow - Simulate user input (y, password, keychain, audit) → Vault created
func TestFirstRun_InteractiveFlow(t *testing.T) {
	// Setup temporary environment
	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/vault.enc"
	configPath := tmpDir + "/config.yaml"

	// Create minimal config without vault
	configContent := []byte("clipboard_timeout: 30\naudit_enabled: false\n")
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Simulate user accepting guided init with password input
	input := "y\nTestPassword123!\nTestPassword123!\ny\ny\n"

	cmd := exec.Command(binaryPath, "--vault", vaultPath, "--config", configPath, "list")
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()

	// Should succeed after guided init
	if err != nil {
		t.Logf("Output: %s", output)
		t.Fatalf("Expected guided init to succeed, got error: %v", err)
	}

	// Verify vault was created
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Error("Expected vault to be created during guided init")
	}
}

// T050: TestFirstRun_NonTTY - Piped stdin → Error with manual init instructions
func TestFirstRun_NonTTY(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/vault.enc"
	configPath := tmpDir + "/config.yaml"

	configContent := []byte("clipboard_timeout: 30\naudit_enabled: false\n")
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Run with piped input (non-TTY)
	cmd := exec.Command(binaryPath, "--vault", vaultPath, "--config", configPath, "list")
	cmd.Stdin = bytes.NewReader([]byte{})
	output, err := cmd.CombinedOutput()

	// Should fail with manual init instructions
	if err == nil {
		t.Error("Expected error in non-TTY mode")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "pass-cli init") {
		t.Errorf("Expected manual init instructions in output, got: %s", outputStr)
	}
}

// T051: TestFirstRun_ExistingVault - Vault present → No prompt, command proceeds normally
func TestFirstRun_ExistingVault(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/vault.enc"
	configPath := tmpDir + "/config.yaml"

	// Create vault and config
	vaultContent := []byte(`{"version":1,"salt":"test","data":"encrypted"}`)
	if err := os.WriteFile(vaultPath, vaultContent, 0600); err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	configContent := []byte("clipboard_timeout: 30\naudit_enabled: false\n")
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Run command - should not trigger first-run detection
	cmd := exec.Command(binaryPath, "--vault", vaultPath, "--config", configPath, "list")
	output, err := cmd.CombinedOutput()

	// Output should NOT contain initialization prompts
	outputStr := string(output)
	if strings.Contains(outputStr, "Would you like to create") {
		t.Errorf("Expected no first-run prompt when vault exists, got: %s", outputStr)
	}

	// May fail with authentication error, but not first-run error
	if err != nil && !strings.Contains(outputStr, "password") {
		t.Logf("Output: %s", outputStr)
	}
}

// T052: TestFirstRun_CustomVaultFlag - `pass-cli --vault /tmp/vault list` → No prompt
func TestFirstRun_CustomVaultFlag(t *testing.T) {
	tmpDir := t.TempDir()
	customVaultPath := tmpDir + "/custom-vault.enc"
	configPath := tmpDir + "/config.yaml"

	configContent := []byte("clipboard_timeout: 30\naudit_enabled: false\n")
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Run with custom --vault flag (non-existent vault)
	cmd := exec.Command(binaryPath, "--vault", customVaultPath, "--config", configPath, "list")
	output, err := cmd.CombinedOutput()

	// Should NOT trigger first-run detection (custom vault flag set)
	outputStr := string(output)
	if strings.Contains(outputStr, "Would you like to create") {
		t.Errorf("Expected no first-run prompt with custom --vault flag, got: %s", outputStr)
	}

	// Should get vault not found error instead
	if err == nil || !strings.Contains(outputStr, "vault") {
		t.Logf("Output: %s", outputStr)
	}
}

// T053: TestFirstRun_VersionCommand - `pass-cli version` with no vault → No prompt
func TestFirstRun_VersionCommand(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	configContent := []byte("clipboard_timeout: 30\naudit_enabled: false\n")
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Run version command (doesn't require vault)
	cmd := exec.Command(binaryPath, "--config", configPath, "version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Output: %s", output)
		t.Fatalf("Version command should succeed: %v", err)
	}

	outputStr := string(output)

	// Should NOT trigger first-run detection
	if strings.Contains(outputStr, "Would you like to create") {
		t.Errorf("Expected no first-run prompt for version command, got: %s", outputStr)
	}

	// Should show version info
	if !strings.Contains(outputStr, "version") && !strings.Contains(outputStr, "v") {
		t.Errorf("Expected version output, got: %s", outputStr)
	}
}
