package health

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// T020: TestRunChecks_AllPass - All checks pass → ExitCode=0, summary correct
func TestRunChecks_AllPass(t *testing.T) {
	// Create temporary test environment
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Setup options pointing to test environment
	opts := CheckOptions{
		CurrentVersion: "v1.2.3",
		GitHubRepo:     "test/pass-cli",
		VaultPath:      vaultPath,
		VaultDir:       tmpDir,
		ConfigPath:     configPath,
	}

	// Create minimal valid vault
	vaultContent := []byte(`{"version":1,"salt":"test","data":"encrypted"}`)
	if err := os.WriteFile(vaultPath, vaultContent, 0600); err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	// Create minimal valid config
	configContent := []byte(`vault_path: ` + vaultPath + `
clipboard_timeout: 30
audit_enabled: false
`)
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Execute all checks
	report := RunChecks(context.Background(), opts)

	// Assertions
	if report.Summary.ExitCode != ExitHealthy {
		t.Errorf("Expected exit code %d, got %d", ExitHealthy, report.Summary.ExitCode)
	}
	if report.Summary.Errors > 0 {
		t.Errorf("Expected 0 errors, got %d", report.Summary.Errors)
	}
	if report.Summary.Warnings > 0 {
		t.Errorf("Expected 0 warnings, got %d", report.Summary.Warnings)
	}

	// Should have 5 checks (version, vault, config, keychain, backup)
	expectedChecks := 5
	if len(report.Checks) != expectedChecks {
		t.Errorf("Expected %d checks, got %d", expectedChecks, len(report.Checks))
	}

	// Verify all checks passed
	for _, check := range report.Checks {
		if check.Status != CheckPass {
			t.Errorf("Check %s did not pass: status=%s, message=%s",
				check.Name, check.Status, check.Message)
		}
	}

	// Verify timestamp is recent
	if report.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

// T021: TestRunChecks_WithWarnings - Some warnings → ExitCode=1, summary correct
func TestRunChecks_WithWarnings(t *testing.T) {
	// This will test the scenario where some checks return warnings
	// For example: old backup file, or config value out of range

	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")
	configPath := filepath.Join(tmpDir, "config.yaml")

	opts := CheckOptions{
		CurrentVersion: "v1.2.3",
		GitHubRepo:     "test/pass-cli",
		VaultPath:      vaultPath,
		VaultDir:       tmpDir,
		ConfigPath:     configPath,
	}

	// Create minimal valid vault
	vaultContent := []byte(`{"version":1,"salt":"test","data":"encrypted"}`)
	if err := os.WriteFile(vaultPath, vaultContent, 0600); err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	// Create config with a warning (clipboard_timeout > 300)
	configContent := []byte(`vault_path: ` + vaultPath + `
clipboard_timeout: 500
audit_enabled: false
`)
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Execute all checks
	report := RunChecks(context.Background(), opts)

	// Assertions
	if report.Summary.ExitCode != ExitWarnings {
		t.Errorf("Expected exit code %d, got %d", ExitWarnings, report.Summary.ExitCode)
	}
	if report.Summary.Warnings == 0 {
		t.Error("Expected at least one warning")
	}
	if report.Summary.Errors > 0 {
		t.Errorf("Expected 0 errors (only warnings), got %d", report.Summary.Errors)
	}

	// Verify summary counts match actual check results
	actualWarnings := 0
	actualErrors := 0
	actualPassed := 0
	for _, check := range report.Checks {
		switch check.Status {
		case CheckPass:
			actualPassed++
		case CheckWarning:
			actualWarnings++
		case CheckError:
			actualErrors++
		}
	}

	if actualWarnings != report.Summary.Warnings {
		t.Errorf("Summary warnings (%d) doesn't match actual (%d)",
			report.Summary.Warnings, actualWarnings)
	}
	if actualPassed != report.Summary.Passed {
		t.Errorf("Summary passed (%d) doesn't match actual (%d)",
			report.Summary.Passed, actualPassed)
	}
}

// T022: TestRunChecks_WithErrors - Some errors → ExitCode=2, summary correct
func TestRunChecks_WithErrors(t *testing.T) {
	// This will test the scenario where some checks return errors
	// For example: missing vault file

	tmpDir := t.TempDir()

	opts := CheckOptions{
		CurrentVersion: "v1.2.3",
		GitHubRepo:     "test/pass-cli",
		VaultPath:      tmpDir + "/nonexistent-vault.enc",  // Doesn't exist
		VaultDir:       tmpDir,
		ConfigPath:     tmpDir + "/config.yaml",
	}

	// Execute all checks
	report := RunChecks(context.Background(), opts)

	// Assertions
	if report.Summary.ExitCode != ExitErrors {
		t.Errorf("Expected exit code %d, got %d", ExitErrors, report.Summary.ExitCode)
	}
	if report.Summary.Errors == 0 {
		t.Error("Expected at least one error")
	}

	// Verify summary counts match actual check results
	actualWarnings := 0
	actualErrors := 0
	actualPassed := 0
	for _, check := range report.Checks {
		switch check.Status {
		case CheckPass:
			actualPassed++
		case CheckWarning:
			actualWarnings++
		case CheckError:
			actualErrors++
		}
	}

	if actualErrors != report.Summary.Errors {
		t.Errorf("Summary errors (%d) doesn't match actual (%d)",
			report.Summary.Errors, actualErrors)
	}

	// Exit code prioritizes errors over warnings
	if report.Summary.Errors > 0 && report.Summary.ExitCode != ExitErrors {
		t.Error("Exit code should be ExitErrors when errors are present")
	}
}
