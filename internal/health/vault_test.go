package health

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// T010: TestVaultCheck_Exists - Vault present, readable, 0600 permissions → Pass status
func TestVaultCheck_Exists(t *testing.T) {
	// Skip on Windows - Windows doesn't support Unix file permissions
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Skipping Unix permission test on Windows")
	}

	// Create temporary vault file with correct permissions
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// Write test vault file
	content := []byte("encrypted vault content")
	if err := os.WriteFile(vaultPath, content, 0600); err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	// Create vault checker
	checker := NewVaultChecker(vaultPath)

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	if result.Status != CheckPass {
		t.Errorf("Expected status %s, got %s", CheckPass, result.Status)
	}
	if result.Name != "vault" {
		t.Errorf("Expected name 'vault', got %s", result.Name)
	}

	details, ok := result.Details.(VaultCheckDetails)
	if !ok {
		t.Fatal("Expected VaultCheckDetails type")
	}
	if !details.Exists {
		t.Error("Expected Exists to be true")
	}
	if !details.Readable {
		t.Error("Expected Readable to be true")
	}
	if details.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), details.Size)
	}
	if details.Permissions != "0600" {
		t.Errorf("Expected permissions 0600, got %s", details.Permissions)
	}
}

// T011: TestVaultCheck_Missing - No vault file → Error status with specific message
func TestVaultCheck_Missing(t *testing.T) {
	// Use non-existent path
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "nonexistent-vault.enc")

	// Create vault checker
	checker := NewVaultChecker(vaultPath)

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	if result.Status != CheckError {
		t.Errorf("Expected status %s, got %s", CheckError, result.Status)
	}
	if result.Message == "" {
		t.Error("Expected error message about missing vault")
	}
	if result.Recommendation == "" {
		t.Error("Expected recommendation to initialize vault")
	}

	details, ok := result.Details.(VaultCheckDetails)
	if !ok {
		t.Fatal("Expected VaultCheckDetails type")
	}
	if details.Exists {
		t.Error("Expected Exists to be false")
	}
	if details.Error == "" {
		t.Error("Expected Error field to be populated")
	}
}

// T012: TestVaultCheck_PermissionsWarning - Vault with 0644 permissions → Warning status
func TestVaultCheck_PermissionsWarning(t *testing.T) {
	// Skip on Windows (different permission model)
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Skipping Unix permission test on Windows")
	}

	// Create temporary vault file with overly permissive permissions
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	content := []byte("encrypted vault content")
	if err := os.WriteFile(vaultPath, content, 0644); err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	// Create vault checker
	checker := NewVaultChecker(vaultPath)

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	if result.Status != CheckWarning {
		t.Errorf("Expected status %s, got %s", CheckWarning, result.Status)
	}
	if result.Message == "" {
		t.Error("Expected warning message about permissions")
	}
	if result.Recommendation == "" {
		t.Error("Expected recommendation to fix permissions")
	}

	details, ok := result.Details.(VaultCheckDetails)
	if !ok {
		t.Fatal("Expected VaultCheckDetails type")
	}
	if !details.Exists {
		t.Error("Expected Exists to be true")
	}
	if details.Permissions != "0644" {
		t.Errorf("Expected permissions 0644, got %s", details.Permissions)
	}
}
