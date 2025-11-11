//go:build integration

package test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIntegration_BackupRestore_Basic tests the full backup and restore workflow
// T013: Integration test for basic restore
func TestIntegration_BackupRestore_Basic(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-restore-test", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault (password must be 12+ characters)
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Add test credential
	_, stderr, err = runCommandWithInput(t, "TestPassword123!\ntestuser\ntestpass123\n\n\n", "--config", configPath, "add", "test-service")
	if err != nil {
		t.Fatalf("add failed: %v\nstderr: %s", err, stderr)
	}

	// Create manual backup
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("backup create failed: %v\nstderr: %s", err, stderr)
	}

	// Corrupt the vault by writing garbage
	if err := os.WriteFile(vaultPath, []byte("corrupted data"), 0600); err != nil {
		t.Fatalf("failed to corrupt vault: %v", err)
	}

	// Restore from backup with --force to skip confirmation
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
	if err != nil {
		t.Fatalf("backup restore failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	// Verify vault can be unlocked and credential retrieved
	stdout, stderr, err = runCommandWithInput(t, "TestPassword123!\n", "--config", configPath, "get", "test-service", "--field", "username")
	if err != nil {
		t.Fatalf("get after restore failed: %v\nstderr: %s", err, stderr)
	}

	if stdout != "testuser\n" {
		t.Errorf("Expected username 'testuser', got %q", stdout)
	}
}

// TestIntegration_BackupRestore_NoBackups tests restore when no backups exist
// T014: Integration test for restore with no backups
func TestIntegration_BackupRestore_NoBackups(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-no-backup", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Remove any automatic backup that may have been created
	backupPath := vaultPath + ".backup"
	_ = os.Remove(backupPath)

	// Try to restore without any backup
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
	if err == nil {
		t.Fatal("Expected error when restoring without backups, got nil")
	}

	// Verify error message is helpful
	if stderr == "" {
		t.Error("Expected error message in stderr, got empty string")
	}
}

// TestIntegration_BackupRestore_CorruptedFallback tests fallback to next valid backup
// T015: Integration test for restore with corrupted backup
func TestIntegration_BackupRestore_CorruptedFallback(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-corrupted-fallback", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Add initial credential (state A)
	_, stderr, err = runCommandWithInput(t, "TestPassword123!\nolduser\noldpass123\n\n\n", "--config", configPath, "add", "test-service")
	if err != nil {
		t.Fatalf("add failed: %v\nstderr: %s", err, stderr)
	}

	// Create manual backup (captures state A with olduser)
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("backup create failed: %v\nstderr: %s", err, stderr)
	}

	// Update credential to different value (state B)
	_, stderr, err = runCommandWithInput(t, "TestPassword123!\nnewuser\nnewpass123\n\n\n", "--config", configPath, "update", "test-service")
	if err != nil {
		t.Fatalf("update failed: %v\nstderr: %s", err, stderr)
	}

	// Verify automatic backup was created (captures state B with newuser)
	automaticBackupPath := vaultPath + ".backup"
	if _, err := os.Stat(automaticBackupPath); os.IsNotExist(err) {
		t.Fatalf("automatic backup was not created at %s", automaticBackupPath)
	}

	// Corrupt the automatic backup (most recent)
	if err := os.WriteFile(automaticBackupPath, []byte("corrupted data"), 0600); err != nil {
		t.Fatalf("failed to corrupt automatic backup: %v", err)
	}

	// Restore from backup with --force (should fallback to manual backup)
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
	if err != nil {
		t.Fatalf("backup restore failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	// Verify vault was restored to state A (olduser), not state B (newuser)
	// This proves it fell back to the manual backup after detecting corruption
	stdout, stderr, err = runCommandWithInput(t, "TestPassword123!\n", "--config", configPath, "get", "test-service", "--field", "username")
	if err != nil {
		t.Fatalf("get after restore failed: %v\nstderr: %s", err, stderr)
	}

	if stdout != "olduser\n" {
		t.Errorf("Expected username 'olduser' (from manual backup fallback), got %q", stdout)
	}
}

// TestIntegration_BackupRestore_ConfirmationPrompt tests the confirmation workflow
// T016: Integration test for restore confirmation prompt
func TestIntegration_BackupRestore_ConfirmationPrompt(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-confirm-test", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Create manual backup
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("backup create failed: %v\nstderr: %s", err, stderr)
	}

	// Try to restore with 'n' response (cancel)
	stdout, stderr, err := runCommandWithInput(t, "n\n", "--config", configPath, "vault", "backup", "restore")
	if err == nil {
		t.Error("Expected error when cancelling restore, got nil")
	}

	// Verify cancellation message
	combinedOutput := stdout + stderr
	if combinedOutput == "" {
		t.Error("Expected cancellation message, got empty output")
	}
}

// TestIntegration_BackupRestore_ForceFlag tests --force flag skips confirmation
// T017: Integration test for restore with --force flag
func TestIntegration_BackupRestore_ForceFlag(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-force-test", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Create manual backup
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("backup create failed: %v\nstderr: %s", err, stderr)
	}

	// Restore with --force (no input needed)
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
	if err != nil {
		t.Fatalf("backup restore with --force failed: %v\nstderr: %s", err, stderr)
	}
}

// TestIntegration_BackupRestore_DryRun tests --dry-run flag
// T018: Integration test for restore with --dry-run flag
func TestIntegration_BackupRestore_DryRun(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-dryrun-test", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Create manual backup
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("backup create failed: %v\nstderr: %s", err, stderr)
	}

	// Get original vault mtime
	vaultInfo, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("failed to stat vault: %v", err)
	}
	originalModTime := vaultInfo.ModTime()

	// Run restore with --dry-run
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "restore", "--dry-run")
	if err != nil {
		t.Fatalf("backup restore --dry-run failed: %v\nstderr: %s", err, stderr)
	}

	// Verify output shows what would be restored
	if stdout == "" {
		t.Error("Expected dry-run output, got empty string")
	}

	// Verify vault file was not modified
	vaultInfo, err = os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("failed to stat vault after dry-run: %v", err)
	}

	if vaultInfo.ModTime() != originalModTime {
		t.Error("Vault file was modified during --dry-run (should be unchanged)")
	}
}

// setupTestEnvironment creates necessary test directories
func setupTestEnvironment(t *testing.T) {
	t.Helper()
	// testDir is created by TestMain in integration_test.go
}
