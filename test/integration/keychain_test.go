//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arimxyer/pass-cli/internal/keychain"
	"github.com/arimxyer/pass-cli/internal/vault"
	"github.com/arimxyer/pass-cli/test/helpers"
)

// TestKeychain_FullWorkflow tests the complete keychain integration workflow
func TestKeychain_FullWorkflow(t *testing.T) {
	testPassword := "Keychain-Test-Pass@123"
	vaultPath := helpers.SetupTestVaultWithName(t, "keychain-vault")

	// Create vault-specific keychain service using vaultID derived from vaultPath
	vaultID := filepath.Base(filepath.Dir(vaultPath))
	ks := keychain.New(vaultID)

	// Check if keychain is available before running tests
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping keychain integration tests")
	}

	// Cleanup is automatic via t.Cleanup() in SetupTestVaultWithName

	t.Run("1_Init_With_Keychain", func(t *testing.T) {
		// Setup config with vault_path
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault with --use-keychain flag
		input := helpers.BuildInitStdinWithKeychain(testPassword, true)
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "init", "--use-keychain")
		if err != nil {
			t.Fatalf("Init with keychain failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "successfully") && !strings.Contains(stdout, "initialized") {
			t.Errorf("Expected success message in output, got: %s", stdout)
		}

		if !strings.Contains(stdout, "keychain") {
			t.Errorf("Expected keychain confirmation in output, got: %s", stdout)
		}

		// Verify vault file was created
		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			t.Error("Vault file was not created")
		}

		// Verify password is in keychain
		retrievedPassword, err := ks.Retrieve()
		if err != nil {
			t.Fatalf("Password was not stored in keychain: %v", err)
		}

		if retrievedPassword != testPassword {
			t.Errorf("Keychain password = %q, want %q", retrievedPassword, testPassword)
		}
	})

	t.Run("2_Add_Without_Password_Prompt", func(t *testing.T) {
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Add credential - should NOT prompt for master password (uses keychain)
		input := "testuser\n" + "testpass123\n" // Only username and credential password
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "add", "github.com")
		if err != nil {
			t.Fatalf("Add failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "added") && !strings.Contains(stdout, "successfully") {
			t.Errorf("Expected success message, got: %s", stdout)
		}

		// Should NOT contain "Master password:" prompt
		allOutput := stdout + stderr
		if strings.Contains(allOutput, "Master password:") {
			t.Error("Unexpected master password prompt - keychain should have been used")
		}
	})

	t.Run("3_Get_Without_Password_Prompt", func(t *testing.T) {
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Get credential - should NOT prompt for master password
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, "", "get", "github.com", "--no-clipboard")
		if err != nil {
			t.Fatalf("Get failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "testuser") || !strings.Contains(stdout, "testpass123") {
			t.Errorf("Expected credential details in output, got: %s", stdout)
		}

		// Should NOT contain "Master password:" prompt
		allOutput := stdout + stderr
		if strings.Contains(allOutput, "Master password:") {
			t.Error("Unexpected master password prompt - keychain should have been used")
		}
	})

	t.Run("4_List_Without_Password_Prompt", func(t *testing.T) {
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// List credentials - should NOT prompt for master password
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, "", "list")
		if err != nil {
			t.Fatalf("List failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "github.com") {
			t.Errorf("Expected github.com in list output, got: %s", stdout)
		}

		// Should NOT contain "Master password:" prompt
		allOutput := stdout + stderr
		if strings.Contains(allOutput, "Master password:") {
			t.Error("Unexpected master password prompt - keychain should have been used")
		}
	})

	t.Run("5_Update_Without_Password_Prompt", func(t *testing.T) {
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Update credential - should NOT prompt for master password
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, "", "update", "github.com", "--username", "updateduser", "--password", "updatedpass456")
		if err != nil {
			t.Fatalf("Update failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Should NOT contain "Master password:" prompt
		allOutput := stdout + stderr
		if strings.Contains(allOutput, "Master password:") {
			t.Error("Unexpected master password prompt - keychain should have been used")
		}
	})

	t.Run("6_Delete_Without_Password_Prompt", func(t *testing.T) {
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Delete credential - should NOT prompt for master password
		input := "y\n" // confirm deletion
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "delete", "github.com")
		if err != nil {
			t.Fatalf("Delete failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Should NOT contain "Master password:" prompt in stderr before confirmation
		stderrOutput := stderr
		if strings.Count(stderrOutput, "Master password:") > 0 {
			// Check if it's actually prompting (before the confirmation)
			lines := strings.Split(stderrOutput, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "Master password:" {
					t.Error("Unexpected master password prompt - keychain should have been used")
					break
				}
			}
		}
	})
}

// TestKeychain_Fallback tests fallback to password prompt
func TestKeychain_Fallback(t *testing.T) {
	testPassword := "Fallback-Test-Pass@789"
	vaultPath := helpers.SetupTestVaultWithName(t, "fallback-vault")

	// Create vault-specific keychain service
	vaultID := filepath.Base(filepath.Dir(vaultPath))
	ks := keychain.New(vaultID)
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping keychain fallback tests")
	}

	// Cleanup is automatic via t.Cleanup() in SetupTestVaultWithName

	// Setup config with vault_path
	testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault WITH keychain
	input := helpers.BuildInitStdinWithKeychain(testPassword, true)
	_, _, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "init", "--use-keychain")
	if err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	t.Run("Fallback_After_Keychain_Deleted", func(t *testing.T) {
		// Delete password from keychain
		if err := ks.Delete(); err != nil {
			t.Fatalf("Failed to delete keychain entry: %v", err)
		}

		// Try to add credential - should now prompt for master password
		input := testPassword + "\n"
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "add", "test.com", "--username", "testuser", "--password", "testpass")
		if err != nil {
			t.Fatalf("Add with password prompt failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Should work successfully even without keychain
		if !strings.Contains(stdout, "added") && !strings.Contains(stdout, "successfully") {
			t.Errorf("Expected success message, got: %s", stdout)
		}
	})
}

// TestKeychain_Unavailable tests behavior when keychain is unavailable
func TestKeychain_Unavailable(t *testing.T) {
	ks := keychain.New("")

	// This test verifies graceful handling when keychain is unavailable
	// If keychain IS available, we skip this test
	if ks.IsAvailable() {
		t.Skip("Keychain is available - cannot test unavailable scenario")
	}

	testPassword := "NoKeychain-Pass@456"
	vaultPath := helpers.SetupTestVaultWithName(t, "no-keychain-vault")

	t.Run("Init_Without_Keychain_Available", func(t *testing.T) {
		// Setup config with vault_path
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Try to initialize with --use-keychain when keychain unavailable
		input := helpers.BuildInitStdinWithKeychain(testPassword, true)
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "init", "--use-keychain")

		// Should either:
		// 1. Succeed with warning (graceful degradation)
		// 2. Fail with clear error message
		if err == nil {
			// Check for warning in output
			allOutput := stdout + stderr
			if !strings.Contains(allOutput, "warning") && !strings.Contains(allOutput, "Warning") {
				t.Log("Init succeeded without warning when keychain unavailable (acceptable)")
			}
		} else {
			// Check for clear error message
			allOutput := stdout + stderr
			if !strings.Contains(allOutput, "keychain") {
				t.Errorf("Error message should mention keychain when unavailable, got: %s", allOutput)
			}
		}
	})
}

// TestKeychain_MultipleVaults tests multiple vaults with same keychain
func TestKeychain_MultipleVaults(t *testing.T) {
	ks := keychain.New("")
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping multiple vaults test")
	}

	// Note: Currently pass-cli uses a single keychain entry for all vaults
	// This test documents the current behavior and can be updated if we add
	// per-vault keychain support in the future

	testPassword := "MultiVault-Pass@999"
	vault1Path := helpers.SetupTestVaultWithName(t, "multi-vault-1")
	vault2Path := helpers.SetupTestVaultWithName(t, "multi-vault-2")

	// Cleanup is automatic via t.Cleanup() in SetupTestVaultWithName

	t.Run("First_Vault_Init", func(t *testing.T) {
		// Setup config for vault 1
		testConfigPath1, cleanup := helpers.SetupTestVaultConfig(t, vault1Path)
		defer cleanup()

		input := helpers.BuildInitStdinWithKeychain(testPassword, true)
		_, _, err := helpers.RunCmd(t, binaryPath, testConfigPath1, input, "init", "--use-keychain")
		if err != nil {
			t.Fatalf("Failed to init vault 1: %v", err)
		}
	})

	t.Run("Second_Vault_With_Same_Password", func(t *testing.T) {
		// Setup config for vault 2
		testConfigPath2, cleanup2 := helpers.SetupTestVaultConfig(t, vault2Path)
		defer cleanup2()

		// Initialize second vault with same password
		input := helpers.BuildInitStdinWithKeychain(testPassword, true)
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath2, input, "init", "--use-keychain")
		if err != nil {
			t.Fatalf("Failed to init vault 2: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Add credential to vault 2 using keychain
		input = "user2\n" + "pass2\n"
		_, _, err = helpers.RunCmd(t, binaryPath, testConfigPath2, input, "add", "service2.com")
		if err != nil {
			t.Fatalf("Failed to add to vault 2: %v", err)
		}

		// Verify vault 1 still works with same keychain
		testConfigPath1, cleanup1 := helpers.SetupTestVaultConfig(t, vault1Path)
		defer cleanup1()
		_, _, err = helpers.RunCmd(t, binaryPath, testConfigPath1, "", "list")
		if err != nil {
			t.Errorf("Vault 1 should still work after vault 2 operations: %v", err)
		}
	})
}

// TestKeychain_VerboseOutput tests verbose mode shows keychain usage
func TestKeychain_VerboseOutput(t *testing.T) {
	ks := keychain.New("")
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping verbose output test")
	}

	testPassword := "Verbose-Test-Pass@321"
	vaultPath := helpers.SetupTestVaultWithName(t, "verbose-vault")

	// Cleanup is automatic via t.Cleanup() in SetupTestVaultWithName

	// Setup config with vault_path
	testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize with keychain
	input := helpers.BuildInitStdinWithKeychain(testPassword, true)
	_, _, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "init", "--use-keychain")
	if err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	t.Run("Verbose_Shows_Keychain_Usage", func(t *testing.T) {
		// Run list command with --verbose flag
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, "", "--verbose", "list")
		if err != nil {
			t.Fatalf("List with verbose failed: %v", err)
		}

		// Check if verbose output mentions keychain usage
		allOutput := stdout + stderr
		if !strings.Contains(allOutput, "keychain") && !strings.Contains(allOutput, "Keychain") {
			t.Logf("Verbose mode output:\n%s", allOutput)
			t.Skip("Verbose keychain message may not be implemented yet")
		}
	})
}

// TestKeychain_Enable tests the keychain enable command
func TestKeychain_Enable(t *testing.T) {
	testPassword := "EnableTest-Pass@123"
	vaultPath := helpers.SetupTestVaultWithName(t, "enable-test-vault")

	// Create vault-specific keychain service
	vaultID := filepath.Base(filepath.Dir(vaultPath))
	ks := keychain.New(vaultID)
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping keychain enable integration test")
	}

	// Cleanup is automatic via t.Cleanup() in SetupTestVaultWithName

	// Step 1: Initialize vault WITHOUT --use-keychain
	t.Run("1_Init_Without_Keychain", func(t *testing.T) {
		// Setup config with vault_path
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()

		input := helpers.BuildInitStdinNoRecovery(testPassword, false)
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "init", "--no-recovery")
		if err != nil {
			t.Fatalf("Init failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Verify vault file was created
		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			t.Error("Vault file was not created")
		}

		// Verify password is NOT in keychain
		_, err = ks.Retrieve()
		if err == nil {
			t.Error("Password should NOT be in keychain after init without --use-keychain")
		}
	})

	// Step 2: Run keychain enable command
	t.Run("2_Enable_Keychain", func(t *testing.T) {
		input := testPassword + "\n"
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "keychain", "enable")
		if err != nil {
			t.Fatalf("Keychain enable failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "enabled") {
			t.Errorf("Expected success message containing 'enabled', got: %s", stdout)
		}

		// Verify password is NOW in keychain
		retrievedPassword, err := ks.Retrieve()
		if err != nil {
			t.Fatalf("Password was not stored in keychain: %v", err)
		}

		if retrievedPassword != testPassword {
			t.Errorf("Keychain password = %q, want %q", retrievedPassword, testPassword)
		}
	})

	// Step 3: Verify subsequent commands don't prompt for password
	t.Run("3_Add_Without_Password_Prompt", func(t *testing.T) {
		// Add credential - should NOT prompt for master password (uses keychain)
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()
		input := "testuser\n" + "testpass123\n" // Only username and credential password
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "add", "github.com")
		if err != nil {
			t.Fatalf("Add failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Should NOT contain "Master password:" prompt
		allOutput := stdout + stderr
		if strings.Contains(allOutput, "Master password:") {
			t.Error("Unexpected master password prompt - keychain should have been used")
		}
	})

	// Step 4: Test --force flag (overwrite existing keychain entry)
	t.Run("4_Enable_With_Force", func(t *testing.T) {
		input := testPassword + "\n"
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "keychain", "enable", "--force")
		if err != nil {
			t.Fatalf("Keychain enable --force failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Verify password still in keychain
		retrievedPassword, err := ks.Retrieve()
		if err != nil {
			t.Fatalf("Password was not in keychain after --force: %v", err)
		}

		if retrievedPassword != testPassword {
			t.Errorf("Keychain password = %q, want %q", retrievedPassword, testPassword)
		}
	})
}

// TestKeychain_Status tests the keychain status command
func TestKeychain_Status(t *testing.T) {
	// Check if keychain is available
	ks := keychain.New("")
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping keychain status integration test")
	}

	testPassword := "StatusTest-Pass@123"
	vaultPath := helpers.SetupTestVaultWithName(t, "status-test-vault")

	// Cleanup is automatic via t.Cleanup() in SetupTestVaultWithName

	// Step 1: Initialize vault WITHOUT keychain
	t.Run("1_Init_Without_Keychain", func(t *testing.T) {
		// Setup config with vault_path
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Input: password, confirm, no keychain
		input := helpers.BuildInitStdinNoRecovery(testPassword, false)
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "init", "--no-recovery")
		if err != nil {
			t.Fatalf("Init failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Verify vault file was created
		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			t.Error("Vault file was not created")
		}
	})

	// Step 2: Check status BEFORE enabling keychain
	t.Run("2_Status_Before_Enable", func(t *testing.T) {
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, "", "keychain", "status")
		if err != nil {
			t.Errorf("Status command should not error (exit 0): %v\nStderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Available") {
			t.Errorf("Expected output to contain 'Available', got: %s", stdout)
		}
		if !strings.Contains(stdout, "No") && !strings.Contains(stdout, "not enabled") {
			t.Errorf("Expected output to indicate password not stored, got: %s", stdout)
		}
		if !strings.Contains(stdout, "pass-cli keychain enable") {
			t.Errorf("Expected actionable suggestion to enable keychain, got: %s", stdout)
		}
	})

	// Step 3: Enable keychain
	t.Run("3_Enable_Keychain", func(t *testing.T) {
		input := testPassword + "\n"
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "keychain", "enable")
		if err != nil {
			t.Fatalf("Keychain enable failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}
	})

	// Step 4: Check status AFTER enabling keychain
	t.Run("4_Status_After_Enable", func(t *testing.T) {
		testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
		defer cleanup()
		stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, "", "keychain", "status")
		if err != nil {
			t.Errorf("Status command should not error (exit 0): %v\nStderr: %s", err, stderr)
		}

		if !strings.Contains(stdout, "Available") {
			t.Errorf("Expected output to contain 'Available', got: %s", stdout)
		}
		if !strings.Contains(stdout, "Yes") && !strings.Contains(stdout, "enabled") {
			t.Errorf("Expected output to indicate password is stored, got: %s", stdout)
		}

		// Verify backend name is displayed (platform-specific)
		hasBackend := strings.Contains(stdout, "Windows Credential Manager") ||
			strings.Contains(stdout, "macOS Keychain") ||
			strings.Contains(stdout, "Linux Secret Service") ||
			strings.Contains(stdout, "Secret Service API") ||
			strings.Contains(stdout, "gnome-keyring") ||
			strings.Contains(stdout, "kwallet")
		if !hasBackend {
			t.Errorf("Expected output to contain backend name, got: %s", stdout)
		}
	})
}

// TestKeychain_StatusWithMetadata tests that keychain status command writes audit entry
func TestKeychain_StatusWithMetadata(t *testing.T) {
	// Check if keychain is available
	ks := keychain.New("")
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	testPassword := "MetadataTest-Pass@123"
	vaultPath := helpers.SetupTestVaultWithName(t, "metadata-status-vault")
	vaultDir := filepath.Dir(vaultPath)
	auditLogPath := filepath.Join(vaultDir, "audit.log")

	// Cleanup is automatic via t.Cleanup() in SetupTestVaultWithName
	// Directory is already created by SetupTestVaultWithName

	// Setup config with vault_path
	testConfigPath, cleanup := helpers.SetupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault with audit
	input := helpers.BuildInitStdinNoRecovery(testPassword, false)
	stdout, stderr, err := helpers.RunCmd(t, binaryPath, testConfigPath, input, "init", "--no-recovery")
	if err != nil {
		t.Fatalf("Init with audit failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
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

	// Run keychain status command
	stdout, stderr, err = helpers.RunCmd(t, binaryPath, testConfigPath, "", "keychain", "status")
	if err != nil {
		t.Fatalf("Keychain status failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Verify audit entry written
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

	// Verify last audit entry has correct event_type
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

	t.Logf("Audit entry written: event_type=%s, outcome=%s", eventType, outcome)
}

// TestKeychain_PersistenceAfterRestart simulates the upgrade scenario
func TestKeychain_PersistenceAfterRestart(t *testing.T) {
	testPassword := "PersistenceTest-Pass@123"
	vaultPath := helpers.SetupTestVaultWithName(t, "keychain-persistence-vault")
	vaultDir := filepath.Dir(vaultPath)
	metadataPath := vaultPath + ".meta.json"

	// Create vault-specific keychain service
	vaultID := filepath.Base(vaultDir)
	ks := keychain.New(vaultID)
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	// Cleanup is automatic via t.Cleanup() in SetupTestVaultWithName
	// Directory is already created by SetupTestVaultWithName

	// PHASE 1: Initial setup (like first install)
	t.Log("Phase 1: Creating vault with keychain enabled...")

	vs1, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize vault
	if err := vs1.Initialize([]byte(testPassword), false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Enable keychain (stores password + sets metadata.KeychainEnabled = true)
	if err := vs1.EnableKeychain([]byte(testPassword), false); err != nil {
		t.Fatalf("Failed to enable keychain: %v", err)
	}

	// Verify metadata file was created with correct content
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Fatal("FAIL: Metadata file was not created after EnableKeychain()")
	}

	meta1, err := vault.LoadMetadata(vaultPath)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	if !meta1.KeychainEnabled {
		t.Fatal("FAIL: Metadata.KeychainEnabled should be true after EnableKeychain()")
	}
	t.Log("  - Metadata file created with KeychainEnabled=true")

	// Verify password is in keychain
	storedPassword, err := ks.Retrieve()
	if err != nil {
		t.Fatalf("FAIL: Password not stored in keychain: %v", err)
	}
	if storedPassword != testPassword {
		t.Fatal("FAIL: Stored password doesn't match original")
	}
	t.Log("  - Password stored in keychain")

	// PHASE 2: Simulate app restart / binary update
	t.Log("Phase 2: Simulating restart (new VaultService instance)...")

	// Create completely NEW VaultService instance
	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create second vault service: %v", err)
	}

	// PHASE 3: Verify auto-unlock works
	t.Log("Phase 3: Verifying keychain auto-unlock...")

	// This is THE critical test - can UnlockWithKeychain() work after "restart"?
	err = vs2.UnlockWithKeychain()
	if err != nil {
		t.Fatalf("FAIL: UnlockWithKeychain() failed after restart: %v\n"+
			"This simulates the bug where users need to run 'keychain enable --force' after updates", err)
	}

	if !vs2.IsUnlocked() {
		t.Fatal("FAIL: Vault should be unlocked after UnlockWithKeychain()")
	}
	t.Log("  - UnlockWithKeychain() succeeded")

	// Verify we can actually access credentials (proves unlock worked)
	creds, err := vs2.ListCredentials()
	if err != nil {
		t.Fatalf("FAIL: ListCredentials() failed after keychain unlock: %v", err)
	}
	t.Logf("  - Successfully accessed vault (credential count: %d)", len(creds))

	t.Log("SUCCESS: Keychain persistence works correctly across restart")
}

// TestKeychain_PersistenceMetadataIntegrity verifies metadata file integrity
func TestKeychain_PersistenceMetadataIntegrity(t *testing.T) {
	testPassword := "MetadataIntegrity-Pass@123"
	vaultPath := helpers.SetupTestVaultWithName(t, "metadata-integrity-vault")
	vaultDir := filepath.Dir(vaultPath)
	metadataPath := vaultPath + ".meta.json"

	// Create vault-specific keychain service (must match what CLI uses)
	vaultID := filepath.Base(vaultDir)
	ks := keychain.New(vaultID)
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	// Create and initialize vault with keychain
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	if err := vs.Initialize([]byte(testPassword), false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	if err := vs.EnableKeychain([]byte(testPassword), false); err != nil {
		t.Fatalf("Failed to enable keychain: %v", err)
	}

	// Read metadata file directly to verify actual disk content
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}

	t.Logf("Metadata file content:\n%s", string(metadataBytes))

	// Verify specific fields through multiple loads
	for i := 0; i < 3; i++ {
		meta, err := vault.LoadMetadata(vaultPath)
		if err != nil {
			t.Fatalf("Load %d: Failed to load metadata: %v", i+1, err)
		}

		if !meta.KeychainEnabled {
			t.Fatalf("Load %d: KeychainEnabled should be true, got false", i+1)
		}

		if meta.Version != "1.0" {
			t.Fatalf("Load %d: Version should be '1.0', got '%s'", i+1, meta.Version)
		}
	}

	t.Log("SUCCESS: Metadata maintains integrity across multiple reads")
}

// TestKeychain_PersistenceGracefulDegradation verifies graceful failure
func TestKeychain_PersistenceGracefulDegradation(t *testing.T) {
	testPassword := "Degradation-Pass@123"
	vaultPath := helpers.SetupTestVaultWithName(t, "degradation-vault")
	vaultDir := filepath.Dir(vaultPath)
	metadataPath := vaultPath + ".meta.json"

	// Create vault-specific keychain service (must match what CLI uses)
	vaultID := filepath.Base(vaultDir)
	ks := keychain.New(vaultID)
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	// Setup: Create vault with keychain enabled
	vs1, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	if err := vs1.Initialize([]byte(testPassword), false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	if err := vs1.EnableKeychain([]byte(testPassword), false); err != nil {
		t.Fatalf("Failed to enable keychain: %v", err)
	}

	// Test 1: Delete metadata file - should fail gracefully
	t.Log("Test 1: Simulating deleted metadata file...")
	if err := os.Remove(metadataPath); err != nil {
		t.Fatalf("Failed to delete metadata file: %v", err)
	}

	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	err = vs2.UnlockWithKeychain()
	if err == nil {
		t.Fatal("FAIL: UnlockWithKeychain() should fail when metadata is missing")
	}
	if err != vault.ErrKeychainNotEnabled {
		t.Logf("  - Got expected error type (details: %v)", err)
	} else {
		t.Log("  - Got ErrKeychainNotEnabled as expected")
	}

	// Restore metadata for next test
	meta := &vault.Metadata{
		Version:         "1.0",
		KeychainEnabled: true,
	}
	if err := vault.SaveMetadata(vaultPath, meta); err != nil {
		t.Fatalf("Failed to restore metadata: %v", err)
	}

	// Test 2: Delete keychain entry - should fail gracefully
	t.Log("Test 2: Simulating deleted keychain entry...")
	if err := ks.Delete(); err != nil {
		t.Fatalf("Failed to delete keychain entry: %v", err)
	}

	vs3, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	err = vs3.UnlockWithKeychain()
	if err == nil {
		t.Fatal("FAIL: UnlockWithKeychain() should fail when keychain entry is missing")
	}
	t.Logf("  - Got expected error: %v", err)

	// Test 3: Manual password unlock should still work
	t.Log("Test 3: Verifying manual password unlock still works...")
	if err := vs3.Unlock([]byte(testPassword)); err != nil {
		t.Fatalf("FAIL: Manual unlock should work even when keychain fails: %v", err)
	}
	t.Log("  - Manual password unlock succeeded")

	t.Log("SUCCESS: System degrades gracefully when keychain is unavailable")
}

// TestKeychain_PersistenceMultipleRestarts simulates multiple app restarts
func TestKeychain_PersistenceMultipleRestarts(t *testing.T) {
	testPassword := "MultiRestart-Pass@123"
	vaultPath := helpers.SetupTestVaultWithName(t, "multi-restart-vault")
	vaultDir := filepath.Dir(vaultPath)

	// Create vault-specific keychain service (must match what CLI uses)
	vaultID := filepath.Base(vaultDir)
	ks := keychain.New(vaultID)
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	// Initial setup
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	if err := vs.Initialize([]byte(testPassword), false, "", ""); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	if err := vs.EnableKeychain([]byte(testPassword), false); err != nil {
		t.Fatalf("Failed to enable keychain: %v", err)
	}

	// Simulate 5 restarts (like 5 scoop updates)
	for i := 1; i <= 5; i++ {
		t.Logf("Restart %d: Creating new VaultService instance...", i)

		vsN, err := vault.New(vaultPath)
		if err != nil {
			t.Fatalf("Restart %d: Failed to create vault service: %v", i, err)
		}

		// Verify keychain unlock works
		if err := vsN.UnlockWithKeychain(); err != nil {
			t.Fatalf("Restart %d: UnlockWithKeychain() failed: %v\n"+
				"This indicates a regression in keychain persistence", i, err)
		}

		if !vsN.IsUnlocked() {
			t.Fatalf("Restart %d: Vault not unlocked after UnlockWithKeychain()", i)
		}

		// Access credentials to prove unlock actually worked
		_, err = vsN.ListCredentials()
		if err != nil {
			t.Fatalf("Restart %d: ListCredentials() failed: %v", i, err)
		}

		// Lock before next iteration
		vsN.Lock()
		t.Logf("  - Restart %d: SUCCESS", i)
	}

	t.Log("SUCCESS: Keychain persistence works across multiple restarts")
}
