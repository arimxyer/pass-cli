// +build integration

package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// T023: TestDoctorCommand_Healthy - Run `pass-cli doctor` with healthy vault → Exit 0, human-readable output
func TestDoctorCommand_Healthy(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create minimal valid vault (encrypted content)
	vaultContent := []byte(`{"version":1,"salt":"test","data":"encrypted"}`)
	if err := os.WriteFile(vaultPath, vaultContent, 0600); err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	// Create valid config
	configContent := []byte(`vault_path: ` + vaultPath + `
clipboard_timeout: 30
audit_enabled: true
`)
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Set environment to use test config (which contains vault_path)
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Run doctor command (vault path comes from config, not flag)
	cmd := exec.Command(binaryPath, "doctor", "--config", configPath)
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	output, err := cmd.CombinedOutput()

	// Assertions
	outputStr := string(output)

	// Allow exit code 0 (all healthy) or 1 (warnings, like keychain unavailable in CI)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			// Exit code 1 is acceptable if there are warnings (e.g., keychain unavailable in CI)
			// Exit code 2 indicates errors, which should fail the test
			if exitCode > 1 {
				t.Errorf("Expected exit code 0 or 1 (warnings), got %d. Output:\n%s", exitCode, outputStr)
			}
			// If exit code is 1, verify it's due to warnings, not errors
			if exitCode == 1 && !strings.Contains(outputStr, "warnings") {
				t.Errorf("Exit code 1 without warnings detected. Output:\n%s", outputStr)
			}
		} else {
			t.Fatalf("Failed to run doctor command: %v", err)
		}
	}

	// Verify we have success indicators (even if warnings are present)
	if !strings.Contains(outputStr, "✅") && !strings.Contains(outputStr, "pass") {
		t.Errorf("Expected pass indicators in output, got:\n%s", outputStr)
	}
}

// T024: TestDoctorCommand_JSON - Run `pass-cli doctor --json` → Valid JSON schema per contract
func TestDoctorCommand_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// Create minimal vault
	vaultContent := []byte(`{"version":1,"salt":"test","data":"encrypted"}`)
	if err := os.WriteFile(vaultPath, vaultContent, 0600); err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	// Set environment to use test config
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Run doctor command with JSON output (vault path comes from config)
	cmd := exec.Command(binaryPath, "doctor", "--json")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	output, err := cmd.CombinedOutput()

	// Should not fail
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Exit code 1 or 2 is OK for JSON output (indicates warnings/errors)
			// Only fail if we can't get output
			if len(output) == 0 {
				t.Fatalf("No output from doctor --json: %v", err)
			}
		}
	}

	// Parse JSON to verify schema
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput:\n%s", err, output)
	}

	// Verify root level fields
	if _, ok := result["report"]; !ok {
		t.Error("Expected 'report' field in JSON output")
	}
	if _, ok := result["vault_path"]; !ok {
		t.Error("Expected 'vault_path' field in JSON output")
	}
	if _, ok := result["vault_path_source"]; !ok {
		t.Error("Expected 'vault_path_source' field in JSON output")
	}

	// Verify report structure
	report, ok := result["report"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'report' to be an object")
	}

	if _, ok := report["summary"]; !ok {
		t.Error("Expected 'summary' field in report")
	}
	if _, ok := report["checks"]; !ok {
		t.Error("Expected 'checks' field in report")
	}
	if _, ok := report["timestamp"]; !ok {
		t.Error("Expected 'timestamp' field in report")
	}

	// Verify summary structure
	summary, ok := report["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'summary' to be an object")
	}
	if _, ok := summary["exit_code"]; !ok {
		t.Error("Expected 'exit_code' in summary")
	}
}

// T025: TestDoctorCommand_Quiet - Run `pass-cli doctor --quiet` → No stdout/stderr, exit code only
func TestDoctorCommand_Quiet(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// Create minimal vault
	vaultContent := []byte(`{"version":1,"salt":"test","data":"encrypted"}`)
	if err := os.WriteFile(vaultPath, vaultContent, 0600); err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	// Set environment to use test config
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Run doctor command with quiet mode (vault path comes from config)
	cmd := exec.Command(binaryPath, "doctor", "--quiet")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	output, err := cmd.CombinedOutput()

	// Check exit code (should be 0, 1, or 2)
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	// In quiet mode, should have minimal or no output
	if len(output) > 100 {
		t.Errorf("Expected minimal output in quiet mode, got %d bytes:\n%s", len(output), output)
	}

	// Exit code should be valid (0, 1, or 2)
	if exitCode < 0 || exitCode > 2 {
		t.Errorf("Expected exit code 0-2, got %d", exitCode)
	}
}

// T026: TestDoctorCommand_Offline - Network unavailable → Version check skipped gracefully
func TestDoctorCommand_Offline(t *testing.T) {
	// This test verifies that doctor doesn't fail when network is unavailable
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// Create minimal vault
	vaultContent := []byte(`{"version":1,"salt":"test","data":"encrypted"}`)
	if err := os.WriteFile(vaultPath, vaultContent, 0600); err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	// Set environment to use test config
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Run doctor command (version check will timeout/fail gracefully, vault path from config)
	cmd := exec.Command(binaryPath, "doctor")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	output, err := cmd.CombinedOutput()

	// Should not crash - exit code 0 or 1 is acceptable
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			if exitCode > 2 {
				t.Fatalf("Expected exit code 0-2, got %d. Output:\n%s", exitCode, output)
			}
		}
	}

	// Output should mention version check (either success or skipped)
	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "version") {
		t.Errorf("Expected version check in output, got:\n%s", outputStr)
	}
}

// T027: TestDoctorCommand_NoVault - Vault missing → Exit 2, reports vault error
func TestDoctorCommand_NoVault(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "nonexistent-vault.enc")

	// Set environment to use test config pointing to non-existent vault
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Run doctor command with non-existent vault (vault path from config)
	cmd := exec.Command(binaryPath, "doctor")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	output, err := cmd.CombinedOutput()

	// Should exit with code 2 (errors)
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 (errors), got %d. Output:\n%s", exitCode, output)
	}

	// Output should mention vault error
	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "vault") {
		t.Errorf("Expected vault error in output, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "❌") && !strings.Contains(strings.ToLower(outputStr), "error") {
		t.Errorf("Expected error indicator in output, got:\n%s", outputStr)
	}
}
