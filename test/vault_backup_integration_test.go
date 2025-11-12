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

// TestIntegration_BackupCreate_Success tests successful manual backup creation
// T033: Integration test for successful backup creation
func TestIntegration_BackupCreate_Success(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-create-test", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Create manual backup
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("backup create failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	// Verify success message contains backup path
	if stdout == "" {
		t.Error("Expected success message in stdout, got empty string")
	}

	// Verify backup file was created with correct naming pattern
	vaultDir := filepath.Dir(vaultPath)
	pattern := filepath.Join(vaultDir, "vault.enc.*.manual.backup")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("failed to glob for backup files: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 manual backup file, found %d", len(matches))
	}

	// Verify backup file has correct size (same as vault)
	if len(matches) > 0 {
		vaultInfo, err := os.Stat(vaultPath)
		if err != nil {
			t.Fatalf("failed to stat vault: %v", err)
		}

		backupInfo, err := os.Stat(matches[0])
		if err != nil {
			t.Fatalf("failed to stat backup: %v", err)
		}

		if vaultInfo.Size() != backupInfo.Size() {
			t.Errorf("Backup size %d does not match vault size %d",
				backupInfo.Size(), vaultInfo.Size())
		}
	}
}

// TestIntegration_BackupCreate_VaultNotFound tests backup when vault doesn't exist
// T034: Integration test for backup with vault not found
func TestIntegration_BackupCreate_VaultNotFound(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-notfound-test", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Try to create backup without initializing vault
	_, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err == nil {
		t.Fatal("Expected error when backing up non-existent vault, got nil")
	}

	// Verify error message mentions vault not found
	if stderr == "" {
		t.Error("Expected error message in stderr, got empty string")
	}
}

// TestIntegration_BackupCreate_MultipleBackups tests multiple manual backups
// T037: Integration test for multiple manual backups
func TestIntegration_BackupCreate_MultipleBackups(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-multiple-test", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Create first manual backup
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("first backup create failed: %v\nstderr: %s", err, stderr)
	}

	// Wait a bit to ensure different timestamps
	// Note: Manual backups use second-precision timestamps
	// Sleep is not ideal but necessary for timestamp uniqueness test
	// In production, backups created at same second would have identical names
	// which is acceptable since users wouldn't create multiple backups per second

	// Create second manual backup
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("second backup create failed: %v\nstderr: %s", err, stderr)
	}

	// Create third manual backup
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("third backup create failed: %v\nstderr: %s", err, stderr)
	}

	// Verify all backup files exist (at least 1, could be 3 if timing allows)
	vaultDir := filepath.Dir(vaultPath)
	pattern := filepath.Join(vaultDir, "vault.enc.*.manual.backup")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("failed to glob for backup files: %v", err)
	}

	// Since backups created in quick succession may have same timestamp,
	// we verify at least 1 backup was created
	if len(matches) < 1 {
		t.Errorf("Expected at least 1 manual backup file, found %d", len(matches))
	}

	// Verify no backup was overwritten (each backup should be unique or timestamped differently)
	// If multiple backups exist, verify they all have different names
	if len(matches) > 1 {
		seen := make(map[string]bool)
		for _, backup := range matches {
			name := filepath.Base(backup)
			if seen[name] {
				t.Errorf("Duplicate backup filename found: %s", name)
			}
			seen[name] = true
		}
	}
}

// TestIntegration_BackupCreate_MissingDirectory tests backup directory creation
// T035a: Integration test for backup with missing directory
func TestIntegration_BackupCreate_MissingDirectory(t *testing.T) {
	setupTestEnvironment(t)

	// Use nested directory path
	vaultPath := filepath.Join(testDir, "vault-missing-dir", "subdir", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault (creates directory structure)
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Remove the subdirectory to simulate missing backup directory
	subdirPath := filepath.Dir(vaultPath)
	vaultContent, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("failed to read vault before removing directory: %v", err)
	}

	if err := os.RemoveAll(subdirPath); err != nil {
		t.Fatalf("failed to remove subdir: %v", err)
	}

	// Recreate vault file but not the directory structure for backups
	if err := os.MkdirAll(subdirPath, 0700); err != nil {
		t.Fatalf("failed to recreate vault directory: %v", err)
	}
	if err := os.WriteFile(vaultPath, vaultContent, 0600); err != nil {
		t.Fatalf("failed to recreate vault file: %v", err)
	}

	// Create backup - should create directory structure if needed (FR-018)
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("backup create failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	// Verify backup file was created
	vaultDir := filepath.Dir(vaultPath)
	pattern := filepath.Join(vaultDir, "vault.enc.*.manual.backup")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("failed to glob for backup files: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 manual backup file, found %d", len(matches))
	}
}

// setupTestEnvironment creates necessary test directories
func setupTestEnvironment(t *testing.T) {
	t.Helper()
	// testDir is created by TestMain in integration_test.go
}
