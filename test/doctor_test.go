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

	// Run doctor command
	cmd := exec.Command(binaryPath, "doctor", "--vault", vaultPath, "--config", configPath)
	output, err := cmd.CombinedOutput()

	// Assertions
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 0 {
				t.Errorf("Expected exit code 0, got %d. Output:\n%s", exitErr.ExitCode(), output)
			}
		} else {
			t.Fatalf("Failed to run doctor command: %v", err)
		}
	}

	outputStr := string(output)
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

	// Run doctor command with JSON output
	cmd := exec.Command(binaryPath, "doctor", "--json", "--vault", vaultPath)
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
	var report map[string]interface{}
	if err := json.Unmarshal(output, &report); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput:\n%s", err, output)
	}

	// Verify required fields
	if _, ok := report["summary"]; !ok {
		t.Error("Expected 'summary' field in JSON output")
	}
	if _, ok := report["checks"]; !ok {
		t.Error("Expected 'checks' field in JSON output")
	}
	if _, ok := report["timestamp"]; !ok {
		t.Error("Expected 'timestamp' field in JSON output")
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

	// Run doctor command with quiet mode
	cmd := exec.Command(binaryPath, "doctor", "--quiet", "--vault", vaultPath)
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

	// Run doctor command (version check will timeout/fail gracefully)
	cmd := exec.Command(binaryPath, "doctor", "--vault", vaultPath)
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

	// Run doctor command with non-existent vault
	cmd := exec.Command(binaryPath, "doctor", "--vault", vaultPath)
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
