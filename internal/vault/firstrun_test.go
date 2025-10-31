package vault

import (
	"os"
	"testing"
)

// T041: TestDetectFirstRun_VaultExists - Vault present → ShouldPrompt=false
func TestDetectFirstRun_VaultExists(t *testing.T) {
	// Create temporary vault file
	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/vault.enc"
	if err := os.WriteFile(vaultPath, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	// Set default vault path to temp vault for testing
	oldDefault := getDefaultVaultPath
	defer func() { getDefaultVaultPath = oldDefault }()
	getDefaultVaultPath = func() string { return vaultPath }

	// Detect first run for a command that requires vault
	state := DetectFirstRun("get", "")

	// Assertions
	if state.ShouldPrompt {
		t.Error("Expected ShouldPrompt=false when vault exists")
	}
	if !state.VaultExists {
		t.Error("Expected VaultExists=true")
	}
	if state.IsFirstRun {
		t.Error("Expected IsFirstRun=false when vault exists")
	}
}

// T042: TestDetectFirstRun_VaultMissing_RequiresVault - Vault missing, `get` command → ShouldPrompt=true
func TestDetectFirstRun_VaultMissing_RequiresVault(t *testing.T) {
	// Set default vault path to non-existent file
	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/nonexistent.enc"

	oldDefault := getDefaultVaultPath
	defer func() { getDefaultVaultPath = oldDefault }()
	getDefaultVaultPath = func() string { return vaultPath }

	// Detect first run for a command that requires vault
	state := DetectFirstRun("get", "")

	// Assertions
	if !state.ShouldPrompt {
		t.Error("Expected ShouldPrompt=true when vault missing and command requires vault")
	}
	if state.VaultExists {
		t.Error("Expected VaultExists=false")
	}
	if !state.IsFirstRun {
		t.Error("Expected IsFirstRun=true when vault missing")
	}
	if !state.CommandRequiresVault {
		t.Error("Expected CommandRequiresVault=true for 'get' command")
	}
	if state.CustomVaultPath {
		t.Error("Expected CustomVaultPath=false when no custom vault path configured")
	}
}

// T043: TestDetectFirstRun_VaultMissing_NoVaultRequired - Vault missing, `version` command → ShouldPrompt=false
func TestDetectFirstRun_VaultMissing_NoVaultRequired(t *testing.T) {
	// Set default vault path to non-existent file
	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/nonexistent.enc"

	oldDefault := getDefaultVaultPath
	defer func() { getDefaultVaultPath = oldDefault }()
	getDefaultVaultPath = func() string { return vaultPath }

	// Detect first run for a command that doesn't require vault
	state := DetectFirstRun("version", "")

	// Assertions
	if state.ShouldPrompt {
		t.Error("Expected ShouldPrompt=false for 'version' command (no vault required)")
	}
	if state.VaultExists {
		t.Error("Expected VaultExists=false")
	}
	if state.CommandRequiresVault {
		t.Error("Expected CommandRequiresVault=false for 'version' command")
	}
}

// T044: TestDetectFirstRun_CustomVaultPath - custom vault_path configured → ShouldPrompt=false
func TestDetectFirstRun_CustomVaultPath(t *testing.T) {
	// Set default vault path to non-existent file
	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/nonexistent.enc"
	customVaultPath := "/tmp/custom-vault.enc"

	oldDefault := getDefaultVaultPath
	defer func() { getDefaultVaultPath = oldDefault }()
	getDefaultVaultPath = func() string { return vaultPath }

	// Detect first run with custom vault flag
	state := DetectFirstRun("get", customVaultPath)

	// Assertions
	if state.ShouldPrompt {
		t.Error("Expected ShouldPrompt=false when custom vault path is configured")
	}
	if !state.CustomVaultPath {
		t.Error("Expected CustomVaultPath=true when custom vault path is provided")
	}
	if state.VaultPath != customVaultPath {
		t.Errorf("Expected VaultPath=%s, got %s", customVaultPath, state.VaultPath)
	}
}

// T045: TestRunGuidedInit_NonTTY - Stdin piped (non-TTY) → Returns ErrNonTTY
func TestRunGuidedInit_NonTTY(t *testing.T) {
	// This test will verify that RunGuidedInit returns ErrNonTTY when stdin is not a TTY
	// In real implementation, we'll mock the TTY check

	// For now, this test will fail because RunGuidedInit doesn't exist yet
	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/vault.enc"

	err := RunGuidedInit(vaultPath, false) // false = not a TTY

	if err == nil {
		t.Error("Expected error when not running in TTY")
	}
	if err != ErrNonTTY {
		t.Errorf("Expected ErrNonTTY, got %v", err)
	}
}

// T046: TestRunGuidedInit_UserDeclines - User types 'n' → Returns ErrUserDeclined
func TestRunGuidedInit_UserDeclines(t *testing.T) {
	// This test will verify that RunGuidedInit returns ErrUserDeclined when user declines
	// We'll need to mock user input for this test

	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/vault.enc"

	// Mock user input: 'n' to decline
	err := RunGuidedInitWithInput(vaultPath, true, "n\n")

	if err == nil {
		t.Error("Expected error when user declines")
	}
	if err != ErrUserDeclined {
		t.Errorf("Expected ErrUserDeclined, got %v", err)
	}
}

// T047: TestRunGuidedInit_Success - Mock user input → Vault created
func TestRunGuidedInit_Success(t *testing.T) {
	// This test will verify successful guided initialization with mocked input
	// User input: y (proceed), password, password (confirm), y (keychain), y (audit)

	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/vault.enc"

	// Mock complete user input flow
	input := "y\nTestPassword123!\nTestPassword123!\ny\ny\n"
	err := RunGuidedInitWithInput(vaultPath, true, input)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	// Verify vault was created
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Error("Expected vault file to be created")
	}
}

// T048: TestRunGuidedInit_PasswordPolicyFailure - Invalid password 3 times → Error after retry limit
func TestRunGuidedInit_PasswordPolicyFailure(t *testing.T) {
	// This test will verify that guided init fails after 3 invalid password attempts

	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/vault.enc"

	// Mock user input: y (proceed), then 3 invalid passwords
	input := "y\nweak\nweak\nweak\nweak\n"
	err := RunGuidedInitWithInput(vaultPath, true, input)

	if err == nil {
		t.Error("Expected error after 3 invalid password attempts")
	}

	// Vault should NOT be created
	if _, err := os.Stat(vaultPath); err == nil {
		t.Error("Expected vault NOT to be created after password policy failure")
	}
}
