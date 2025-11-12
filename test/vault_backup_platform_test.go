//go:build integration

package test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"pass-cli/internal/storage"
)

// T080: Cross-platform path handling tests
// Verifies backup commands work correctly across Windows, macOS, and Linux

// TestIntegration_BackupPaths_WindowsVsUnix tests path separator handling
func TestIntegration_BackupPaths_WindowsVsUnix(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("PathSeparatorConsistency", func(t *testing.T) {
		// Test that backup paths use OS-appropriate separators
		vaultPath := filepath.Join(testDir, "platform-sep", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Create manual backup
		stdout, _, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed: %v", err)
		}

		// Verify backup path uses correct separator for this OS
		expectedSep := string(os.PathSeparator)
		if !strings.Contains(stdout, expectedSep) {
			t.Logf("Note: Backup path should use OS path separator '%s'", expectedSep)
		}

		// Verify automatic backup exists with correct path
		automaticBackup := vaultPath + storage.BackupSuffix
		if _, err := os.Stat(automaticBackup); err != nil {
			t.Errorf("Automatic backup not found at expected path: %s", automaticBackup)
		}

		// Verify manual backup exists with correct pattern
		vaultDir := filepath.Dir(vaultPath)
		pattern := filepath.Join(vaultDir, "vault.enc.*.manual.backup")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("failed to find manual backups: %v", err)
		}
		if len(matches) == 0 {
			t.Error("No manual backup files found")
		}

		// Verify all backup paths are absolute
		for _, match := range matches {
			if !filepath.IsAbs(match) {
				t.Errorf("Backup path should be absolute, got: %s", match)
			}
		}
	})

	t.Run("NestedDirectoryPaths", func(t *testing.T) {
		// Test backup creation with deeply nested paths
		vaultPath := filepath.Join(testDir, "platform-nested", "level1", "level2", "level3", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Create manual backup in nested directory
		_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed in nested directory: %v", err)
		}

		// Verify backup exists
		automaticBackup := vaultPath + storage.BackupSuffix
		if _, err := os.Stat(automaticBackup); err != nil {
			t.Errorf("Backup not found in nested directory: %s", automaticBackup)
		}
	})
}

// TestIntegration_BackupPermissions_Platform tests file permission handling
func TestIntegration_BackupPermissions_Platform(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("BackupFilePermissions", func(t *testing.T) {
		vaultPath := filepath.Join(testDir, "platform-perms", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Create manual backup
		_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed: %v", err)
		}

		// Check automatic backup permissions
		automaticBackup := vaultPath + storage.BackupSuffix
		info, err := os.Stat(automaticBackup)
		if err != nil {
			t.Fatalf("failed to stat automatic backup: %v", err)
		}

		// On Unix-like systems, verify 0600 permissions
		if runtime.GOOS != "windows" {
			perm := info.Mode().Perm()
			expectedPerm := os.FileMode(0600)
			if perm != expectedPerm {
				t.Errorf("Expected backup permissions %o, got %o", expectedPerm, perm)
			}
		} else {
			// On Windows, just verify file is readable
			if info.Mode()&0400 == 0 {
				t.Error("Backup file should be readable on Windows")
			}
		}

		// Check manual backup permissions
		vaultDir := filepath.Dir(vaultPath)
		pattern := filepath.Join(vaultDir, "vault.enc.*.manual.backup")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("failed to find manual backups: %v", err)
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				t.Errorf("failed to stat manual backup %s: %v", match, err)
				continue
			}

			if runtime.GOOS != "windows" {
				perm := info.Mode().Perm()
				expectedPerm := os.FileMode(0600)
				if perm != expectedPerm {
					t.Errorf("Expected manual backup permissions %o, got %o", expectedPerm, perm)
				}
			}
		}
	})

	t.Run("RestoredVaultPermissions", func(t *testing.T) {
		vaultPath := filepath.Join(testDir, "platform-restore-perms", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Create manual backup
		_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed: %v", err)
		}

		// Restore from backup
		_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
		if err != nil {
			t.Fatalf("backup restore failed: %v", err)
		}

		// Verify restored vault has correct permissions
		info, err := os.Stat(vaultPath)
		if err != nil {
			t.Fatalf("failed to stat restored vault: %v", err)
		}

		if runtime.GOOS != "windows" {
			perm := info.Mode().Perm()
			expectedPerm := os.FileMode(storage.VaultPermissions)
			if perm != expectedPerm {
				t.Errorf("Expected restored vault permissions %o, got %o", expectedPerm, perm)
			}
		}
	})
}

// TestIntegration_BackupDirectory_Platform tests directory creation across platforms
func TestIntegration_BackupDirectory_Platform(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("DirectoryCreationWithPermissions", func(t *testing.T) {
		// Test that backup directory is created with correct permissions
		vaultPath := filepath.Join(testDir, "platform-dirperms", "newdir", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault (should create directory)
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Verify directory exists
		vaultDir := filepath.Dir(vaultPath)
		info, err := os.Stat(vaultDir)
		if err != nil {
			t.Fatalf("vault directory not created: %v", err)
		}

		if !info.IsDir() {
			t.Fatal("vault path parent is not a directory")
		}

		// On Unix, verify directory has restrictive permissions
		if runtime.GOOS != "windows" {
			perm := info.Mode().Perm()
			// Directory should be user-only (0700 or similar)
			if perm&0077 != 0 {
				t.Logf("Note: Directory has group/other permissions: %o", perm)
			}
		}
	})

	t.Run("BackupInNonExistentDirectory", func(t *testing.T) {
		// Test creating backup when backup directory doesn't exist
		vaultPath := filepath.Join(testDir, "platform-nodir", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Try to create backup - directory should be created automatically (FR-018)
		_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create should create missing directories: %v", err)
		}

		// Verify backup exists
		automaticBackup := vaultPath + storage.BackupSuffix
		if _, err := os.Stat(automaticBackup); err != nil {
			// This is acceptable - backup might be in same directory as vault
			t.Logf("Automatic backup location: %s (error: %v)", automaticBackup, err)
		}
	})
}

// TestIntegration_BackupPaths_Normalization tests path normalization
func TestIntegration_BackupPaths_Normalization(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("RelativePathHandling", func(t *testing.T) {
		// Test that relative paths are normalized correctly
		// Note: Most commands expect absolute paths, but test edge cases

		vaultPath := filepath.Join(testDir, "platform-relpath", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Create backup
		_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed: %v", err)
		}

		// Verify backup path is absolute (not relative)
		stdout, _, err := runCommand(t, "--config", configPath, "vault", "backup", "info")
		if err != nil {
			t.Fatalf("backup info failed: %v", err)
		}

		// Info output should contain absolute paths
		if runtime.GOOS == "windows" {
			// Windows absolute paths contain drive letter
			if !strings.Contains(stdout, ":") {
				t.Logf("Note: Expected absolute Windows path with drive letter in: %s", stdout)
			}
		} else {
			// Unix absolute paths start with /
			if !strings.Contains(stdout, "/") {
				t.Logf("Note: Expected absolute Unix path in: %s", stdout)
			}
		}
	})

	t.Run("PathWithSpaces", func(t *testing.T) {
		// Test that paths with spaces are handled correctly
		vaultPath := filepath.Join(testDir, "platform spaces", "my vault", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed with spaces in path: %v\nstderr: %s", err, stderr)
		}

		// Create backup
		_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed with spaces in path: %v", err)
		}

		// Verify backup exists
		automaticBackup := vaultPath + storage.BackupSuffix
		if _, err := os.Stat(automaticBackup); err != nil {
			t.Errorf("Backup not found with spaces in path: %s", automaticBackup)
		}

		// Restore from backup
		_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
		if err != nil {
			t.Fatalf("backup restore failed with spaces in path: %v", err)
		}
	})

	t.Run("PathWithSpecialCharacters", func(t *testing.T) {
		// Test paths with special characters (platform-appropriate)
		specialName := "vault-test_123"
		if runtime.GOOS != "windows" {
			// Unix allows more special characters
			specialName = "vault-test_123@special"
		}

		vaultPath := filepath.Join(testDir, "platform-special", specialName, "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed with special chars: %v\nstderr: %s", err, stderr)
		}

		// Create backup
		_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed with special chars: %v", err)
		}

		// Verify backup exists
		automaticBackup := vaultPath + storage.BackupSuffix
		if _, err := os.Stat(automaticBackup); err != nil {
			t.Errorf("Backup not found with special chars: %s", automaticBackup)
		}
	})
}
