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

	"pass-cli/internal/keychain"
	"pass-cli/internal/vault"
)

// T014: Integration test for corrupted metadata fallback
// Tests that VaultService falls back to self-discovery when metadata is corrupted
func TestIntegration_CorruptedMetadataFallback(t *testing.T) {
	// Check if keychain is available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	testPassword := "CorruptTest-Pass@123"
	vaultDir := filepath.Join(testDir, "corrupt-metadata-vault")
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
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init with audit failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify metadata file created
	metaPath := vault.MetadataPath(vaultPath)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata file was not created")
	}

	// Corrupt metadata file
	if err := os.WriteFile(metaPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to corrupt metadata: %v", err)
	}

	// Run keychain status command (should use fallback self-discovery)
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "keychain", "status")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Keychain status failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify warning about corrupted metadata
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Warning") && !strings.Contains(stderrOutput, "metadata") {
		t.Error("Expected warning about corrupted metadata in stderr")
	}

	// Verify fallback self-discovery worked (audit.log should still be used)
	if _, err := os.Stat(auditLogPath); err != nil {
		t.Error("Audit log should exist for fallback self-discovery")
	}

	t.Logf("✓ Fallback self-discovery succeeded despite corrupted metadata")
}

// T015: Integration test for multiple vaults in same directory
// Tests that metadata correctly identifies the right vault when multiple vaults exist
func TestIntegration_MultipleVaultsInDirectory(t *testing.T) {
	// Check if keychain is available
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available - skipping test")
	}

	testPassword1 := "Vault1-Pass@123"
	testPassword2 := "Vault2-Pass@123"
	vaultDir := filepath.Join(testDir, "multi-vault-dir")
	vault1Path := filepath.Join(vaultDir, "vault1.enc")
	vault2Path := filepath.Join(vaultDir, "vault2.enc")

	// Ensure clean state
	defer cleanupKeychain(t, ks)
	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	// Create directory
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Setup config for vault1
	testConfigPath1, cleanup1 := setupTestVaultConfig(t, vault1Path)
	defer cleanup1()

	// Initialize vault1 with audit
	input1 := testPassword1 + "\n" + testPassword1 + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath1)
	cmd.Stdin = strings.NewReader(input1)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init vault1 failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Setup config for vault2
	testConfigPath2, cleanup2 := setupTestVaultConfig(t, vault2Path)
	defer cleanup2()

	// Initialize vault2 without audit
	input2 := testPassword2 + "\n" + testPassword2 + "\n" + "n\n" + "n\n"
	cmd = exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath2)
	cmd.Stdin = strings.NewReader(input2)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init vault2 failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify vault1.meta exists (but not vault2.meta since audit wasn't enabled)
	meta1Path := vault.MetadataPath(vault1Path)
	if _, err := os.Stat(meta1Path); os.IsNotExist(err) {
		t.Fatal("Metadata file should exist for vault1")
	}

	// Verify the metadata filename correctly identifies vault1
	// Metadata naming: <vault-path>.meta.json (e.g., vault1.enc.meta.json)
	expectedFilename := filepath.Base(vault1Path) + ".meta.json"
	actualFilename := filepath.Base(meta1Path)
	if actualFilename != expectedFilename {
		t.Errorf("Metadata filename should be %q, got %q", expectedFilename, actualFilename)
	}

	// Verify vault2 also has metadata (created for all vaults to track keychain status)
	// but with audit_enabled=false
	meta2Path := vault.MetadataPath(vault2Path)
	if _, err := os.Stat(meta2Path); os.IsNotExist(err) {
		t.Error("vault2 should have metadata (metadata now created for all vaults)")
	}

	// Verify vault2 metadata shows audit disabled
	data2, err := os.ReadFile(meta2Path)
	if err != nil {
		t.Fatalf("Failed to read vault2 metadata: %v", err)
	}
	var meta2 vault.Metadata
	if err := json.Unmarshal(data2, &meta2); err != nil {
		t.Fatalf("Failed to parse vault2 metadata: %v", err)
	}
	if meta2.AuditEnabled {
		t.Error("vault2 metadata should show audit_enabled=false")
	}

	// Read and verify metadata content is valid
	data, err := os.ReadFile(meta1Path)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}
	var meta vault.Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}
	if !meta.AuditEnabled {
		t.Error("Metadata should show audit_enabled=true")
	}

	t.Logf("✓ Metadata correctly identifies vault1 among multiple vaults (via filename: %s)", actualFilename)
}

// T021: Integration test for automatic metadata creation on vault unlock with audit
// Tests that metadata is created when unlocking a vault that has audit enabled but no metadata file
func TestIntegration_AutoMetadataCreationOnUnlock(t *testing.T) {
	testPassword := "UnlockTest-Pass@123"
	vaultDir := filepath.Join(testDir, "unlock-metadata-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	// Create vault with audit enabled
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify metadata created by init
	metaPath := vault.MetadataPath(vaultPath)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata should be created by init audit logging (enabled by default)")
	}

	// Delete metadata to simulate old vault
	if err := os.Remove(metaPath); err != nil {
		t.Fatalf("Failed to delete metadata: %v", err)
	}

	// Unlock vault (should recreate metadata since audit is enabled in vault)
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "list")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(testPassword + "\n")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("List command failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify metadata was recreated
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("Metadata should be recreated on unlock when audit enabled in vault")
	}

	t.Logf("✓ Metadata automatically created on unlock")
}

// T022: Integration test for no metadata creation when audit disabled
// Tests that metadata is NOT created for vaults without audit logging
func TestIntegration_NoMetadataWhenAuditDisabled(t *testing.T) {
	testPassword := "NoAudit-Pass@123"
	vaultDir := filepath.Join(testDir, "no-audit-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	// Create vault WITHOUT audit
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify metadata IS created (but with audit_enabled=false)
	// Metadata is now created for all vaults to track keychain status
	metaPath := vault.MetadataPath(vaultPath)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("Metadata should be created even when audit disabled (to track keychain status)")
	}

	// Verify metadata shows audit disabled
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}
	var meta vault.Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}
	if meta.AuditEnabled {
		t.Error("Metadata should show audit_enabled=false when audit disabled")
	}

	// Unlock vault (metadata should persist)
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "list")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(testPassword + "\n")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("List command failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify metadata still exists and shows audit disabled
	data, err = os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read metadata after unlock: %v", err)
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("Failed to parse metadata after unlock: %v", err)
	}
	if meta.AuditEnabled {
		t.Error("Metadata should still show audit_enabled=false after unlock")
	}

	t.Logf("✓ Metadata created with audit_enabled=false for audit-disabled vault")
}

// T023: Integration test for metadata creation via init audit logging (enabled by default)
// Tests that init command creates metadata when audit logging (enabled by default) flag is used
func TestIntegration_MetadataCreatedByInit(t *testing.T) {
	testPassword := "InitAudit-Pass@123"
	vaultDir := filepath.Join(testDir, "init-audit-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Init with audit logging (enabled by default)
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init with audit failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify metadata created
	metaPath := vault.MetadataPath(vaultPath)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata file should be created by init audit logging (enabled by default)")
	}

	// Verify metadata content
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	// Check for required fields per data-model.md
	if !strings.Contains(string(data), "audit_enabled") {
		t.Error("Metadata should contain audit_enabled field")
	}

	if !strings.Contains(string(data), "keychain_enabled") {
		t.Error("Metadata should contain keychain_enabled field")
	}

	if !strings.Contains(string(data), "version") {
		t.Error("Metadata should contain version field")
	}

	t.Logf("✓ Metadata created by init audit logging (enabled by default)")
}

// T024: Integration test for metadata update when audit settings change
// Tests that metadata is synchronized when vault audit config changes
func TestIntegration_MetadataUpdateOnAuditChange(t *testing.T) {
	testPassword := "ChangeAudit-Pass@123"
	vaultDir := filepath.Join(testDir, "change-audit-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Create vault with audit
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Read initial metadata
	metaPath := vault.MetadataPath(vaultPath)
	initialData, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read initial metadata: %v", err)
	}

	if !strings.Contains(string(initialData), `"audit_enabled": true`) {
		t.Error("Initial metadata should have audit enabled")
	}

	// Manually corrupt metadata to simulate mismatch (set audit_enabled to false)
	corruptedMeta := strings.Replace(string(initialData), `"audit_enabled": true`, `"audit_enabled": false`, 1)
	if err := os.WriteFile(metaPath, []byte(corruptedMeta), 0644); err != nil {
		t.Fatalf("Failed to write corrupted metadata: %v", err)
	}

	// Unlock vault (should detect mismatch and update metadata)
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "list")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(testPassword + "\n")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	// Read updated metadata
	updatedData, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read updated metadata: %v", err)
	}

	// Verify metadata was corrected (vault settings take precedence)
	if !strings.Contains(string(updatedData), `"audit_enabled": true`) {
		t.Error("Metadata should be updated to match vault audit settings (audit_enabled: true)")
	}

	t.Logf("✓ Metadata updated when mismatch detected")
}

// T025: Integration test for backward compatibility with old vaults
// Tests that vaults created before metadata feature still work without breaking changes
func TestIntegration_BackwardCompatibilityOldVaults(t *testing.T) {
	testPassword := "OldVault-Pass@123"
	vaultDir := filepath.Join(testDir, "old-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Create vault without metadata (simulate old vault)
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Delete metadata if created (to truly simulate old vault)
	metaPath := vault.MetadataPath(vaultPath)
	_ = os.Remove(metaPath) // Best effort cleanup to simulate old vault

	// Try to use vault (should work without metadata)
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "list")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(testPassword + "\n")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Old vault should work without metadata: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Add credential to verify full functionality
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "add", "test-service", "--username", "testuser", "--password", "testpass")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(testPassword + "\n")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Add command should work without metadata: %v", err)
	}

	// List credentials
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "list")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(testPassword + "\n")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "test-service") {
		t.Error("Should be able to list credentials in old vault")
	}

	t.Logf("✓ Old vaults work without metadata (backward compatible)")
}

// T032: Integration test for metadata deleted, fallback self-discovery succeeds
// Tests that VaultService uses fallback when metadata file is deleted
func TestIntegration_MetadataDeletedFallback(t *testing.T) {
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available")
	}

	testPassword := "DeletedMeta-Pass@123"
	vaultDir := filepath.Join(testDir, "deleted-meta-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	auditLogPath := filepath.Join(vaultDir, "audit.log")

	defer cleanupKeychain(t, ks)
	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Create vault with audit
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify audit.log exists
	if _, err := os.Stat(auditLogPath); os.IsNotExist(err) {
		t.Fatal("Audit log should exist")
	}

	// Delete metadata file
	metaPath := vault.MetadataPath(vaultPath)
	if err := os.Remove(metaPath); err != nil {
		t.Fatalf("Failed to delete metadata: %v", err)
	}

	// Run keychain status (should use fallback self-discovery)
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "keychain", "status")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Keychain status should work with fallback: %v", err)
	}

	// Verify audit entry written via fallback
	data, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	if len(data) == 0 {
		t.Error("Audit log should have entries from fallback")
	}

	t.Logf("✓ Fallback self-discovery succeeded when metadata deleted")
}

// T033: Integration test for metadata corrupted (invalid JSON), fallback succeeds
// Already covered in T014, this is an alias/duplicate test for US3 completeness
func TestIntegration_CorruptedMetadataFallbackUS3(t *testing.T) {
	// This is the same as T014 but categorized under US3
	// See TestIntegration_CorruptedMetadataFallback for implementation
	t.Skip("Covered by T014 (TestIntegration_CorruptedMetadataFallback)")
}

// T034: Integration test for audit.log exists but no metadata, best-effort logging
// Tests that system finds audit.log via self-discovery when no metadata exists
func TestIntegration_AuditLogExistsNoMetadata(t *testing.T) {
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available")
	}

	testPassword := "NoMeta-Pass@123"
	vaultDir := filepath.Join(testDir, "no-meta-audit-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	auditLogPath := filepath.Join(vaultDir, "audit.log")

	defer cleanupKeychain(t, ks)
	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Create vault with audit
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Delete metadata but keep audit.log
	metaPath := vault.MetadataPath(vaultPath)
	if err := os.Remove(metaPath); err != nil {
		t.Fatalf("Failed to delete metadata: %v", err)
	}

	// Verify audit.log still exists
	if _, err := os.Stat(auditLogPath); os.IsNotExist(err) {
		t.Fatal("Audit log should still exist")
	}

	// Get initial audit log size
	initialData, _ := os.ReadFile(auditLogPath)
	initialSize := len(initialData)

	// Run command (should discover audit.log and use it)
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "keychain", "status")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command should work with audit.log present: %v", err)
	}

	// Verify audit log grew (new entry written)
	finalData, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	if len(finalData) <= initialSize {
		t.Error("Audit log should have new entries via best-effort logging")
	}

	t.Logf("✓ Best-effort logging succeeded with audit.log but no metadata")
}

// T035: Integration test for metadata indicates audit but audit.log missing, creates new log
// Tests graceful handling when metadata says audit enabled but log file is missing
func TestIntegration_MetadataWithMissingAuditLog(t *testing.T) {
	ks := keychain.New()
	if !ks.IsAvailable() {
		t.Skip("System keychain not available")
	}

	testPassword := "MissingLog-Pass@123"
	vaultDir := filepath.Join(testDir, "missing-log-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	auditLogPath := filepath.Join(vaultDir, "audit.log")

	defer cleanupKeychain(t, ks)
	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Create vault with audit
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Delete audit.log but keep metadata
	if err := os.Remove(auditLogPath); err != nil {
		t.Fatalf("Failed to delete audit log: %v", err)
	}

	// Verify metadata still exists
	metaPath := vault.MetadataPath(vaultPath)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata should still exist")
	}

	// Run command (should create new audit.log per FR-013)
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "keychain", "status")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command should work even with missing audit.log: %v", err)
	}

	// Verify audit.log was recreated
	if _, err := os.Stat(auditLogPath); os.IsNotExist(err) {
		t.Error("Audit log should be recreated when missing")
	}

	// Verify it has entries
	data, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read recreated audit log: %v", err)
	}

	if len(data) == 0 {
		t.Error("Recreated audit log should have entries")
	}

	t.Logf("✓ New audit log created when missing (FR-013)")
}

// T036: Integration test for unknown metadata version number, logs warning and attempts parsing
// Tests forward compatibility with future metadata versions
func TestIntegration_UnknownMetadataVersion(t *testing.T) {
	testPassword := "UnknownVer-Pass@123"
	vaultDir := filepath.Join(testDir, "unknown-version-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	defer func() { _ = os.RemoveAll(vaultDir) }() // Best effort cleanup

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Setup config with vault_path
	testConfigPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Create vault with audit
	input := testPassword + "\n" + testPassword + "\n" + "n\n" + "n\n"
	cmd := exec.Command(binaryPath, "init")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Read metadata and change version to 99
	metaPath := vault.MetadataPath(vaultPath)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	// Replace version "1.0" with version "99.0"
	modifiedData := strings.Replace(string(data), `"version": "1.0"`, `"version": "99.0"`, 1)
	if !strings.Contains(modifiedData, `"version": "99.0"`) {
		t.Fatalf("Failed to modify version in metadata (original: %s)", string(data))
	}
	if err := os.WriteFile(metaPath, []byte(modifiedData), 0644); err != nil {
		t.Fatalf("Failed to write modified metadata: %v", err)
	}

	// Run command (should log warning but continue)
	// Reuse parent testConfigPath from deferred setup
	cmd = exec.Command(binaryPath, "keychain", "status")
	cmd.Env = append(os.Environ(), "PASS_CLI_TEST=1", "PASS_CLI_CONFIG="+testConfigPath)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command should work despite unknown version: %v", err)
	}

	// Verify warning message in stderr
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Warning") && !strings.Contains(stderrOutput, "version") {
		t.Error("Expected warning about unknown metadata version")
	}

	t.Logf("✓ Unknown version handled gracefully (FR-017)")
}
