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

// TestIntegration_VerifyAudit tests the audit log verification command.
// This test ensures HMAC signatures are consistent across all code paths:
// - During init (uses getVaultID -> directory name)
// - During New() autodiscovery (must also use directory name)
// - During Unlock() restore (uses stored VaultID)
// - During verify-audit (uses getVaultID -> directory name)
//
// Bug context: Previously vault.New() used full vault path as VaultID but
// init/verify used directory name, causing HMAC verification failures.
//
// Note: Requires system keychain for audit HMAC key storage.
func TestIntegration_VerifyAudit(t *testing.T) {
	// Skip if keychain is not available (audit logging requires keychain for HMAC keys)
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping verify-audit integration test (audit requires keychain for HMAC keys)")
	}

	testPassword := "Verify-Audit-Pass@123"

	// Create a unique vault directory for this test
	vaultDir := filepath.Join(testDir, "verify-audit-test")
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	// Cleanup after test
	defer cleanupVaultPath(t, vaultPath)

	// Setup config for this specific vault
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Helper to run commands with this vault's config
	runCmd := func(input string, args ...string) (string, string, error) {
		cmd := exec.Command(binaryPath, args...)
		if input != "" {
			cmd.Stdin = strings.NewReader(input)
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+configPath)
		err := cmd.Run()
		return stdout.String(), stderr.String(), err
	}

	t.Run("1_Init_Vault_With_Audit", func(t *testing.T) {
		// Initialize vault (audit enabled by default)
		// Input: password, confirm password, no keychain, no passphrase, skip verification
		input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n" + "n\n"
		stdout, stderr, err := runCmd(input, "init")

		if err != nil {
			t.Fatalf("Init failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Verify audit.log was created
		auditLogPath := filepath.Join(vaultDir, "audit.log")
		if _, err := os.Stat(auditLogPath); os.IsNotExist(err) {
			t.Fatal("Audit log was not created during init")
		}
	})

	t.Run("2_Add_Credential_Logs_Audit", func(t *testing.T) {
		// Add a credential - this triggers New() -> Unlock() -> LogAudit()
		input := testPassword + "\n"
		stdout, stderr, err := runCmd(input, "add", "test-service.com", "--username", "testuser", "--password", "testpass123")

		if err != nil {
			t.Fatalf("Add failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "added") && !strings.Contains(stdout, "successfully") {
			t.Errorf("Expected success message, got: %s", stdout)
		}
	})

	t.Run("3_Get_Credential_Logs_Audit", func(t *testing.T) {
		// Get the credential - another operation that logs to audit
		input := testPassword + "\n"
		stdout, stderr, err := runCmd(input, "get", "test-service.com", "--no-clipboard")

		if err != nil {
			t.Fatalf("Get failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "testuser") {
			t.Errorf("Expected credential output, got: %s", stdout)
		}
	})

	t.Run("4_Verify_Audit_All_Entries_Valid", func(t *testing.T) {
		// This is the critical test - verify ALL audit entries have valid HMAC
		// If VaultID is inconsistent, some entries will fail verification
		stdout, stderr, err := runCmd("", "verify-audit")

		if err != nil {
			t.Fatalf("verify-audit failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Check for success message
		if !strings.Contains(stdout, "integrity verified") {
			t.Errorf("Expected 'integrity verified' message, got: %s", stdout)
		}

		// Ensure no invalid entries
		if strings.Contains(stdout, "Invalid entries:") && !strings.Contains(stdout, "Invalid entries: 0") {
			t.Errorf("Found invalid audit entries (HMAC verification failures):\n%s", stdout)
		}

		// Check that there are valid entries (not an empty audit log)
		if strings.Contains(stdout, "Total entries: 0") {
			t.Error("Audit log is empty - expected entries from init, add, and get operations")
		}
	})

	t.Run("5_Multiple_Operations_Then_Verify", func(t *testing.T) {
		// Perform several more operations to stress test VaultID consistency
		operations := []struct {
			input string
			args  []string
		}{
			{testPassword + "\n", []string{"list"}},
			{testPassword + "\n", []string{"get", "test-service.com", "--no-clipboard", "--field", "username"}},
			{testPassword + "\n", []string{"update", "test-service.com", "--password", "newpass456", "--force"}},
			{testPassword + "\n", []string{"get", "test-service.com", "--no-clipboard"}},
		}

		for i, op := range operations {
			_, stderr, err := runCmd(op.input, op.args...)
			if err != nil {
				t.Logf("Operation %d (%v) warning: %v\nStderr: %s", i+1, op.args, err, stderr)
				// Don't fail - some operations might have expected warnings
			}
		}

		// Final verification - all entries must still be valid
		stdout, stderr, err := runCmd("", "verify-audit")

		if err != nil {
			t.Fatalf("Final verify-audit failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "integrity verified") {
			t.Errorf("Final verification failed - expected 'integrity verified', got:\n%s", stdout)
		}

		// Log the final count for debugging
		t.Logf("Final audit verification:\n%s", stdout)
	})
}

// TestIntegration_VerifyAudit_ConsistentVaultID specifically tests that
// the VaultID is consistent between vault.New() autodiscovery and
// the verify-audit command.
//
// Note: Requires system keychain for audit HMAC key storage.
func TestIntegration_VerifyAudit_ConsistentVaultID(t *testing.T) {
	// Skip if keychain is not available (audit logging requires keychain for HMAC keys)
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping verify-audit integration test (audit requires keychain for HMAC keys)")
	}

	testPassword := "Consistent-VaultID@123"

	// Create a unique vault directory
	vaultDir := filepath.Join(testDir, "consistent-vaultid-test")
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	defer cleanupVaultPath(t, vaultPath)

	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	runCmd := func(input string, args ...string) (string, string, error) {
		cmd := exec.Command(binaryPath, args...)
		if input != "" {
			cmd.Stdin = strings.NewReader(input)
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+configPath)
		err := cmd.Run()
		return stdout.String(), stderr.String(), err
	}

	// Initialize vault
	// Input: password, confirm, no keychain, no passphrase, skip verification
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n" + "n\n"
	_, stderr, err := runCmd(input, "init")
	if err != nil {
		t.Fatalf("Init failed: %v\nStderr: %s", err, stderr)
	}

	// Count audit entries from init
	stdout, _, _ := runCmd("", "verify-audit")
	if !strings.Contains(stdout, "Valid entries:") {
		t.Fatalf("Could not parse initial audit count from: %s", stdout)
	}
	t.Logf("After init: %s", stdout)

	// Perform operations that trigger vault.New() -> autodiscovery -> EnableAudit
	// Each operation should use the SAME VaultID as init
	for i := 0; i < 3; i++ {
		input := testPassword + "\n"
		_, stderr, err := runCmd(input, "list")
		if err != nil {
			t.Logf("List %d warning: %v, stderr: %s", i+1, err, stderr)
		}
	}

	// Verify all entries - if VaultID was inconsistent, some will fail
	stdout, stderr, err = runCmd("", "verify-audit")
	if err != nil {
		t.Fatalf("verify-audit failed after list operations: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Parse and verify no invalid entries
	if strings.Contains(stdout, "FAILED") || strings.Contains(stdout, "Invalid entries: ") && !strings.Contains(stdout, "Invalid entries: 0") {
		t.Fatalf("HMAC verification failures detected - VaultID inconsistency:\n%s", stdout)
	}

	t.Logf("Final verification passed:\n%s", stdout)
}
