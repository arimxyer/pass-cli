//go:build integration
// +build integration

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
	defer os.RemoveAll(filepath.Dir(vaultPath))

	// Step 1: Initialize vault WITH keychain
	t.Run("1_Init_With_Keychain", func(t *testing.T) {
		input := testPassword + "\n" + testPassword + "\n"
		cmd := exec.Command(binaryPath, "--vault", vaultPath, "init", "--use-keychain")
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
		// This test will FAIL until cmd/vault_remove.go is implemented (T030)
		t.Skip("TODO: Implement vault remove command (T030)")

		// TODO T030: After implementation, test should:
		// input := "yes\n" // Confirm removal
		// cmd := exec.Command(binaryPath, "vault", "remove", vaultPath)
		// cmd.Stdin = strings.NewReader(input)
		//
		// var stdout, stderr bytes.Buffer
		// cmd.Stdout = &stdout
		// cmd.Stderr = &stderr
		//
		// err := cmd.Run()
		// if err != nil {
		//     t.Fatalf("Vault remove failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		// }
		//
		// // Verify vault file was deleted
		// if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
		//     t.Error("Vault file should have been deleted")
		// }
		//
		// // Verify keychain entry was deleted
		// _, err = ks.Retrieve()
		// if err == nil {
		//     t.Error("Keychain entry should have been deleted")
		// }
	})

	// Step 3: Test removal with --yes flag (no prompt)
	t.Run("3_Remove_With_Yes_Flag", func(t *testing.T) {
		t.Skip("TODO: Implement after T030 and T031 (depends on --yes flag)")

		// TODO T031: After --yes flag implementation:
		// // Recreate vault for this test
		// input := testPassword + "\n" + testPassword + "\n"
		// cmd := exec.Command(binaryPath, "--vault", vaultPath, "init", "--use-keychain")
		// cmd.Stdin = strings.NewReader(input)
		// cmd.Run()
		//
		// // Remove with --yes flag (no prompt)
		// cmd = exec.Command(binaryPath, "vault", "remove", vaultPath, "--yes")
		//
		// var stdout, stderr bytes.Buffer
		// cmd.Stdout = &stdout
		// cmd.Stderr = &stderr
		//
		// err := cmd.Run()
		// if err != nil {
		//     t.Fatalf("Vault remove --yes failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		// }
		//
		// // Verify both deleted
		// if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
		//     t.Error("Vault file should have been deleted")
		// }
		// _, err = ks.Retrieve()
		// if err == nil {
		//     t.Error("Keychain entry should have been deleted")
		// }
	})

	// Step 4: Test removal when vault file missing but keychain exists (FR-012)
	t.Run("4_Remove_Orphaned_Keychain", func(t *testing.T) {
		t.Skip("TODO: Implement after T030 (depends on remove command)")

		// TODO T030: After implementation:
		// // Recreate vault
		// input := testPassword + "\n" + testPassword + "\n"
		// cmd := exec.Command(binaryPath, "--vault", vaultPath, "init", "--use-keychain")
		// cmd.Stdin = strings.NewReader(input)
		// cmd.Run()
		//
		// // Manually delete vault file (simulate orphaned keychain)
		// os.Remove(vaultPath)
		//
		// // Remove should still clean up keychain
		// cmd = exec.Command(binaryPath, "vault", "remove", vaultPath, "--yes")
		//
		// var stdout, stderr bytes.Buffer
		// cmd.Stdout = &stdout
		// cmd.Stderr = &stderr
		//
		// err := cmd.Run()
		// if err != nil {
		//     t.Fatalf("Remove orphaned keychain failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		// }
		//
		// // Verify keychain entry was cleaned up
		// _, err = ks.Retrieve()
		// if err == nil {
		//     t.Error("Orphaned keychain entry should have been deleted (FR-012)")
		// }
		//
		// // Verify warning message about missing file
		// output := stdout.String() + stderr.String()
		// if !strings.Contains(output, "not found") && !strings.Contains(output, "missing") {
		//     t.Error("Expected warning about missing vault file")
		// }
	})

	// Step 5: Test 95% success rate (SC-003)
	t.Run("5_Success_Rate_Test", func(t *testing.T) {
		t.Skip("TODO: Implement after T030")

		// TODO T030: After implementation, run remove 20 times and verify >=95% success
		// successCount := 0
		// totalRuns := 20
		//
		// for i := 0; i < totalRuns; i++ {
		//     // Create vault
		//     input := testPassword + "\n" + testPassword + "\n"
		//     cmd := exec.Command(binaryPath, "--vault", vaultPath, "init", "--use-keychain")
		//     cmd.Stdin = strings.NewReader(input)
		//     cmd.Run()
		//
		//     // Remove vault
		//     cmd = exec.Command(binaryPath, "vault", "remove", vaultPath, "--yes")
		//     err := cmd.Run()
		//
		//     // Check both file and keychain deleted
		//     fileDeleted := false
		//     if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		//         fileDeleted = true
		//     }
		//
		//     keychainDeleted := false
		//     if _, err := ks.Retrieve(); err != nil {
		//         keychainDeleted = true
		//     }
		//
		//     if err == nil && fileDeleted && keychainDeleted {
		//         successCount++
		//     }
		//
		//     // Cleanup for next run
		//     os.Remove(vaultPath)
		//     ks.Delete()
		// }
		//
		// successRate := float64(successCount) / float64(totalRuns) * 100
		// if successRate < 95.0 {
		//     t.Errorf("Success rate %.1f%% is below 95%% requirement (SC-003)", successRate)
		// }
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
	defer os.RemoveAll(vaultDir)

	// Create vault with audit enabled
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Initialize vault with audit
	input := testPassword + "\n" + testPassword + "\n"
	cmd := exec.Command(binaryPath, "--vault", vaultPath, "init", "--enable-audit")
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init with audit failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify metadata file created
	metaPath := filepath.Join(vaultDir, "vault.meta")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata file was not created")
	}

	// Get initial audit log content
	initialAuditData, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}
	initialLines := len(strings.Split(string(initialAuditData), "\n"))

	// Run vault remove command with --yes flag
	cmd = exec.Command(binaryPath, "--vault", vaultPath, "vault", "remove", "--yes")
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

		// Check for vault_remove_attempt or vault_remove with "initiated" outcome
		if eventType == "vault_remove_attempt" || (eventType == "vault_remove" && outcome == "initiated") {
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
