//go:build integration

package test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	defer func() { _ = os.RemoveAll(filepath.Dir(vaultPath)) }() // Best effort cleanup

	// Step 1: Initialize vault WITHOUT keychain
	t.Run("1_Init_Without_Keychain", func(t *testing.T) {
		// Setup config with vault_path
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
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
	})

	// Step 2: Check status BEFORE enabling keychain
	t.Run("2_Status_Before_Enable", func(t *testing.T) {
		// T032: Unskipped - This test will FAIL until cmd/keychain_status.go is implemented
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()
		cmd := exec.Command(binaryPath, "keychain", "status")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Errorf("Status command should not error (exit 0): %v\nStderr: %s", err, stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "Available") {
			t.Errorf("Expected output to contain 'Available', got: %s", output)
		}
		if !strings.Contains(output, "No") && !strings.Contains(output, "not enabled") {
			t.Errorf("Expected output to indicate password not stored, got: %s", output)
		}
		if !strings.Contains(output, "pass-cli keychain enable") {
			t.Errorf("Expected actionable suggestion to enable keychain, got: %s", output)
		}
	})

	// Step 3: Enable keychain
	t.Run("3_Enable_Keychain", func(t *testing.T) {
		// T034: Unskipped - Enable keychain to test status reporting
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
	})

	// Step 4: Check status AFTER enabling keychain
	t.Run("4_Status_After_Enable", func(t *testing.T) {
		// T035: Unskipped - Test status reporting after keychain enabled
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()
		cmd := exec.Command(binaryPath, "keychain", "status")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Errorf("Status command should not error (exit 0): %v\nStderr: %s", err, stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "Available") {
			t.Errorf("Expected output to contain 'Available', got: %s", output)
		}
		if !strings.Contains(output, "Yes") && !strings.Contains(output, "enabled") {
			t.Errorf("Expected output to indicate password is stored, got: %s", output)
		}

		// Verify backend name is displayed (platform-specific)
		hasBackend := strings.Contains(output, "Windows Credential Manager") ||
			strings.Contains(output, "macOS Keychain") ||
			strings.Contains(output, "Linux Secret Service")
		if !hasBackend {
			t.Errorf("Expected output to contain backend name, got: %s", output)
		}
	})
}

// T012: Integration test for keychain status with metadata
// Tests that keychain status command writes audit entry when vault has metadata
// Verifies FR-007 compliance and event_type matches internal/security/audit.go constants
func TestIntegration_KeychainStatusWithMetadata(t *testing.T) {
	// Check if keychain is available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	testPassword := "MetadataTest-Pass@123"
	vaultDir := filepath.Join(testDir, "metadata-status-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	auditLogPath := filepath.Join(vaultDir, "audit.log")

	// Ensure clean state
	defer cleanupKeychain(t, ks)
	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	// Create vault with audit enabled
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault with audit
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init", "--enable-audit")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init with audit failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify metadata file created
	metaPath := vaultPath + ".meta.json"
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata file was not created")
	}

	// Get initial audit log line count
	initialLines := 0
	if _, err := os.Stat(auditLogPath); err == nil {
		data, _ := os.ReadFile(auditLogPath)
		initialLines = len(strings.Split(string(data), "\n")) - 1
	}

	// Run keychain status command (reuse config from init)
	cmd = exec.Command(binaryPath, "keychain", "status")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Keychain status failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify audit entry written (FR-007)
	time.Sleep(100 * time.Millisecond) // Allow audit flush
	data, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	auditLines := strings.Split(string(data), "\n")
	newLines := len(auditLines) - initialLines - 1

	if newLines < 1 {
		t.Fatal("No audit entry written for keychain status command")
	}

	// Verify last audit entry has correct event_type (matches internal/security/audit.go)
	lastLine := auditLines[len(auditLines)-2] // -2 because last is empty
	var auditEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lastLine), &auditEntry); err != nil {
		t.Fatalf("Failed to parse audit entry: %v", err)
	}

	eventType, ok := auditEntry["event_type"].(string)
	if !ok {
		t.Fatal("Audit entry missing event_type field")
	}

	// Verify event type matches constant from internal/security/audit.go
	if eventType != "keychain_status" {
		t.Errorf("Expected event_type 'keychain_status', got %q", eventType)
	}

	outcome, ok := auditEntry["outcome"].(string)
	if !ok || outcome == "" {
		t.Error("Audit entry missing outcome field")
	}

	t.Logf("✓ Audit entry written: event_type=%s, outcome=%s", eventType, outcome)
}
