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

// T029: Integration test for remove command
// Tests: creates vault with keychain, removes, verifies 95% success rate across multiple runs
func TestIntegration_VaultRemove(t *testing.T) {
	// Check if keychain is available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping vault remove integration test")
	}

	testPassword := "RemoveTest-Pass@123"
	vaultPath := filepath.Join(testDir, "remove-test-vault", "vault.enc")

	// Ensure clean state
	defer cleanupKeychain(t, ks)
	defer func() { _ = os.RemoveAll(filepath.Dir(vaultPath)) }() // Best effort cleanup

	// Step 1: Initialize vault WITH keychain
	t.Run("1_Init_With_Keychain", func(t *testing.T) {
		// Setup config with vault_path
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		input := testPassword + "\n" + testPassword + "\n"
		cmd := exec.Command(binaryPath, "init", "--use-keychain")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		cmd.Stdin = strings.NewReader(input)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Fatalf("Init with keychain failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		}

		// Verify vault file was created
		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			t.Error("Vault file was not created")
		}

		// Verify password is in keychain
		_, err = ks.Retrieve()
		if err != nil {
			t.Fatalf("Password was not stored in keychain: %v", err)
		}
	})

	// Step 2: Remove vault with confirmation
	t.Run("2_Remove_With_Confirmation", func(t *testing.T) {
		// T022: Unskipped - This test will FAIL until cmd/vault_remove.go is implemented
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		input := "yes\n" // Confirm removal
		cmd := exec.Command(binaryPath, "vault", "remove")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		cmd.Stdin = strings.NewReader(input)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Fatalf("Vault remove failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		}

		// Verify vault file was deleted
		if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
			t.Error("Vault file should have been deleted")
		}

		// Verify keychain entry was deleted
		_, err = ks.Retrieve()
		if err == nil {
			t.Error("Keychain entry should have been deleted")
		}

		// Verify metadata file was deleted
		metaPath := vaultPath + ".meta.json"
		if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
			t.Error("Metadata file should have been deleted")
		}
	})

	// Step 3: Test removal with --yes flag (no prompt)
	t.Run("3_Remove_With_Yes_Flag", func(t *testing.T) {
		// T024: Unskipped - Test --yes flag
		// Recreate vault for this test
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()
		input := testPassword + "\n" + testPassword + "\n"
		cmd := exec.Command(binaryPath, "init", "--use-keychain")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		cmd.Stdin = strings.NewReader(input)
		_ = cmd.Run() // Best effort setup

		// Remove with --yes flag (no prompt)
		cmd = exec.Command(binaryPath, "vault", "remove", "--yes")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Fatalf("Vault remove --yes failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		}

		// Verify vault file deleted
		if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
			t.Error("Vault file should have been deleted")
		}

		// Verify keychain entry deleted
		_, err = ks.Retrieve()
		if err == nil {
			t.Error("Keychain entry should have been deleted")
		}

		// Verify metadata file deleted
		metaPath := vaultPath + ".meta.json"
		if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
			t.Error("Metadata file should have been deleted")
		}
	})

	// Step 4: Test removal when vault file missing but keychain exists (FR-012)
	t.Run("4_Remove_Orphaned_Keychain", func(t *testing.T) {
		// T025: Unskipped - Test orphaned keychain cleanup
		// Recreate vault
		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()
		input := testPassword + "\n" + testPassword + "\n"
		cmd := exec.Command(binaryPath, "init", "--use-keychain")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		cmd.Stdin = strings.NewReader(input)
		_ = cmd.Run() // Best effort setup

		// Manually delete vault file (simulate orphaned keychain)
		_ = os.Remove(vaultPath) // Best effort cleanup to simulate orphaned keychain

		// Remove should still clean up keychain
		cmd = exec.Command(binaryPath, "vault", "remove", "--yes")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			t.Fatalf("Remove orphaned keychain failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		}

		// Verify keychain entry was cleaned up
		_, err = ks.Retrieve()
		if err == nil {
			t.Error("Orphaned keychain entry should have been deleted (FR-012)")
		}

		// Verify warning message about missing file
		output := stdout.String() + stderr.String()
		if !strings.Contains(output, "not found") && !strings.Contains(output, "missing") && !strings.Contains(output, "Warning") {
			t.Error("Expected warning about missing vault file")
		}
	})

	// Step 5: Test 95% success rate (SC-003)
	t.Run("5_Success_Rate_Test", func(t *testing.T) {
		// T026: Unskipped - Test 95% success rate requirement
		successCount := 0
		totalRuns := 20

		testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		for i := 0; i < totalRuns; i++ {
			// Create vault
			input := testPassword + "\n" + testPassword + "\n"
			cmd := exec.Command(binaryPath, "init", "--use-keychain")
			cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
			cmd.Stdin = strings.NewReader(input)
			_ = cmd.Run() // Best effort setup

			// Remove vault
			cmd = exec.Command(binaryPath, "vault", "remove", "--yes")
			cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
			err := cmd.Run()

			// Check both file and keychain deleted
			fileDeleted := false
			if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
				fileDeleted = true
			}

			keychainDeleted := false
			if _, err := ks.Retrieve(); err != nil {
				keychainDeleted = true
			}

			if err == nil && fileDeleted && keychainDeleted {
				successCount++
			}

			// Cleanup for next run
			_ = os.Remove(vaultPath) // Best effort cleanup
			metaPath := vaultPath + ".meta.json"
			_ = os.Remove(metaPath) // Best effort cleanup
			_ = ks.Delete() // Best effort cleanup
		}

		successRate := float64(successCount) / float64(totalRuns) * 100
		t.Logf("Success rate: %.1f%% (%d/%d)", successRate, successCount, totalRuns)
		if successRate < 95.0 {
			t.Errorf("Success rate %.1f%% is below 95%% requirement (SC-003)", successRate)
		}
	})
}

// T013: Integration test for vault remove with metadata
// Tests that vault remove command writes audit entries (attempt + success) when vault has metadata
func TestIntegration_VaultRemoveWithMetadata(t *testing.T) {
	// Check if keychain is available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	testPassword := "RemoveTest-Pass@123"
	vaultDir := filepath.Join(testDir, "remove-metadata-vault")
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
	input := testPassword + "\n" + testPassword + "\n"
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

	// Get initial audit log content
	initialAuditData, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}
	initialLines := len(strings.Split(string(initialAuditData), "\n"))

	// Run vault remove command with --yes flag (uses vault_path from config)
	cmd = exec.Command(binaryPath, "vault", "remove", "--yes")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Vault remove failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify vault file deleted
	if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
		t.Error("Vault file was not deleted")
	}

	// Verify metadata file deleted
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Error("Metadata file was not deleted after vault removal")
	}

	// Verify audit entries written (attempt + success)
	time.Sleep(100 * time.Millisecond) // Allow audit flush
	finalAuditData, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read audit log after removal: %v", err)
	}

	auditLines := strings.Split(string(finalAuditData), "\n")
	newLines := len(auditLines) - initialLines

	if newLines < 2 {
		t.Fatalf("Expected at least 2 audit entries (attempt + success), got %d new lines", newLines)
	}

	// Parse and verify audit entries
	foundAttempt := false
	foundSuccess := false

	for i := initialLines - 1; i < len(auditLines); i++ {
		if strings.TrimSpace(auditLines[i]) == "" {
			continue
		}

		var auditEntry map[string]interface{}
		if err := json.Unmarshal([]byte(auditLines[i]), &auditEntry); err != nil {
			continue // Skip non-JSON lines
		}

		eventType, ok := auditEntry["event_type"].(string)
		if !ok {
			continue
		}

		outcome, _ := auditEntry["outcome"].(string)

		// Check for vault_remove_attempt or vault_remove with "attempt" outcome
		if eventType == "vault_remove_attempt" || (eventType == "vault_remove" && outcome == "attempt") {
			foundAttempt = true
			t.Logf("✓ Found attempt entry: event_type=%s, outcome=%s", eventType, outcome)
		}

		// Check for vault_remove with "success" outcome
		if eventType == "vault_remove" && outcome == "success" {
			foundSuccess = true
			t.Logf("✓ Found success entry: event_type=%s, outcome=%s", eventType, outcome)
		}
	}

	if !foundAttempt {
		t.Error("No vault_remove_attempt audit entry found")
	}

	if !foundSuccess {
		t.Error("No vault_remove success audit entry found")
	}
}
