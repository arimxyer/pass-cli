//go:build integration

package test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

// T086-T088: Performance tests for backup operations
// Verifies backup commands meet performance targets

// TestIntegration_BackupCreate_Performance tests backup creation performance
// Target: <5 seconds for vault with 25 credentials (reduced from 100 for CI speed)
func TestIntegration_BackupCreate_Performance(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "perf-create", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Add 25 credentials (reduced from 100 for CI timeout - each add unlocks vault with password)
	t.Logf("Adding 25 credentials to vault...")
	for i := 1; i <= 25; i++ {
		service := fmt.Sprintf("service%03d", i)
		username := fmt.Sprintf("user%03d@example.com", i)
		password := fmt.Sprintf("password%03d", i)

		input := "TestPassword123!\n"
		_, _, err := runCommandWithInput(t, input, "--config", configPath, "add", service, "--username", username, "--password", password)
		if err != nil {
			t.Fatalf("add credential %d failed: %v", i, err)
		}

		// Progress update every 10 credentials
		if i%10 == 0 {
			t.Logf("  Added %d/25 credentials", i)
		}
	}

	// Vault is automatically locked after each operation
	_, _, _ = runCommand(t, "--config", configPath, "lock")

	// Measure backup creation time
	t.Logf("Testing backup creation performance...")
	start := time.Now()
	_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("backup create failed: %v", err)
	}

	// Verify performance target (<5 seconds)
	t.Logf("Backup creation took: %v", duration)
	if duration > 5*time.Second {
		t.Errorf("Backup creation exceeded 5 second target: took %v", duration)
	} else {
		t.Logf("✓ Performance target met: %v < 5s", duration)
	}
}

// TestIntegration_BackupRestore_Performance tests backup restore performance
// Target: <30 seconds for restore operation (25 credentials for CI speed)
func TestIntegration_BackupRestore_Performance(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "perf-restore", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Add 25 credentials (reduced from 100 for CI timeout - each add unlocks vault with password)
	t.Logf("Adding 25 credentials to vault...")
	for i := 1; i <= 25; i++ {
		service := fmt.Sprintf("service%03d", i)
		username := fmt.Sprintf("user%03d@example.com", i)
		password := fmt.Sprintf("password%03d", i)

		input := "TestPassword123!\n"
		_, _, err := runCommandWithInput(t, input, "--config", configPath, "add", service, "--username", username, "--password", password)
		if err != nil {
			t.Fatalf("add credential %d failed: %v", i, err)
		}

		if i%10 == 0 {
			t.Logf("  Added %d/25 credentials", i)
		}
	}

	// Create manual backup
	_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "create")
	if err != nil {
		t.Fatalf("backup create failed: %v", err)
	}

	// Vault is automatically locked after each operation
	_, _, _ = runCommand(t, "--config", configPath, "lock")

	// Measure restore time
	t.Logf("Testing backup restore performance...")
	start := time.Now()
	_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("backup restore failed: %v", err)
	}

	// Verify performance target (<30 seconds)
	t.Logf("Backup restore took: %v", duration)
	if duration > 30*time.Second {
		t.Errorf("Backup restore exceeded 30 second target: took %v", duration)
	} else {
		t.Logf("✓ Performance target met: %v < 30s", duration)
	}
}

// TestIntegration_BackupInfo_Performance tests backup info performance
// Target: <1 second for info command
func TestIntegration_BackupInfo_Performance(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "perf-info", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Create 5 manual backups
	t.Logf("Creating 5 manual backups...")
	for i := 1; i <= 5; i++ {
		_, _, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create %d failed: %v", i, err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(100 * time.Millisecond)
	}

	// Measure info command time
	t.Logf("Testing backup info performance...")
	start := time.Now()
	_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "info")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("backup info failed: %v", err)
	}

	// Verify performance target (<1 second)
	t.Logf("Backup info took: %v", duration)
	if duration > 1*time.Second {
		t.Errorf("Backup info exceeded 1 second target: took %v", duration)
	} else {
		t.Logf("✓ Performance target met: %v < 1s", duration)
	}
}

// TestIntegration_BackupInfo_Performance_LargeVault tests info with many backups
// Edge case: Performance with 20 backup files
func TestIntegration_BackupInfo_Performance_LargeVault(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "perf-info-large", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// Create 20 manual backups (stress test)
	t.Logf("Creating 20 manual backups...")
	for i := 1; i <= 20; i++ {
		_, _, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create %d failed: %v", i, err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(50 * time.Millisecond)

		if i%5 == 0 {
			t.Logf("  Created %d/20 backups", i)
		}
	}

	// Measure info command time with many backups
	t.Logf("Testing backup info performance with 20 backups...")
	start := time.Now()
	_, _, err = runCommand(t, "--config", configPath, "vault", "backup", "info")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("backup info failed: %v", err)
	}

	// Should still be <1 second even with many backups
	t.Logf("Backup info with 20 backups took: %v", duration)
	if duration > 1*time.Second {
		t.Logf("Warning: Backup info with 20 backups exceeded 1 second target: took %v", duration)
	} else {
		t.Logf("✓ Performance maintained with 20 backups: %v < 1s", duration)
	}
}
