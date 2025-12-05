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

	"pass-cli/internal/storage"
)

// TestCLI_InitCreatesV2VaultByDefault verifies that `pass-cli init` creates
// a v2 vault with key wrapping when recovery is enabled (default behavior).
func TestCLI_InitCreatesV2VaultByDefault(t *testing.T) {
	// Create isolated test directory
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	configPath := filepath.Join(tempDir, "config.yml")

	// Write config file
	configContent := "vault_path: " + vaultPath + "\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Input for init command:
	// 1. password
	// 2. confirm password
	// 3. "n" for passphrase
	// 4. "n" to skip verification
	testPassword := "Test123!@#Password"
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"

	cmd := exec.Command(binaryPath, "init")
	cmd.Stdin = strings.NewReader(input)
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+configPath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("init command failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify success message
	if !strings.Contains(stdout.String(), "successfully") {
		t.Errorf("Expected success message, got: %s", stdout.String())
	}

	// Verify recovery phrase was displayed
	if !strings.Contains(stdout.String(), "Recovery Phrase") {
		t.Errorf("Expected recovery phrase to be displayed, got: %s", stdout.String())
	}

	// Verify vault file exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault file was not created")
	}

	// Read and parse vault file to check version
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault file: %v", err)
	}

	var encryptedVault struct {
		Metadata storage.VaultMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(vaultData, &encryptedVault); err != nil {
		t.Fatalf("Failed to parse vault file: %v", err)
	}

	// Verify it's a v2 vault
	if encryptedVault.Metadata.Version != 2 {
		t.Errorf("Expected vault version 2, got %d", encryptedVault.Metadata.Version)
	}

	// Verify WrappedDEK fields are present (v2 signature)
	if len(encryptedVault.Metadata.WrappedDEK) == 0 {
		t.Error("WrappedDEK field is empty - vault is not v2 format")
	}
	if len(encryptedVault.Metadata.WrappedDEKNonce) == 0 {
		t.Error("WrappedDEKNonce field is empty - vault is not v2 format")
	}
}

// TestCLI_InitNoRecoveryCreatesV1Vault verifies that `pass-cli init --no-recovery`
// creates a v1 vault without key wrapping.
func TestCLI_InitNoRecoveryCreatesV1Vault(t *testing.T) {
	// Create isolated test directory
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	configPath := filepath.Join(tempDir, "config.yml")

	// Write config file
	configContent := "vault_path: " + vaultPath + "\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Input for init command:
	// 1. password
	// 2. confirm password
	testPassword := "Test123!@#Password"
	input := testPassword + "\n" + testPassword + "\n"

	cmd := exec.Command(binaryPath, "init", "--no-recovery")
	cmd.Stdin = strings.NewReader(input)
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+configPath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("init --no-recovery command failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify success message
	if !strings.Contains(stdout.String(), "successfully") {
		t.Errorf("Expected success message, got: %s", stdout.String())
	}

	// Verify NO recovery phrase was displayed
	if strings.Contains(stdout.String(), "Recovery Phrase") {
		t.Errorf("Did not expect recovery phrase with --no-recovery flag, got: %s", stdout.String())
	}

	// Verify vault file exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault file was not created")
	}

	// Read and parse vault file to check version
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault file: %v", err)
	}

	var encryptedVault struct {
		Metadata storage.VaultMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(vaultData, &encryptedVault); err != nil {
		t.Fatalf("Failed to parse vault file: %v", err)
	}

	// Verify it's a v1 vault
	if encryptedVault.Metadata.Version != 1 {
		t.Errorf("Expected vault version 1, got %d", encryptedVault.Metadata.Version)
	}

	// Verify WrappedDEK fields are NOT present (v1 signature)
	if len(encryptedVault.Metadata.WrappedDEK) != 0 {
		t.Error("WrappedDEK field should be empty for v1 vault")
	}
}

// TestCLI_VaultMigrateUpgradesV1ToV2 verifies that `pass-cli vault migrate`
// upgrades a v1 vault to v2 format.
func TestCLI_VaultMigrateUpgradesV1ToV2(t *testing.T) {
	// Create isolated test directory
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	configPath := filepath.Join(tempDir, "config.yml")

	// Write config file
	configContent := "vault_path: " + vaultPath + "\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	testPassword := "Test123!@#Password"

	// Step 1: Create v1 vault with --no-recovery
	initInput := testPassword + "\n" + testPassword + "\n"
	initCmd := exec.Command(binaryPath, "init", "--no-recovery")
	initCmd.Stdin = strings.NewReader(initInput)
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+configPath)

	var initStdout, initStderr bytes.Buffer
	initCmd.Stdout = &initStdout
	initCmd.Stderr = &initStderr

	if err := initCmd.Run(); err != nil {
		t.Fatalf("init --no-recovery failed: %v\nStdout: %s\nStderr: %s", err, initStdout.String(), initStderr.String())
	}

	// Verify it's v1
	vaultDataBefore, _ := os.ReadFile(vaultPath)
	var vaultBefore struct {
		Metadata storage.VaultMetadata `json:"metadata"`
	}
	json.Unmarshal(vaultDataBefore, &vaultBefore)
	if vaultBefore.Metadata.Version != 1 {
		t.Fatalf("Expected v1 vault before migration, got v%d", vaultBefore.Metadata.Version)
	}

	// Step 2: Run migration
	// Input: "y" to proceed, password, "n" for passphrase, "n" to skip verification
	migrateInput := "y\n" + testPassword + "\n" + "n\n" + "n\n"
	migrateCmd := exec.Command(binaryPath, "vault", "migrate")
	migrateCmd.Stdin = strings.NewReader(migrateInput)
	migrateCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+configPath)

	var migrateStdout, migrateStderr bytes.Buffer
	migrateCmd.Stdout = &migrateStdout
	migrateCmd.Stderr = &migrateStderr

	if err := migrateCmd.Run(); err != nil {
		t.Fatalf("vault migrate failed: %v\nStdout: %s\nStderr: %s", err, migrateStdout.String(), migrateStderr.String())
	}

	// Verify success message
	if !strings.Contains(migrateStdout.String(), "successfully") {
		t.Errorf("Expected success message, got: %s", migrateStdout.String())
	}

	// Verify recovery phrase was displayed
	if !strings.Contains(migrateStdout.String(), "Recovery Phrase") {
		t.Errorf("Expected recovery phrase in migration output, got: %s", migrateStdout.String())
	}

	// Step 3: Verify vault is now v2
	vaultDataAfter, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault after migration: %v", err)
	}

	var vaultAfter struct {
		Metadata storage.VaultMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(vaultDataAfter, &vaultAfter); err != nil {
		t.Fatalf("Failed to parse vault after migration: %v", err)
	}

	if vaultAfter.Metadata.Version != 2 {
		t.Errorf("Expected vault version 2 after migration, got %d", vaultAfter.Metadata.Version)
	}

	if len(vaultAfter.Metadata.WrappedDEK) == 0 {
		t.Error("WrappedDEK should be present after migration")
	}
}

// TestCLI_VaultMigrateSkipsV2Vault verifies that `pass-cli vault migrate`
// correctly detects v2 vaults and skips migration.
func TestCLI_VaultMigrateSkipsV2Vault(t *testing.T) {
	// Create isolated test directory
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	configPath := filepath.Join(tempDir, "config.yml")

	// Write config file
	configContent := "vault_path: " + vaultPath + "\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	testPassword := "Test123!@#Password"

	// Create v2 vault (default)
	initInput := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	initCmd := exec.Command(binaryPath, "init")
	initCmd.Stdin = strings.NewReader(initInput)
	initCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+configPath)

	var initStdout, initStderr bytes.Buffer
	initCmd.Stdout = &initStdout
	initCmd.Stderr = &initStderr

	if err := initCmd.Run(); err != nil {
		t.Fatalf("init failed: %v\nStdout: %s\nStderr: %s", err, initStdout.String(), initStderr.String())
	}

	// Run migration on v2 vault - should be a no-op
	migrateCmd := exec.Command(binaryPath, "vault", "migrate")
	migrateCmd.Stdin = strings.NewReader("") // No input needed
	migrateCmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+configPath)

	var migrateStdout, migrateStderr bytes.Buffer
	migrateCmd.Stdout = &migrateStdout
	migrateCmd.Stderr = &migrateStderr

	if err := migrateCmd.Run(); err != nil {
		t.Fatalf("vault migrate failed: %v\nStdout: %s\nStderr: %s", err, migrateStdout.String(), migrateStderr.String())
	}

	// Verify message indicates already v2
	if !strings.Contains(migrateStdout.String(), "already") {
		t.Errorf("Expected 'already v2' message, got: %s", migrateStdout.String())
	}
}
