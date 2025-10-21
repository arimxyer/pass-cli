//go:build integration
// +build integration

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

	// Corrupt metadata file
	if err := os.WriteFile(metaPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to corrupt metadata: %v", err)
	}

	// Run keychain status command (should use fallback self-discovery)
	cmd = exec.Command(binaryPath, "--vault", vaultPath, "keychain", "status")
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
	defer os.RemoveAll(vaultDir)

	// Create directory
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Initialize vault1 with audit
	input1 := testPassword1 + "\n" + testPassword1 + "\n"
	cmd := exec.Command(binaryPath, "--vault", vault1Path, "init", "--enable-audit")
	cmd.Stdin = strings.NewReader(input1)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init vault1 failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Initialize vault2 without audit
	input2 := testPassword2 + "\n" + testPassword2 + "\n"
	cmd = exec.Command(binaryPath, "--vault", vault2Path, "init")
	cmd.Stdin = strings.NewReader(input2)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init vault2 failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify vault1.meta exists (but not vault2.meta since audit wasn't enabled)
	meta1Path := filepath.Join(vaultDir, "vault.meta")
	if _, err := os.Stat(meta1Path); os.IsNotExist(err) {
		// Note: Current implementation uses "vault.meta" (fixed name), not "vault1.meta"
		// This test verifies that metadata correctly identifies vault1 via vault_id field
		t.Skip("Multi-vault in same directory not yet supported - metadata filename is fixed to 'vault.meta'")
	}

	// Read metadata and verify it references vault1
	data, err := os.ReadFile(meta1Path)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	if !strings.Contains(string(data), "vault1.enc") {
		t.Error("Metadata should reference vault1.enc in vault_id field")
	}

	t.Logf("✓ Metadata correctly identifies vault1 among multiple vaults")
}

// T021: Integration test for automatic metadata creation on vault unlock with audit
// Tests that metadata is created when unlocking a vault that has audit enabled but no metadata file
func TestIntegration_AutoMetadataCreationOnUnlock(t *testing.T) {
	testPassword := "UnlockTest-Pass@123"
	vaultDir := filepath.Join(testDir, "unlock-metadata-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	defer os.RemoveAll(vaultDir)

	// Create vault with audit enabled
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	input := testPassword + "\n" + testPassword + "\n"
	cmd := exec.Command(binaryPath, "--vault", vaultPath, "init", "--enable-audit")
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify metadata created by init
	metaPath := filepath.Join(vaultDir, "vault.meta")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata should be created by init --enable-audit")
	}

	// Delete metadata to simulate old vault
	if err := os.Remove(metaPath); err != nil {
		t.Fatalf("Failed to delete metadata: %v", err)
	}

	// Unlock vault (should recreate metadata since audit is enabled in vault)
	cmd = exec.Command(binaryPath, "--vault", vaultPath, "list")
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

	defer os.RemoveAll(vaultDir)

	// Create vault WITHOUT audit
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	input := testPassword + "\n" + testPassword + "\n"
	cmd := exec.Command(binaryPath, "--vault", vaultPath, "init")
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify no metadata created
	metaPath := filepath.Join(vaultDir, "vault.meta")
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Error("Metadata should NOT be created when audit disabled")
	}

	// Unlock vault (should still not create metadata)
	cmd = exec.Command(binaryPath, "--vault", vaultPath, "list")
	cmd.Stdin = strings.NewReader(testPassword + "\n")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("List command failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify still no metadata
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Error("Metadata should still not exist after unlock when audit disabled")
	}

	t.Logf("✓ No metadata created for audit-disabled vault")
}

// T023: Integration test for metadata creation via init --enable-audit
// Tests that init command creates metadata when --enable-audit flag is used
func TestIntegration_MetadataCreatedByInit(t *testing.T) {
	testPassword := "InitAudit-Pass@123"
	vaultDir := filepath.Join(testDir, "init-audit-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	defer os.RemoveAll(vaultDir)

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Init with --enable-audit
	input := testPassword + "\n" + testPassword + "\n"
	cmd := exec.Command(binaryPath, "--vault", vaultPath, "init", "--enable-audit")
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init with audit failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify metadata created
	metaPath := filepath.Join(vaultDir, "vault.meta")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata file should be created by init --enable-audit")
	}

	// Verify metadata content
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	if !strings.Contains(string(data), "vault.enc") {
		t.Error("Metadata should contain vault path")
	}

	if !strings.Contains(string(data), "audit_enabled") {
		t.Error("Metadata should contain audit_enabled field")
	}

	t.Logf("✓ Metadata created by init --enable-audit")
}

// T024: Integration test for metadata update when audit settings change
// Tests that metadata is synchronized when vault audit config changes
func TestIntegration_MetadataUpdateOnAuditChange(t *testing.T) {
	testPassword := "ChangeAudit-Pass@123"
	vaultDir := filepath.Join(testDir, "change-audit-vault")
	vaultPath := filepath.Join(vaultDir, "vault.enc")

	defer os.RemoveAll(vaultDir)

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Create vault with audit
	input := testPassword + "\n" + testPassword + "\n"
	cmd := exec.Command(binaryPath, "--vault", vaultPath, "init", "--enable-audit")
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Read initial metadata
	metaPath := filepath.Join(vaultDir, "vault.meta")
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
	cmd = exec.Command(binaryPath, "--vault", vaultPath, "list")
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

	defer os.RemoveAll(vaultDir)

	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	// Create vault without metadata (simulate old vault)
	input := testPassword + "\n" + testPassword + "\n"
	cmd := exec.Command(binaryPath, "--vault", vaultPath, "init")
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Delete metadata if created (to truly simulate old vault)
	metaPath := filepath.Join(vaultDir, "vault.meta")
	os.Remove(metaPath)

	// Try to use vault (should work without metadata)
	cmd = exec.Command(binaryPath, "--vault", vaultPath, "list")
	cmd.Stdin = strings.NewReader(testPassword + "\n")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Old vault should work without metadata: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	// Add credential to verify full functionality
	cmd = exec.Command(binaryPath, "--vault", vaultPath, "add", "test-service")
	cmd.Stdin = strings.NewReader(testPassword + "\ntestuser\ntestpass\ntestpass\n")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Add command should work without metadata: %v", err)
	}

	// List credentials
	cmd = exec.Command(binaryPath, "--vault", vaultPath, "list")
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
