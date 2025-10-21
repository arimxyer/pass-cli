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

// T049: TestFirstRun_InteractiveFlow - Verify first-run detection triggers
// NOTE: Full interactive TTY testing requires manual testing; exec.Command cannot simulate real TTY
// This test verifies that first-run detection logic triggers correctly
func TestFirstRun_InteractiveFlow(t *testing.T) {
	// Setup temporary environment
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	// Set HOME to tmpDir so default vault path is in test directory
	homeDir := tmpDir + "/.pass-cli"
	if err := os.MkdirAll(homeDir, 0700); err != nil {
		t.Fatalf("Failed to create .pass-cli dir: %v", err)
	}

	// Create minimal config without vault
	configContent := []byte("clipboard_timeout: 30\naudit_enabled: false\n")
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	cmd := exec.Command(binaryPath, "--config", configPath, "list")
	cmd.Env = append(os.Environ(), "HOME="+tmpDir, "USERPROFILE="+tmpDir)
	cmd.Stdin = strings.NewReader("")
	output, err := cmd.CombinedOutput()

	// Should fail with non-TTY error (first-run detection triggered)
	if err == nil {
		t.Error("Expected error when no vault exists")
	}

	outputStr := string(output)

	// Verify first-run detection triggered (shows non-TTY message OR init instructions)
	if !strings.Contains(outputStr, "pass-cli init") {
		t.Errorf("Expected first-run detection message, got: %s", outputStr)
	}

	// The detailed interactive flow is tested in unit tests (TestRunGuidedInit_Success)
	// Real TTY testing requires manual validation per quickstart.md
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
