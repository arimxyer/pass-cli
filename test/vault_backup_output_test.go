//go:build integration

package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T079: CLI output formatting consistency tests
// Verifies consistent message formats across backup commands

// TestIntegration_BackupOutput_SuccessMessages tests success message consistency
func TestIntegration_BackupOutput_SuccessMessages(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "output-success", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	t.Run("CreateSuccessFormat", func(t *testing.T) {
		stdout, _, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed: %v", err)
		}

		// Verify success message format
		if !strings.Contains(stdout, "✅") && !strings.Contains(stdout, "success") {
			t.Error("Success message missing success indicator")
		}

		// Verify contains backup path
		if !strings.Contains(stdout, "Backup:") {
			t.Error("Success message missing 'Backup:' label")
		}

		// Verify contains size
		if !strings.Contains(stdout, "Size:") {
			t.Error("Success message missing 'Size:' label")
		}

		// Verify contains timestamp
		if !strings.Contains(stdout, "Created:") {
			t.Error("Success message missing 'Created:' label")
		}

		// Verify contains next steps
		if !strings.Contains(stdout, "restore") {
			t.Error("Success message missing guidance about restore")
		}
	})

	t.Run("RestoreSuccessFormat", func(t *testing.T) {
		stdout, _, err := runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
		if err != nil {
			t.Fatalf("backup restore failed: %v", err)
		}

		// Verify success message format
		if !strings.Contains(stdout, "✅") && !strings.Contains(stdout, "success") {
			t.Error("Success message missing success indicator")
		}

		// Verify contains restored from path
		if !strings.Contains(stdout, "Restored from:") {
			t.Error("Success message missing 'Restored from:' label")
		}

		// Verify contains backup type
		if !strings.Contains(stdout, "Backup type:") {
			t.Error("Success message missing 'Backup type:' label")
		}

		// Verify contains next steps
		if !strings.Contains(stdout, "unlock") {
			t.Error("Success message missing guidance about unlocking")
		}
	})

	t.Run("InfoOutputFormat", func(t *testing.T) {
		stdout, _, err := runCommand(t, "--config", configPath, "vault", "backup", "info")
		if err != nil {
			t.Fatalf("backup info failed: %v", err)
		}

		// Verify header format
		if !strings.Contains(stdout, "Backup Status") {
			t.Error("Info output missing status header")
		}

		// Verify backup listing format
		if !strings.Contains(stdout, "Backup") {
			t.Error("Info output missing backup information")
		}

		// Verify total size displayed
		if !strings.Contains(stdout, "Total backup size:") {
			t.Error("Info output missing total size")
		}
	})
}

// TestIntegration_BackupOutput_ErrorMessages tests error message consistency
func TestIntegration_BackupOutput_ErrorMessages(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("VaultNotFoundFormat", func(t *testing.T) {
		vaultPath := filepath.Join(testDir, "output-error-notfound", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		_, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err == nil {
			t.Fatal("Expected error for non-existent vault")
		}

		// Verify error message format
		if !strings.Contains(stderr, "Error:") && !strings.Contains(stderr, "error") {
			t.Error("Error message missing 'Error:' prefix")
		}

		// Verify contains helpful context
		if !strings.Contains(stderr, "vault") && !strings.Contains(stderr, "not found") {
			t.Error("Error message not descriptive enough")
		}
	})

	t.Run("NoBackupFormat", func(t *testing.T) {
		vaultPath := filepath.Join(testDir, "output-error-nobackup", "vault.enc")
		configPath, cleanup := setupTestVaultConfig(t, vaultPath)
		defer cleanup()

		// Initialize vault
		_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
		if err != nil {
			t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
		}

		// Remove backup
		_ = os.Remove(vaultPath + ".backup")

		_, stderr, err = runCommand(t, "--config", configPath, "vault", "backup", "restore", "--force")
		if err == nil {
			t.Fatal("Expected error for no backup")
		}

		// Verify error message format
		if !strings.Contains(stderr, "Error:") && !strings.Contains(stderr, "error") {
			t.Error("Error message missing 'Error:' prefix")
		}

		// Verify contains helpful guidance
		if !strings.Contains(stderr, "backup") {
			t.Error("Error message should mention backup")
		}
	})
}

// TestIntegration_BackupOutput_VerboseMode tests verbose output consistency
func TestIntegration_BackupOutput_VerboseMode(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "output-verbose", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	t.Run("CreateVerboseFormat", func(t *testing.T) {
		stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "create", "--verbose")
		if err != nil {
			t.Fatalf("backup create failed: %v", err)
		}

		combinedOutput := stdout + stderr

		// Verify verbose markers
		if !strings.Contains(combinedOutput, "[VERBOSE]") {
			t.Error("Verbose output missing [VERBOSE] markers")
		}

		// Verify shows vault path
		if !strings.Contains(combinedOutput, "Vault path:") {
			t.Error("Verbose output missing vault path")
		}

		// Verify shows operation progress
		if !strings.Contains(combinedOutput, "completed") || !strings.Contains(combinedOutput, "Creating") {
			t.Error("Verbose output missing operation progress")
		}
	})

	t.Run("RestoreVerboseFormat", func(t *testing.T) {
		stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "restore", "--verbose", "--force")
		if err != nil {
			t.Fatalf("backup restore failed: %v", err)
		}

		combinedOutput := stdout + stderr

		// Verify verbose markers
		if !strings.Contains(combinedOutput, "[VERBOSE]") {
			t.Error("Verbose output missing [VERBOSE] markers")
		}

		// Verify shows vault path
		if !strings.Contains(combinedOutput, "Vault path:") {
			t.Error("Verbose output missing vault path")
		}

		// Verify shows operation progress
		if !strings.Contains(combinedOutput, "completed") || !strings.Contains(combinedOutput, "Starting") {
			t.Error("Verbose output missing operation progress")
		}
	})

	t.Run("InfoVerboseFormat", func(t *testing.T) {
		stdout, stderr, err := runCommand(t, "--config", configPath, "vault", "backup", "info", "--verbose")
		if err != nil {
			t.Fatalf("backup info failed: %v", err)
		}

		combinedOutput := stdout + stderr

		// Verbose mode for info should show more details
		if strings.Contains(combinedOutput, "[VERBOSE]") {
			// If verbose markers exist, verify they're consistent
			if !strings.Contains(combinedOutput, "Vault path:") {
				t.Error("Verbose info output missing vault path")
			}
		}

		// Verify verbose info shows full paths
		if !strings.Contains(combinedOutput, "Path:") && !strings.Contains(combinedOutput, ".backup") {
			t.Error("Verbose info output should show full backup paths")
		}
	})
}

// TestIntegration_BackupOutput_WarningMessages tests warning message consistency
func TestIntegration_BackupOutput_WarningMessages(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "output-warnings", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	t.Run("RestoreWarningFormat", func(t *testing.T) {
		// Create backup
		_, _, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed: %v", err)
		}

		// Try restore without --force to see warning
		stdout, _, _ := runCommandWithInput(t, "n\n", "--config", configPath, "vault", "backup", "restore")

		// Verify warning format
		if !strings.Contains(stdout, "⚠️") && !strings.Contains(stdout, "Warning") {
			t.Error("Warning message missing warning indicator")
		}

		// Verify explains consequences
		if !strings.Contains(stdout, "overwrite") {
			t.Error("Warning should explain it will overwrite vault")
		}
	})
}

// TestIntegration_BackupOutput_StructureConsistency tests output structure consistency
func TestIntegration_BackupOutput_StructureConsistency(t *testing.T) {
	setupTestEnvironment(t)

	vaultPath := filepath.Join(testDir, "output-structure", "vault.enc")
	configPath, cleanup := setupTestVaultConfig(t, vaultPath)
	defer cleanup()

	// Initialize vault
	_, stderr, err := runCommandWithInput(t, "TestPassword123!\nTestPassword123!\n", "--config", configPath, "init")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	t.Run("EmptyLinesConsistency", func(t *testing.T) {
		// Create backup
		stdout, _, err := runCommand(t, "--config", configPath, "vault", "backup", "create")
		if err != nil {
			t.Fatalf("backup create failed: %v", err)
		}

		// Check for consistent spacing (success message should have blank lines)
		lines := strings.Split(stdout, "\n")
		hasEmptyLines := false
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				hasEmptyLines = true
				break
			}
		}
		if !hasEmptyLines {
			t.Error("Output should have blank lines for readability")
		}
	})

	t.Run("LabelConsistency", func(t *testing.T) {
		// All commands should use consistent label formats (key: value)
		stdout, _, err := runCommand(t, "--config", configPath, "vault", "backup", "info")
		if err != nil {
			t.Fatalf("backup info failed: %v", err)
		}

		// Check for consistent colon usage in labels
		hasColonLabels := strings.Contains(stdout, ":") &&
			(strings.Contains(stdout, "Backup") || strings.Contains(stdout, "Total"))

		if !hasColonLabels {
			t.Error("Output should use consistent 'Label:' format")
		}
	})
}
