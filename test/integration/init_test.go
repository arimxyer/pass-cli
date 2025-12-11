//go:build integration
package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"pass-cli/internal/crypto"
	"pass-cli/internal/keychain"
	"pass-cli/internal/storage"
	"pass-cli/internal/vault"
	"pass-cli/test/helpers"
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

	testPassword := "Test123!@#Password"
	input := helpers.BuildInitStdin(helpers.DefaultInitOptions(testPassword))

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

	testPassword := "Test123!@#Password"
	input := helpers.BuildInitStdinNoRecovery(testPassword, false)

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
	initInput := helpers.BuildInitStdinNoRecovery(testPassword, false)
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
	vaultDataBefore, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault file: %v", err)
	}
	var vaultBefore struct {
		Metadata storage.VaultMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(vaultDataBefore, &vaultBefore); err != nil {
		t.Fatalf("Failed to unmarshal vault: %v", err)
	}
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
	initInput := helpers.BuildInitStdin(helpers.DefaultInitOptions(testPassword))
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

// T015: Integration test: init with recovery creates v2 vault
func TestInitWithRecoveryCreatesV2Vault(t *testing.T) {
	vaultPath := helpers.SetupTestVaultWithName(t, "init-recovery-test")
	// Cleanup is automatic via t.Cleanup()

	// Create vault service
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize with recovery enabled
	password := []byte("Test123!@#Password")
	_, err = vs.InitializeWithRecovery(password, false, "", "", nil)
	if err != nil {
		t.Fatalf("InitializeWithRecovery() error = %v", err)
	}

	// Load raw vault file to check version
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

	// Verify version is 2
	if encryptedVault.Metadata.Version != 2 {
		t.Errorf("Vault version = %d, want 2", encryptedVault.Metadata.Version)
	}

	// Verify WrappedDEK fields are present
	if len(encryptedVault.Metadata.WrappedDEK) != 48 {
		t.Errorf("WrappedDEK length = %d, want 48", len(encryptedVault.Metadata.WrappedDEK))
	}
	if len(encryptedVault.Metadata.WrappedDEKNonce) != 12 {
		t.Errorf("WrappedDEKNonce length = %d, want 12", len(encryptedVault.Metadata.WrappedDEKNonce))
	}

	// Verify recovery metadata has version "2"
	meta, err := vault.LoadMetadata(vaultPath)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	if meta.Recovery == nil {
		t.Fatal("Recovery metadata is nil")
	}
	if meta.Recovery.Version != "2" {
		t.Errorf("Recovery.Version = %q, want \"2\"", meta.Recovery.Version)
	}
}

// T016: Integration test: init with --no-recovery creates v1 vault
func TestInitWithNoRecoveryCreatesV1Vault(t *testing.T) {
	vaultPath := helpers.SetupTestVaultWithName(t, "init-norecovery-test")
	// Cleanup is automatic via t.Cleanup()

	// Create vault service
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize WITHOUT recovery (standard init)
	password := []byte("Test123!@#Password")
	err = vs.Initialize(password, false, "", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Load raw vault file to check version
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

	// Verify version is 1 (legacy)
	if encryptedVault.Metadata.Version != 1 {
		t.Errorf("Vault version = %d, want 1", encryptedVault.Metadata.Version)
	}

	// Verify WrappedDEK fields are NOT present (v1 vault)
	if len(encryptedVault.Metadata.WrappedDEK) != 0 {
		t.Errorf("WrappedDEK should be empty for v1 vault, got length %d", len(encryptedVault.Metadata.WrappedDEK))
	}
}

// T017: Unit test: VaultMetadata v2 serialization
func TestVaultMetadataV2Serialization(t *testing.T) {
	// Create v2 metadata with wrapped DEK fields
	wrappedDEK := make([]byte, 48) // 32-byte DEK + 16-byte tag
	wrappedDEKNonce := make([]byte, 12)
	salt := make([]byte, 32)

	// Fill with test data
	for i := range wrappedDEK {
		wrappedDEK[i] = byte(i)
	}
	for i := range wrappedDEKNonce {
		wrappedDEKNonce[i] = byte(i + 100)
	}
	for i := range salt {
		salt[i] = byte(i + 200)
	}

	metadata := storage.VaultMetadata{
		Version:         2,
		Salt:            salt,
		Iterations:      crypto.DefaultIterations,
		WrappedDEK:      wrappedDEK,
		WrappedDEKNonce: wrappedDEKNonce,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	// Deserialize back
	var decoded storage.VaultMetadata
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	// Verify fields
	if decoded.Version != 2 {
		t.Errorf("Version = %d, want 2", decoded.Version)
	}
	if len(decoded.WrappedDEK) != 48 {
		t.Errorf("WrappedDEK length = %d, want 48", len(decoded.WrappedDEK))
	}
	if len(decoded.WrappedDEKNonce) != 12 {
		t.Errorf("WrappedDEKNonce length = %d, want 12", len(decoded.WrappedDEKNonce))
	}

	// Verify data integrity
	for i := range wrappedDEK {
		if decoded.WrappedDEK[i] != wrappedDEK[i] {
			t.Errorf("WrappedDEK[%d] = %d, want %d", i, decoded.WrappedDEK[i], wrappedDEK[i])
		}
	}
}

// Test v1 metadata omits wrapped DEK fields
func TestVaultMetadataV1OmitsWrappedDEK(t *testing.T) {
	metadata := storage.VaultMetadata{
		Version:    1,
		Salt:       make([]byte, 32),
		Iterations: crypto.DefaultIterations,
		// WrappedDEK and WrappedDEKNonce intentionally left empty
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	// Check JSON doesn't contain wrapped_dek fields (omitempty)
	jsonStr := string(jsonData)
	if contains(jsonStr, "wrapped_dek") {
		t.Error("V1 metadata JSON should not contain wrapped_dek (omitempty)")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

// T026: Integration test: unlock v2 vault with correct password
func TestUnlockV2VaultWithCorrectPassword(t *testing.T) {
	vaultPath := helpers.SetupTestVaultWithName(t, "unlock-v2-correct-pwd")
	// Cleanup is automatic via t.Cleanup()
	passwordStr := "Test123!@#Password"

	// Create v2 vault with recovery enabled
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Make a fresh copy for initialization (will be cleared)
	initPassword := []byte(passwordStr)
	_, err = vs.InitializeWithRecovery(initPassword, false, "", "", nil)
	if err != nil {
		t.Fatalf("InitializeWithRecovery() error = %v", err)
	}

	// Verify vault is v2
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

	if encryptedVault.Metadata.Version != 2 {
		t.Fatalf("Expected v2 vault, got version %d", encryptedVault.Metadata.Version)
	}

	// Create new vault service to simulate fresh unlock
	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create second vault service: %v", err)
	}

	// Unlock with correct password (make a fresh copy from string)
	unlockPassword := []byte(passwordStr)
	err = vs2.Unlock(unlockPassword)
	if err != nil {
		t.Fatalf("Unlock() with correct password should succeed, got error: %v", err)
	}

	// Verify vault is unlocked
	if !vs2.IsUnlocked() {
		t.Error("Vault should be unlocked after successful Unlock()")
	}
}

// T027: Integration test: unlock v2 vault with wrong password fails
func TestUnlockV2VaultWithWrongPassword(t *testing.T) {
	vaultPath := helpers.SetupTestVaultWithName(t, "unlock-v2-wrong-pwd")
	// Cleanup is automatic via t.Cleanup()

	// Create v2 vault with recovery enabled
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	initPassword := []byte("Test123!@#Password")
	_, err = vs.InitializeWithRecovery(initPassword, false, "", "", nil)
	if err != nil {
		t.Fatalf("InitializeWithRecovery() error = %v", err)
	}

	// Create new vault service to simulate fresh unlock
	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create second vault service: %v", err)
	}

	// Unlock with wrong password
	wrongPassword := []byte("Wrong123!@#Password")
	err = vs2.Unlock(wrongPassword)
	if err == nil {
		t.Fatal("Unlock() with wrong password should fail")
	}

	// Verify vault is NOT unlocked
	if vs2.IsUnlocked() {
		t.Error("Vault should NOT be unlocked after failed Unlock()")
	}
}

// T028: Integration test: unlock v1 vault still works (backward compat)
func TestUnlockV1VaultBackwardCompatibility(t *testing.T) {
	vaultPath := helpers.SetupTestVaultWithName(t, "unlock-v1-compat")
	// Cleanup is automatic via t.Cleanup()
	passwordStr := "Test123!@#Password"

	// Create v1 vault (without recovery)
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	initPassword := []byte(passwordStr)
	err = vs.Initialize(initPassword, false, "", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Verify vault is v1
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

	if encryptedVault.Metadata.Version != 1 {
		t.Fatalf("Expected v1 vault, got version %d", encryptedVault.Metadata.Version)
	}

	// Create new vault service to simulate fresh unlock
	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create second vault service: %v", err)
	}

	// Unlock with correct password (make a fresh copy from string)
	unlockPassword := []byte(passwordStr)
	err = vs2.Unlock(unlockPassword)
	if err != nil {
		t.Fatalf("Unlock() v1 vault with correct password should succeed, got error: %v", err)
	}

	// Verify vault is unlocked
	if !vs2.IsUnlocked() {
		t.Error("V1 vault should be unlocked after successful Unlock()")
	}
}

// T028.1: Integration test: unlock with corrupted/missing WrappedDEK metadata fails gracefully
func TestUnlockV2VaultCorruptedMetadataFailsGracefully(t *testing.T) {
	vaultPath := helpers.SetupTestVaultWithName(t, "unlock-v2-corrupt")
	// Cleanup is automatic via t.Cleanup()
	passwordStr := "Test123!@#Password"

	// Create v2 vault with recovery enabled
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	initPassword := []byte(passwordStr)
	_, err = vs.InitializeWithRecovery(initPassword, false, "", "", nil)
	if err != nil {
		t.Fatalf("InitializeWithRecovery() error = %v", err)
	}

	// Read and corrupt the vault file
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault file: %v", err)
	}

	var encryptedVault storage.EncryptedVault
	if err := json.Unmarshal(vaultData, &encryptedVault); err != nil {
		t.Fatalf("Failed to parse vault file: %v", err)
	}

	// Corrupt the WrappedDEK by truncating it
	encryptedVault.Metadata.WrappedDEK = encryptedVault.Metadata.WrappedDEK[:16] // Should be 48 bytes

	// Write corrupted vault back
	corruptedData, err := json.Marshal(encryptedVault)
	if err != nil {
		t.Fatalf("Failed to marshal corrupted vault: %v", err)
	}
	if err := os.WriteFile(vaultPath, corruptedData, 0600); err != nil {
		t.Fatalf("Failed to write corrupted vault: %v", err)
	}

	// Create new vault service to simulate fresh unlock
	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create second vault service: %v", err)
	}

	// Unlock should fail gracefully (not panic)
	unlockPassword := []byte(passwordStr)
	err = vs2.Unlock(unlockPassword)
	if err == nil {
		t.Fatal("Unlock() with corrupted WrappedDEK should fail")
	}

	// Verify error message doesn't leak key material
	errMsg := err.Error()
	if len(errMsg) > 200 {
		t.Error("Error message should be concise, not leak large amounts of data")
	}
}

// Helper test: verify both v1 and v2 can add and retrieve credentials
func TestV1AndV2VaultsCanManageCredentials(t *testing.T) {
	passwordStr := "Test123!@#Password"

	testCases := []struct {
		name     string
		initFunc func(vs *vault.VaultService, password []byte) error
		version  int
	}{
		{
			name: "v1_vault",
			initFunc: func(vs *vault.VaultService, password []byte) error {
				return vs.Initialize(password, false, "", "")
			},
			version: 1,
		},
		{
			name: "v2_vault",
			initFunc: func(vs *vault.VaultService, password []byte) error {
				_, err := vs.InitializeWithRecovery(password, false, "", "", nil)
				return err
			},
			version: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vaultPath := helpers.SetupTestVaultWithName(t, "cred-mgmt-"+tc.name)
			// Cleanup is automatic via t.Cleanup()

			// Create vault
			vs, err := vault.New(vaultPath)
			if err != nil {
				t.Fatalf("Failed to create vault service: %v", err)
			}

			initPassword := []byte(passwordStr)
			if err := tc.initFunc(vs, initPassword); err != nil {
				t.Fatalf("Init error = %v", err)
			}

			// Re-create vault service and unlock
			vs2, err := vault.New(vaultPath)
			if err != nil {
				t.Fatalf("Failed to create second vault service: %v", err)
			}

			unlockPassword := []byte(passwordStr)
			if err := vs2.Unlock(unlockPassword); err != nil {
				t.Fatalf("Unlock() error = %v", err)
			}

			// Add a credential
			credPassword := []byte("secret123")
			if err := vs2.AddCredential("test-service", "testuser", credPassword, "", "", ""); err != nil {
				t.Fatalf("AddCredential() error = %v", err)
			}

			// Lock and unlock again
			vs2.Lock()

			vs3, err := vault.New(vaultPath)
			if err != nil {
				t.Fatalf("Failed to create third vault service: %v", err)
			}

			unlockPassword2 := []byte(passwordStr)
			if err := vs3.Unlock(unlockPassword2); err != nil {
				t.Fatalf("Second Unlock() error = %v", err)
			}

			// Retrieve credential
			cred, err := vs3.GetCredential("test-service", false)
			if err != nil {
				t.Fatalf("GetCredential() error = %v", err)
			}

			if cred.Username != "testuser" {
				t.Errorf("Username = %q, want %q", cred.Username, "testuser")
			}
			if string(cred.Password) != "secret123" {
				t.Errorf("Password mismatch")
			}

			// Clear credential password after checking
			crypto.ClearBytes(cred.Password)
		})
	}
}

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
	ks := keychain.New("")
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
	ks := keychain.New("")
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
