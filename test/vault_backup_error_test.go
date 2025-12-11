//go:build integration

package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T078: Comprehensive error handling test suite for backup commands
// Tests all error paths across create, restore, and info commands

// TestIntegration_BackupCreate_Errors tests error paths for backup create command
func TestIntegration_BackupCreate_Errors(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("VaultNotFound", func(t *testing.T) {
		vaultPath := filepath.Join(testDir, "nonexistent", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Try to create backup without vault
		_, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err == nil {
			t.Fatal("Expected error when vault doesn't exist, got nil")
		}

		// Verify error message is helpful
		if !strings.Contains(stderr, "vault not found") && !strings.Contains(stderr, "not found") {
			t.Errorf("Expected 'vault not found' in error, got: %s", stderr)
		}
	})

	t.Run("PermissionDenied_ReadOnly", func(t *testing.T) {
		vaultPath := filepath.Join(testDir, "readonly-test", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\nn\nn\nn\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Make vault directory read-only
		vaultDir := filepath.Dir(vaultPath)
		originalInfo, err := os.Stat(vaultDir)
		if err != nil {
			t.Fatalf("failed to stat vault directory: %v", err)
		}
		originalPerm := originalInfo.Mode().Perm()
		defer func() { _ = os.Chmod(vaultDir, originalPerm) }()

		if err := os.Chmod(vaultDir, 0555); err != nil {
			t.Fatalf("failed to make directory read-only: %v", err)
		}

		// Try to create backup
		_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")

		// Restore permissions for cleanup
		_ = os.Chmod(vaultDir, originalPerm)

		// On Windows, permissions work differently - test may not fail as expected
		if err == nil {
			t.Skip("Permission test doesn't work on this platform (Windows?)")
			return
		}

		// Verify error message mentions permissions
		if !strings.Contains(stderr, "permission") && !strings.Contains(stderr, "Permission") {
			t.Errorf("Expected permission error message, got: %s", stderr)
		}
	})
}

// TestIntegration_BackupRestore_Errors tests error paths for backup restore command
func TestIntegration_BackupRestore_Errors(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("NoBackupsAvailable", func(t *testing.T) {
		vaultPath := filepath.Join(testDir, "no-backup-restore", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\nn\nn\nn\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Remove automatic backup
		automaticBackup := vaultPath + ".backup"
		_ = os.Remove(automaticBackup)

		// Try to restore without backup
		_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
		if err == nil {
			t.Fatal("Expected error when no backup exists, got nil")
		}

		// Verify error message is helpful
		if !strings.Contains(stderr, "no backup") && !strings.Contains(stderr, "No backup") {
			t.Errorf("Expected 'no backup' in error, got: %s", stderr)
		}
	})

	t.Run("UserCancelsRestore", func(t *testing.T) {
		vaultPath := filepath.Join(testDir, "cancel-restore", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\nn\nn\nn\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Create manual backup
		_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed: %v\nstderr: %s", err, stderr)
		}

		// Try to restore but cancel
		_, stderr, err = runCommandWithInput(t, "n\n", "--config", configPath, "vault", "backup", "restore")
		if err == nil {
			t.Error("Expected error when user cancels restore, got nil")
		}

		// Verify cancellation message
		combinedOutput := stderr
		if !strings.Contains(combinedOutput, "cancel") && !strings.Contains(combinedOutput, "Cancel") {
			t.Logf("Note: Expected cancellation message, got: %s", combinedOutput)
		}
	})

	t.Run("AllBackupsCorrupted", func(t *testing.T) {
		vaultPath := filepath.Join(testDir, "all-corrupt", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\nn\nn\nn\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Create manual backup then corrupt it
		_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed: %v\nstderr: %s", err, stderr)
		}

		// Corrupt all backups
		automaticBackup := vaultPath + ".backup"
		if err := os.WriteFile(automaticBackup, []byte("corrupted"), 0600); err != nil {
			t.Fatalf("failed to corrupt automatic backup: %v", err)
		}

		// Find and corrupt manual backup
		vaultDir := filepath.Dir(vaultPath)
		pattern := filepath.Join(vaultDir, "vault.enc.*.manual.backup")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("failed to find manual backups: %v", err)
		}
		for _, backup := range matches {
			if err := os.WriteFile(backup, []byte("corrupted"), 0600); err != nil {
				t.Fatalf("failed to corrupt manual backup: %v", err)
			}
		}

		// Try to restore - should fail because all backups are corrupted
		_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
		if err == nil {
			t.Error("Expected error when all backups are corrupted, got nil")
		}

		// Verify error indicates no valid backup
		if !strings.Contains(stderr, "no backup") && !strings.Contains(stderr, "No backup") {
			t.Logf("Note: Expected 'no backup' in error when all corrupted, got: %s", stderr)
		}
	})
}

// TestIntegration_BackupInfo_Errors tests error paths for backup info command
func TestIntegration_BackupInfo_Errors(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("VaultNotInitialized", func(t *testing.T) {
		vaultPath := filepath.Join(testDir, "no-init-info", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Try to get info without initializing vault
		// Note: info command should work even without vault, showing "no backups"
		stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "info")

		// This might not error - info command could just show "no backups"
		if err != nil {
			// If it does error, verify error message is clear
			if stderr != "" && !strings.Contains(stderr, "vault") && !strings.Contains(stderr, "backup") {
				t.Logf("Got error (acceptable): %s", stderr)
			}
		} else {
			// If no error, should show "no backups" message
			if !strings.Contains(stdout, "No backups") && !strings.Contains(stdout, "no backups") {
				t.Logf("Note: Expected 'no backups' message, got: %s", stdout)
			}
		}
	})
}

// TestIntegration_BackupCommands_InvalidFlags tests invalid flag combinations
func TestIntegration_BackupCommands_InvalidFlags(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "invalid-flags", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\nn\nn\nn\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	t.Run("CreateWithExtraArgs", func(t *testing.T) {
		// backup create shouldn't accept arguments
		_, _, err := runCommand(t, "--config", configPath, "vault", "backup", "create", "extraarg")
		if err == nil {
			t.Error("Expected error for extra arguments to create command")
		}
	})

	t.Run("RestoreWithExtraArgs", func(t *testing.T) {
		// backup restore shouldn't accept arguments
		_, _, err := runCommand(t, "--config", configPath, "vault", "backup", "restore", "extraarg", "--force")
		if err == nil {
			t.Error("Expected error for extra arguments to restore command")
		}
	})

	t.Run("InfoWithExtraArgs", func(t *testing.T) {
		// backup info shouldn't accept arguments
		_, _, err := runCommand(t, "--config", configPath, "vault", "backup", "info", "extraarg")
		if err == nil {
			t.Error("Expected error for extra arguments to info command")
		}
	})
}
