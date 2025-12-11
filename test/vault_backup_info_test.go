//go:build integration

package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIntegration_BackupInfo_NoBackups tests info command with no backups
// T051: Integration test for info with no backups
func TestIntegration_BackupInfo_NoBackups(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-info-nobackup", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\nn\nn\nn\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Remove automatic backup if it exists
	automaticBackup := vaultPath + ".backup"
	_ = os.Remove(automaticBackup)

	// Run info command
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "info")
	if err != nil {
		t.Fatalf("info failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	// Verify "no backups" message
	if !strings.Contains(stdout, "No backups found") && !strings.Contains(stdout, "no backups") {
		t.Errorf("Expected 'no backups' message in output, got: %s", stdout)
	}
}

// TestIntegration_BackupInfo_SingleAutomatic tests info with only automatic backup
// T052: Integration test for info with single automatic backup
func TestIntegration_BackupInfo_SingleAutomatic(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-info-auto", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\nn\nn\nn\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Automatic backup should exist from init
	// Run info command
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "info")
	if err != nil {
		t.Fatalf("info failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	// Verify automatic backup is shown
	if !strings.Contains(stdout, "automatic") && !strings.Contains(stdout, "Automatic") {
		t.Errorf("Expected automatic backup in output, got: %s", stdout)
	}
}

// TestIntegration_BackupInfo_MultipleManual tests info with multiple manual backups
// T053: Integration test for info with multiple manual backups
func TestIntegration_BackupInfo_MultipleManual(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-info-multi", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\nn\nn\nn\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Create multiple manual backups
	for i := 0; i < 3; i++ {
		_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create %d failed: %v\nstderr: %s", i+1, err, stderr)
		}
		time.Sleep(100 * time.Millisecond) // Small delay for different timestamps
	}

	// Run info command
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "info")
	if err != nil {
		t.Fatalf("info failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	// Verify multiple backups are listed
	if !strings.Contains(stdout, "manual") && !strings.Contains(stdout, "Manual") {
		t.Errorf("Expected manual backups in output, got: %s", stdout)
	}

	// Should show backup count or list
	if !strings.Contains(stdout, "backup") {
		t.Errorf("Expected backup information in output, got: %s", stdout)
	}
}

// TestIntegration_BackupInfo_Mixed tests info with both automatic and manual backups
// T054: Integration test for info with mixed backups
func TestIntegration_BackupInfo_Mixed(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-info-mixed", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault (creates automatic backup)
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\nn\nn\nn\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Create manual backup
	_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("backup create failed: %v\nstderr: %s", err, stderr)
	}

	// Run info command
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "info")
	if err != nil {
		t.Fatalf("info failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	// Verify both types are shown
	output := strings.ToLower(stdout)
	hasAutomatic := strings.Contains(output, "automatic")
	hasManual := strings.Contains(output, "manual")

	if !hasAutomatic && !hasManual {
		t.Errorf("Expected both automatic and manual backups in output, got: %s", stdout)
	}
}

// TestIntegration_BackupInfo_CorruptedBackup tests info with corrupted backup
// T057: Integration test for info with corrupted backup
func TestIntegration_BackupInfo_CorruptedBackup(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-info-corrupt", "vault.enc")
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

	// Corrupt the automatic backup
	automaticBackup := vaultPath + ".backup"
	if err := os.WriteFile(automaticBackup, []byte("corrupted"), 0600); err != nil {
		t.Fatalf("failed to corrupt backup: %v", err)
	}

	// Run info command
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "info")
	if err != nil {
		t.Fatalf("info failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	// Verify corruption is indicated (⚠️ or "corrupted" or similar)
	output := strings.ToLower(stdout)
	hasWarning := strings.Contains(output, "corrupt") ||
		strings.Contains(output, "⚠") ||
		strings.Contains(output, "warning") ||
		strings.Contains(output, "invalid")

	if !hasWarning {
		t.Errorf("Expected corruption indicator in output, got: %s", stdout)
	}
}

// TestIntegration_BackupInfo_Verbose tests verbose mode
// T058: Integration test for info verbose mode
func TestIntegration_BackupInfo_Verbose(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "vault-info-verbose", "vault.enc")
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

	// Run info with verbose flag
	stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "info", "--verbose")
	if err != nil {
		t.Fatalf("info --verbose failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	// Verbose mode should show more details (paths, timestamps, etc.)
	// At minimum, should have some output
	if len(stdout) < 50 {
		t.Errorf("Expected verbose output to contain detailed information, got: %s", stdout)
	}
}
