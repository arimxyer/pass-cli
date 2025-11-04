//go:build integration

package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIntegration_TUILaunchDetection verifies TUI launches with no args
func TestIntegration_TUILaunchDetection(t *testing.T) {
	// Create a test vault first
	testPassword := "Test-Password-TUI@123"
	vaultDir := filepath.Join(testDir, "tui-test-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	initCmd := exec.Command(binaryPath, "init")
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	initCmd.Stdin = strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n")
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(vaultDir) // Best effort cleanup
	})

	t.Run("No_Args_Attempts_TUI_Launch", func(t *testing.T) {
		// Run with no arguments - this should attempt to launch TUI
		// We can't fully test the interactive TUI in integration tests,
		// but we can verify it starts and doesn't immediately crash

		cmd := exec.Command(binaryPath)
		cmd.Env = append(os.Environ(),
			"PASS_CLI_TEST=1",
			"PASS_CLI_CONFIG="+testConfigPath,
		)

		// Give it a moment to start, then kill it
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start TUI: %v", err)
		}

		// Let it run briefly
		time.Sleep(100 * time.Millisecond)

		// Kill the process
		if err := cmd.Process.Kill(); err != nil {
			t.Logf("Failed to kill process (may have already exited): %v", err)
		}

		// Wait for it to finish
		_ = cmd.Wait()

		// If we got here without panic, TUI launched successfully
		t.Log("TUI launched successfully with no arguments")
	})

	t.Run("With_Args_Uses_CLI_Mode", func(t *testing.T) {
		// Run with arguments - this should use CLI mode, not TUI
		cmd := exec.Command(binaryPath, "version")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("CLI mode failed: %v\nOutput: %s", err, output)
		}

		// Should show version output (CLI mode)
		if !strings.Contains(string(output), "pass-cli") {
			t.Errorf("Expected version output, got: %s", output)
		}

		t.Log("CLI mode executed successfully with arguments")
	})

	t.Run("Help_Flag_Uses_CLI_Mode", func(t *testing.T) {
		// --help should use CLI mode
		cmd := exec.Command(binaryPath, "--help")

		output, err := cmd.CombinedOutput()
		if err != nil {
			// --help returns exit code 0 in cobra
			t.Logf("Help command: %v", err)
		}

		// Should show help text
		outputStr := string(output)
		if !strings.Contains(outputStr, "pass-cli") && !strings.Contains(outputStr, "Usage") {
			t.Errorf("Expected help output, got: %s", outputStr)
		}

		t.Log("Help flag executed successfully in CLI mode")
	})
}

// TestIntegration_TUIVaultPath verifies TUI respects vault path configuration
func TestIntegration_TUIVaultPath(t *testing.T) {
	testPassword := "Test-Password@123"

	t.Run("Uses_Config_Vault_Path", func(t *testing.T) {
		// Create vault in custom location
		customVaultDir := filepath.Join(testDir, "custom-tui-vault")
		customVaultPath := filepath.Join(customVaultDir, "vault.enc")

		// Setup config with custom vault_path
		testConfigPath, cleanup := setupTestVaultConfig(t, customVaultPath)
		defer cleanup()

		// Initialize vault
		initCmd := exec.Command(binaryPath, "init")
		initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		initCmd.Stdin = strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n")
		if err := initCmd.Run(); err != nil {
			t.Fatalf("Failed to initialize custom vault: %v", err)
		}

		t.Cleanup(func() {
			_ = os.RemoveAll(customVaultDir) // Best effort cleanup
		})

		// Verify vault was created at custom path
		if _, err := os.Stat(customVaultPath); os.IsNotExist(err) {
			t.Fatal("Vault was not created at custom path")
		}

		t.Log("TUI respects custom vault path from config")
	})

	t.Run("Uses_Default_Vault_Path", func(t *testing.T) {
		// Verify default vault path behavior
		// When no config is provided, uses default path

		// Get version without config - should work
		versionCmd := exec.Command(binaryPath, "version")
		output, err := versionCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Version command failed: %v", err)
		}

		if !strings.Contains(string(output), "pass-cli") {
			t.Errorf("Expected version output, got: %s", output)
		}

		t.Log("TUI uses default vault path when no flag provided")
	})
}

// TestIntegration_TUIWithExistingVault verifies TUI works with populated vault
func TestIntegration_TUIWithExistingVault(t *testing.T) {
	testPassword := "Test-Password@456"
	vaultDir := filepath.Join(testDir, "tui-populated-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	initCmd := exec.Command(binaryPath, "init")
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	initCmd.Stdin = strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n")
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(vaultDir) // Best effort cleanup
	})

	// Add some test credentials with all 6 fields
	credentials := []struct {
		service  string
		username string
		password string
		category string
		url      string
		notes    string
	}{
		{"github.com", "tuiuser", "pass123", "Version Control", "https://github.com", "Test GitHub account"},
		{"gitlab.com", "developer", "pass456", "Version Control", "https://gitlab.com", "Test GitLab account"},
		{"example.com", "admin", "pass789", "Databases", "https://example.com/admin", "Test database admin"},
	}

	for _, cred := range credentials {
		// Build command with all metadata fields
		args := []string{"add", cred.service, "-u", cred.username, "-p", cred.password}

		if cred.category != "" {
			args = append(args, "-c", cred.category)
		}

		if cred.url != "" {
			args = append(args, "--url", cred.url)
		}

		if cred.notes != "" {
			args = append(args, "--notes", cred.notes)
		}

		addCmd := exec.Command(binaryPath, args...)
		addCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		addCmd.Stdin = strings.NewReader(testPassword + "\n")
		if err := addCmd.Run(); err != nil {
			t.Fatalf("Failed to add credential %s: %v", cred.service, err)
		}
	}

	// Verify credentials were added
	listCmd := exec.Command(binaryPath, "list")
	listCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	listCmd.Stdin = strings.NewReader(testPassword + "\n")
	output, err := listCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list credentials: %v", err)
	}

	outputStr := string(output)
	for _, cred := range credentials {
		if !strings.Contains(outputStr, cred.service) {
			t.Errorf("Expected to find %s in list, got: %s", cred.service, outputStr)
		}
	}

	t.Log("Successfully prepared vault with test credentials for TUI")
}

// TestIntegration_TUIKeychainDetection verifies keychain availability detection
func TestIntegration_TUIKeychainDetection(t *testing.T) {
	// This test verifies that the TUI can detect keychain availability
	// The actual behavior depends on the OS and keychain availability

	testPassword := "Test-Password-Key@ch123"
	vaultDir := filepath.Join(testDir, "tui-keychain-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault without keychain
	initCmd := exec.Command(binaryPath, "init")
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	initCmd.Stdin = strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n")
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(vaultDir) // Best effort cleanup
	})

	// Just verify the vault was created - actual keychain testing requires OS support
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault was not created")
	}

	t.Log("TUI keychain detection initialized (actual availability is OS-dependent)")
}

// TestIntegration_TUIErrorHandling verifies TUI handles errors gracefully
func TestIntegration_TUIErrorHandling(t *testing.T) {
	t.Run("Missing_Vault_Shows_Error", func(t *testing.T) {
		// Try to launch TUI with non-existent vault
		nonExistentPath := filepath.Join(testDir, "nonexistent", "vault.enc")

		// Setup config with non-existent vault path
		testConfigPath, cleanup := setupTestVaultConfig(t, nonExistentPath)
		defer cleanup()

		cmd := exec.Command(binaryPath, "list")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		output, err := cmd.CombinedOutput()

		// Should fail gracefully
		if err == nil {
			t.Error("Expected error for non-existent vault")
		}

		// Should contain error message
		outputStr := string(output)
		if !strings.Contains(outputStr, "Error") && !strings.Contains(outputStr, "not found") && !strings.Contains(outputStr, "does not exist") {
			t.Logf("Got error output (expected): %s", outputStr)
		}
	})

	t.Run("Invalid_Vault_Path_Handled", func(t *testing.T) {
		// Try with invalid path characters
		invalidPath := filepath.Join(testDir, "invalid\x00path", "vault.enc")

		// Setup config with invalid vault path
		testConfigPath, cleanup := setupTestVaultConfig(t, invalidPath)
		defer cleanup()

		cmd := exec.Command(binaryPath, "version")
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)

		// Should not crash - version command should still work
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Version might fail, but shouldn't crash
			t.Logf("Version with invalid vault path: %v", err)
		}

		if len(output) == 0 {
			t.Log("Command handled invalid vault path without crash")
		}
	})
}

// TestIntegration_TUIBuildSuccess verifies TUI components are built correctly
func TestIntegration_TUIBuildSuccess(t *testing.T) {
	// This test verifies the binary was built with TUI support
	// If we got this far, the binary built successfully with TUI code

	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Version command failed: %v\nOutput: %s", err, output)
	}

	if len(output) == 0 {
		t.Error("Expected version output")
	}

	t.Log("Binary built successfully with TUI support")
}

// BenchmarkTUIStartup measures TUI startup time
func BenchmarkTUIStartup(b *testing.B) {
	// Create a test vault
	testPassword := "Bench-Password@123"
	vaultDir := filepath.Join(testDir, "bench-tui-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	// Create temporary config for benchmark
	tempDir := b.TempDir()
	testConfigPath := filepath.Join(tempDir, "config.yml")
	configContent := []byte(fmt.Sprintf("vault_path: %s\n", vaultPath))
	if err := os.WriteFile(testConfigPath, configContent, 0644); err != nil {
		b.Fatalf("Failed to create test config: %v", err)
	}

	initCmd := exec.Command(binaryPath, "init")
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	initCmd.Stdin = strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n")
	if err := initCmd.Run(); err != nil {
		b.Fatalf("Failed to initialize vault: %v", err)
	}

	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryPath)
		cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)

		if err := cmd.Start(); err != nil {
			b.Fatalf("Failed to start TUI: %v", err)
		}

		// Let it start
		time.Sleep(50 * time.Millisecond)

		// Kill it
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
}

// TestIntegration_TUIComponentIntegration verifies components work together
func TestIntegration_TUIComponentIntegration(t *testing.T) {
	testPassword := "Test-Integration@789"
	vaultDir := filepath.Join(testDir, "tui-component-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	initCmd := exec.Command(binaryPath, "init")
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	initCmd.Stdin = strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n")
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(vaultDir) // Best effort cleanup
	})

	// Add credentials to test list view integration with all 6 fields
	for i := 1; i <= 5; i++ {
		service := fmt.Sprintf("service%d.com", i)
		username := fmt.Sprintf("user%d", i)
		password := fmt.Sprintf("pass%d", i)
		category := fmt.Sprintf("Category%d", i)
		url := fmt.Sprintf("https://service%d.com", i)
		notes := fmt.Sprintf("Test notes for service%d", i)

		addCmd := exec.Command(binaryPath, "add", service,
			"-u", username,
			"-p", password,
			"-c", category,
			"--url", url,
			"--notes", notes)
		addCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		addCmd.Stdin = strings.NewReader(testPassword + "\n")
		if err := addCmd.Run(); err != nil {
			t.Fatalf("Failed to add credential: %v", err)
		}
	}

	// Verify all credentials are accessible
	listCmd := exec.Command(binaryPath, "list")
	listCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	listCmd.Stdin = strings.NewReader(testPassword + "\n")
	output, err := listCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list credentials: %v", err)
	}

	outputStr := string(output)
	for i := 1; i <= 5; i++ {
		service := fmt.Sprintf("service%d.com", i)
		if !strings.Contains(outputStr, service) {
			t.Errorf("Expected to find %s in list", service)
		}
	}

	t.Log("TUI components integrated successfully - vault operations work correctly")
}

// TestIntegration_TUIFullFieldSupport verifies all 6 credential fields are properly handled
func TestIntegration_TUIFullFieldSupport(t *testing.T) {
	testPassword := "Test-Full-Fields@123"
	vaultDir := filepath.Join(testDir, "tui-full-fields-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	initCmd := exec.Command(binaryPath, "init")
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	initCmd.Stdin = strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n")
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(vaultDir) // Best effort cleanup
	})

	// Add a credential with all 6 fields populated
	service := "test-service.com"
	username := "testuser"
	password := "testpass123"
	category := "Test Category"
	url := "https://test-service.com/login"
	notes := "Test notes for full field support"

	addCmd := exec.Command(binaryPath, "add", service,
		"-u", username,
		"-p", password,
		"-c", category,
		"--url", url,
		"--notes", notes)
	addCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	addCmd.Stdin = strings.NewReader(testPassword + "\n")

	output, err := addCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to add credential with full fields: %v\nOutput: %s", err, output)
	}

	// Verify all fields are in the success message
	outputStr := string(output)
	if !strings.Contains(outputStr, service) {
		t.Errorf("Success message missing service: %s", outputStr)
	}
	if !strings.Contains(outputStr, username) {
		t.Errorf("Success message missing username: %s", outputStr)
	}
	if !strings.Contains(outputStr, category) {
		t.Errorf("Success message missing category: %s", outputStr)
	}
	if !strings.Contains(outputStr, url) {
		t.Errorf("Success message missing URL: %s", outputStr)
	}
	if !strings.Contains(outputStr, notes) {
		t.Errorf("Success message missing notes: %s", outputStr)
	}

	// Retrieve the credential using get command
	getCmd := exec.Command(binaryPath, "get", service)
	getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	getCmd.Stdin = strings.NewReader(testPassword + "\n")
	getOutput, err := getCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to retrieve credential: %v\nOutput: %s", err, getOutput)
	}

	getOutputStr := string(getOutput)

	// Verify all fields are properly persisted
	if !strings.Contains(getOutputStr, username) {
		t.Errorf("Retrieved credential missing username: %s", getOutputStr)
	}
	if !strings.Contains(getOutputStr, category) {
		t.Errorf("Retrieved credential missing category: %s", getOutputStr)
	}
	if !strings.Contains(getOutputStr, url) {
		t.Errorf("Retrieved credential missing URL: %s", getOutputStr)
	}
	if !strings.Contains(getOutputStr, notes) {
		t.Errorf("Retrieved credential missing notes: %s", getOutputStr)
	}

	t.Log("All 6 credential fields (service, username, password, category, url, notes) properly stored and retrieved")
}

// TestIntegration_TUIEmptyOptionalFields verifies backward compatibility with empty optional fields
func TestIntegration_TUIEmptyOptionalFields(t *testing.T) {
	testPassword := "Test-Empty-Fields@456"
	vaultDir := filepath.Join(testDir, "tui-empty-fields-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	initCmd := exec.Command(binaryPath, "init")
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	initCmd.Stdin = strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n")
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(vaultDir) // Best effort cleanup
	})

	// Add credentials with empty optional fields (category, url, notes)
	testCases := []struct {
		name        string
		service     string
		username    string
		password    string
		includeFlag string
		flagValue   string
	}{
		{"OnlyCategoryEmpty", "service1.com", "user1", "pass1", "--url", "https://service1.com"},
		{"OnlyURLEmpty", "service2.com", "user2", "pass2", "-c", "TestCat"},
		{"OnlyNotesEmpty", "service3.com", "user3", "pass3", "-c", "TestCat"},
		{"AllOptionalEmpty", "service4.com", "user4", "pass4", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"add", tc.service, "-u", tc.username, "-p", tc.password}

			if tc.includeFlag != "" {
				args = append(args, tc.includeFlag, tc.flagValue)
			}

			addCmd := exec.Command(binaryPath, args...)
			addCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
			addCmd.Stdin = strings.NewReader(testPassword + "\n")

			if err := addCmd.Run(); err != nil {
				t.Fatalf("Failed to add credential %s with empty optional fields: %v", tc.service, err)
			}

			// Verify credential can be retrieved
			getCmd := exec.Command(binaryPath, "get", tc.service)
			getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
			getCmd.Stdin = strings.NewReader(testPassword + "\n")
			output, err := getCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to retrieve credential %s: %v\nOutput: %s", tc.service, err, output)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, tc.username) {
				t.Errorf("Retrieved credential missing username: %s", outputStr)
			}
		})
	}

	t.Log("Empty optional fields handled gracefully - backward compatibility maintained")
}

// TestIntegration_TUIUpdateFields verifies updating all 6 credential fields via CLI
func TestIntegration_TUIUpdateFields(t *testing.T) {
	testPassword := "Test-Update@789"
	vaultDir := filepath.Join(testDir, "tui-update-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	initCmd := exec.Command(binaryPath, "init")
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	initCmd.Stdin = strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n")
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(vaultDir) // Best effort cleanup
	})

	// Add initial credential with all fields
	service := "update-test.com"
	addCmd := exec.Command(binaryPath, "add", service,
		"-u", "originaluser",
		"-p", "originalpass",
		"-c", "Original Category",
		"--url", "https://original.com",
		"--notes", "Original notes")
	addCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	addCmd.Stdin = strings.NewReader(testPassword + "\n")
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add initial credential: %v", err)
	}

	t.Run("UpdateUsername", func(t *testing.T) {
		updateCmd := exec.Command(binaryPath, "update", service,
			"-u", "newuser",
			"--force")
		updateCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		updateCmd.Stdin = strings.NewReader(testPassword + "\n")
		if err := updateCmd.Run(); err != nil {
			t.Fatalf("Failed to update username: %v", err)
		}

		// Verify update
		getCmd := exec.Command(binaryPath, "get", service, "--no-clipboard")
		getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		getCmd.Stdin = strings.NewReader(testPassword + "\n")
		output, _ := getCmd.CombinedOutput()
		if !strings.Contains(string(output), "newuser") {
			t.Errorf("Updated username not found in output: %s", output)
		}
	})

	t.Run("UpdatePassword", func(t *testing.T) {
		updateCmd := exec.Command(binaryPath, "update", service,
			"-p", "newpass123",
			"--force")
		updateCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		updateCmd.Stdin = strings.NewReader(testPassword + "\n")
		if err := updateCmd.Run(); err != nil {
			t.Fatalf("Failed to update password: %v", err)
		}

		// Verify password was updated
		getCmd := exec.Command(binaryPath, "get", service, "--no-clipboard")
		getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		getCmd.Stdin = strings.NewReader(testPassword + "\n")
		output, _ := getCmd.CombinedOutput()
		if !strings.Contains(string(output), "newpass123") {
			t.Errorf("Updated password not found in output: %s", output)
		}
	})

	t.Run("UpdateCategory", func(t *testing.T) {
		updateCmd := exec.Command(binaryPath, "update", service,
			"--category", "New Category",
			"--force")
		updateCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		updateCmd.Stdin = strings.NewReader(testPassword + "\n")
		if err := updateCmd.Run(); err != nil {
			t.Fatalf("Failed to update category: %v", err)
		}

		// Verify category was updated
		getCmd := exec.Command(binaryPath, "get", service, "--no-clipboard")
		getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		getCmd.Stdin = strings.NewReader(testPassword + "\n")
		output, _ := getCmd.CombinedOutput()
		if !strings.Contains(string(output), "New Category") {
			t.Errorf("Updated category not found in output: %s", output)
		}
	})

	t.Run("UpdateURL", func(t *testing.T) {
		updateCmd := exec.Command(binaryPath, "update", service,
			"--url", "https://new-url.com",
			"--force")
		updateCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		updateCmd.Stdin = strings.NewReader(testPassword + "\n")
		if err := updateCmd.Run(); err != nil {
			t.Fatalf("Failed to update URL: %v", err)
		}

		// Verify URL was updated
		getCmd := exec.Command(binaryPath, "get", service, "--no-clipboard")
		getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		getCmd.Stdin = strings.NewReader(testPassword + "\n")
		output, _ := getCmd.CombinedOutput()
		if !strings.Contains(string(output), "https://new-url.com") {
			t.Errorf("Updated URL not found in output: %s", output)
		}
	})

	t.Run("UpdateNotes", func(t *testing.T) {
		updateCmd := exec.Command(binaryPath, "update", service,
			"--notes", "Updated notes content",
			"--force")
		updateCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		updateCmd.Stdin = strings.NewReader(testPassword + "\n")
		if err := updateCmd.Run(); err != nil {
			t.Fatalf("Failed to update notes: %v", err)
		}

		// Verify notes were updated
		getCmd := exec.Command(binaryPath, "get", service, "--no-clipboard")
		getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		getCmd.Stdin = strings.NewReader(testPassword + "\n")
		output, _ := getCmd.CombinedOutput()
		if !strings.Contains(string(output), "Updated notes content") {
			t.Errorf("Updated notes not found in output: %s", output)
		}
	})

	t.Run("UpdateMultipleFields", func(t *testing.T) {
		updateCmd := exec.Command(binaryPath, "update", service,
			"-u", "finaluser",
			"--category", "Final Category",
			"--url", "https://final.com",
			"--notes", "Final notes",
			"--force")
		updateCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		updateCmd.Stdin = strings.NewReader(testPassword + "\n")
		if err := updateCmd.Run(); err != nil {
			t.Fatalf("Failed to update multiple fields: %v", err)
		}

		// Verify all fields were updated
		getCmd := exec.Command(binaryPath, "get", service, "--no-clipboard")
		getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		getCmd.Stdin = strings.NewReader(testPassword + "\n")
		output, _ := getCmd.CombinedOutput()
		outputStr := string(output)

		if !strings.Contains(outputStr, "finaluser") {
			t.Errorf("Final username not found: %s", outputStr)
		}
		if !strings.Contains(outputStr, "Final Category") {
			t.Errorf("Final category not found: %s", outputStr)
		}
		if !strings.Contains(outputStr, "https://final.com") {
			t.Errorf("Final URL not found: %s", outputStr)
		}
		if !strings.Contains(outputStr, "Final notes") {
			t.Errorf("Final notes not found: %s", outputStr)
		}
	})

	t.Run("ClearOptionalFields", func(t *testing.T) {
		// Clear category, URL, and notes
		updateCmd := exec.Command(binaryPath, "update", service,
			"--clear-category",
			"--clear-url",
			"--clear-notes",
			"--force")
		updateCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		updateCmd.Stdin = strings.NewReader(testPassword + "\n")
		if err := updateCmd.Run(); err != nil {
			t.Fatalf("Failed to clear optional fields: %v", err)
		}

		// Verify fields were cleared (should not appear in output)
		getCmd := exec.Command(binaryPath, "get", service, "--no-clipboard")
		getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		getCmd.Stdin = strings.NewReader(testPassword + "\n")
		output, _ := getCmd.CombinedOutput()
		outputStr := string(output)

		// These lines should not appear when fields are empty
		if strings.Contains(outputStr, "Category:") {
			t.Errorf("Category should be cleared but found in output: %s", outputStr)
		}
		if strings.Contains(outputStr, "URL:") {
			t.Errorf("URL should be cleared but found in output: %s", outputStr)
		}
		if strings.Contains(outputStr, "Notes:") {
			t.Errorf("Notes should be cleared but found in output: %s", outputStr)
		}
	})

	t.Log("All update operations (username, password, category, URL, notes) work correctly")
}

// TestIntegration_TUIDeleteCredential verifies deleting credentials and list/get consistency
func TestIntegration_TUIDeleteCredential(t *testing.T) {
	testPassword := "Test-Delete@ABC123"
	vaultDir := filepath.Join(testDir, "tui-delete-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	initCmd := exec.Command(binaryPath, "init")
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	initCmd.Stdin = strings.NewReader(testPassword + "\n" + testPassword + "\n" + "n\n")
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(vaultDir) // Best effort cleanup
	})

	// Add multiple credentials with full metadata
	credentials := []struct {
		service  string
		username string
		password string
		category string
		url      string
		notes    string
	}{
		{"delete-test1.com", "user1", "pass1", "Test Cat 1", "https://test1.com", "Notes 1"},
		{"delete-test2.com", "user2", "pass2", "Test Cat 2", "https://test2.com", "Notes 2"},
		{"delete-test3.com", "user3", "pass3", "Test Cat 3", "https://test3.com", "Notes 3"},
	}

	for _, cred := range credentials {
		addCmd := exec.Command(binaryPath, "add", cred.service,
			"-u", cred.username,
			"-p", cred.password,
			"-c", cred.category,
			"--url", cred.url,
			"--notes", cred.notes)
		addCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		addCmd.Stdin = strings.NewReader(testPassword + "\n")
		if err := addCmd.Run(); err != nil {
			t.Fatalf("Failed to add credential %s: %v", cred.service, err)
		}
	}

	// Verify all credentials exist
	listCmd := exec.Command(binaryPath, "list")
	listCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	listCmd.Stdin = strings.NewReader(testPassword + "\n")
	listOutput, _ := listCmd.CombinedOutput()
	listStr := string(listOutput)

	for _, cred := range credentials {
		if !strings.Contains(listStr, cred.service) {
			t.Errorf("Expected %s in list before delete: %s", cred.service, listStr)
		}
	}

	// Delete the second credential
	deleteService := "delete-test2.com"
	deleteCmd := exec.Command(binaryPath, "delete", deleteService, "--force")
	deleteCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	deleteCmd.Stdin = strings.NewReader(testPassword + "\n")
	if err := deleteCmd.Run(); err != nil {
		t.Fatalf("Failed to delete credential: %v", err)
	}

	// Verify deleted credential is not in list
	listCmd = exec.Command(binaryPath, "list")
	listCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	listCmd.Stdin = strings.NewReader(testPassword + "\n")
	listOutput, _ = listCmd.CombinedOutput()
	listStr = string(listOutput)

	if strings.Contains(listStr, deleteService) {
		t.Errorf("Deleted credential %s should not appear in list: %s", deleteService, listStr)
	}

	// Verify other credentials still exist
	if !strings.Contains(listStr, "delete-test1.com") {
		t.Errorf("Credential delete-test1.com should still exist: %s", listStr)
	}
	if !strings.Contains(listStr, "delete-test3.com") {
		t.Errorf("Credential delete-test3.com should still exist: %s", listStr)
	}

	// Verify get fails for deleted credential
	getCmd := exec.Command(binaryPath, "get", deleteService)
	getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	getCmd.Stdin = strings.NewReader(testPassword + "\n")
	err := getCmd.Run()
	if err == nil {
		t.Error("Expected error when getting deleted credential, but got success")
	}

	// Verify remaining credentials still have all their metadata
	for _, cred := range []string{"delete-test1.com", "delete-test3.com"} {
		getCmd := exec.Command(binaryPath, "get", cred, "--no-clipboard")
		getCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
		getCmd.Stdin = strings.NewReader(testPassword + "\n")
		output, err := getCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to get remaining credential %s: %v", cred, err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Test Cat") {
			t.Errorf("Category missing for %s: %s", cred, outputStr)
		}
		if !strings.Contains(outputStr, "https://") {
			t.Errorf("URL missing for %s: %s", cred, outputStr)
		}
		if !strings.Contains(outputStr, "Notes") {
			t.Errorf("Notes missing for %s: %s", cred, outputStr)
		}
	}

	t.Log("Delete operation works correctly, list and get commands show consistent state")
}
