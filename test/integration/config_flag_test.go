//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/arimxyer/pass-cli/test/helpers"
)

// TestIntegration_ConfigFlagWithoutDefaultVault tests issue #65:
// pass-cli should work with --config flag pointing to custom vault_path
// even when no vault exists at the default location.
func TestIntegration_ConfigFlagWithoutDefaultVault(t *testing.T) {
	testPassword := "Test-Config-Flag@123"

	// Create a custom vault location (NOT in the default ~/.pass-cli location)
	customVaultDir := t.TempDir()
	customVaultPath := filepath.Join(customVaultDir, "custom-vault", "vault.enc")

	// Ensure the custom vault directory exists
	if err := os.MkdirAll(filepath.Dir(customVaultPath), 0700); err != nil {
		t.Fatalf("Failed to create custom vault directory: %v", err)
	}

	// Create config file in a separate location
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yml")
	configContent := "vault_path: " + customVaultPath + "\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Register cleanup for keychain entries
	t.Cleanup(func() {
		helpers.CleanupVaultPath(t, customVaultPath)
	})

	t.Run("init_with_config_flag", func(t *testing.T) {
		// Initialize vault using --config flag (NOT environment variable)
		input := helpers.BuildInitStdin(helpers.DefaultInitOptions(testPassword))
		stdout, stderr, err := runCmdWithConfigFlag(t, configPath, input, "init")

		if err != nil {
			t.Fatalf("Init with --config flag failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Verify vault was created at custom location
		if _, err := os.Stat(customVaultPath); os.IsNotExist(err) {
			t.Errorf("Vault was not created at custom path %s", customVaultPath)
		}
	})

	t.Run("add_credential_with_config_flag", func(t *testing.T) {
		input := helpers.BuildUnlockStdin(testPassword)
		stdout, stderr, err := runCmdWithConfigFlag(t, configPath, input,
			"add", "test-service", "--username", "testuser", "--password", "testpass123")

		if err != nil {
			t.Fatalf("Add with --config flag failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "added") && !strings.Contains(stdout, "successfully") {
			t.Errorf("Expected success message, got: %s", stdout)
		}
	})

	t.Run("get_credential_with_config_flag", func(t *testing.T) {
		input := helpers.BuildUnlockStdin(testPassword)
		stdout, stderr, err := runCmdWithConfigFlag(t, configPath, input,
			"get", "test-service", "--no-clipboard")

		if err != nil {
			t.Fatalf("Get with --config flag failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "testuser") || !strings.Contains(stdout, "testpass123") {
			t.Errorf("Expected credential in output, got: %s", stdout)
		}
	})

	t.Run("list_with_config_flag", func(t *testing.T) {
		input := helpers.BuildUnlockStdin(testPassword)
		stdout, stderr, err := runCmdWithConfigFlag(t, configPath, input, "list")

		if err != nil {
			t.Fatalf("List with --config flag failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if !strings.Contains(stdout, "test-service") {
			t.Errorf("Expected test-service in list output, got: %s", stdout)
		}
	})

	t.Run("doctor_with_config_flag", func(t *testing.T) {
		stdout, stderr, err := runCmdWithConfigFlag(t, configPath, "", "doctor")

		// Doctor exits with code 1 for warnings, which is expected
		// Just check that it ran and shows the correct vault path
		_ = err

		// Doctor should show the custom vault path
		if !strings.Contains(stdout, customVaultPath) && !strings.Contains(stdout, "custom-vault") {
			t.Errorf("Expected custom vault path in doctor output, got: %s\nStderr: %s", stdout, stderr)
		}
	})
}

// runCmdWithConfigFlag executes pass-cli using the --config flag (not environment variable).
// This tests the exact scenario from issue #65.
func runCmdWithConfigFlag(t *testing.T, configPath, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	// Prepend --config flag to args
	fullArgs := append([]string{"--config", configPath}, args...)

	// Build environment variables
	// On macOS, we cannot override HOME because it breaks keychain access
	// (macOS keychain is tied to the user session, not the HOME directory).
	// On other platforms, use a fake HOME to isolate from user's global config.
	envVars := []string{"PASS_CLI_TEST=1"}
	if runtime.GOOS != "darwin" {
		fakeHome := t.TempDir()
		envVars = append(envVars, "HOME="+fakeHome)
	}

	// Run WITHOUT setting PASS_CLI_CONFIG env var - use only the flag
	return helpers.RunCmdWithEnv(t, binaryPath, stdin, envVars, fullArgs...)
}
