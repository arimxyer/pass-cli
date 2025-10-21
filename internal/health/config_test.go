package health

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// T013: TestConfigCheck_Valid - Valid YAML, all values in range → Pass status
func TestConfigCheck_Valid(t *testing.T) {
	// Create temporary config file with valid values
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	validConfig := `# Pass-CLI configuration
vault_path: ~/.pass-cli/vault.enc
clipboard_timeout: 30
audit_enabled: true
`
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create config checker
	checker := NewConfigChecker(configPath)

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	if result.Status != CheckPass {
		t.Errorf("Expected status %s, got %s", CheckPass, result.Status)
	}
	if result.Name != "config" {
		t.Errorf("Expected name 'config', got %s", result.Name)
	}

	details, ok := result.Details.(ConfigCheckDetails)
	if !ok {
		t.Fatal("Expected ConfigCheckDetails type")
	}
	if !details.Exists {
		t.Error("Expected Exists to be true")
	}
	if !details.Valid {
		t.Error("Expected Valid to be true")
	}
	if len(details.Errors) > 0 {
		t.Errorf("Expected no errors, got %d", len(details.Errors))
	}
	if len(details.UnknownKeys) > 0 {
		t.Errorf("Expected no unknown keys, got %v", details.UnknownKeys)
	}
}

// T014: TestConfigCheck_InvalidValue - clipboard_timeout=500 → Warning status with recommendation
func TestConfigCheck_InvalidValue(t *testing.T) {
	// Create temporary config file with out-of-range value
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidConfig := `vault_path: ~/.pass-cli/vault.enc
clipboard_timeout: 500
audit_enabled: true
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create config checker
	checker := NewConfigChecker(configPath)

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	if result.Status != CheckWarning {
		t.Errorf("Expected status %s, got %s", CheckWarning, result.Status)
	}
	if result.Recommendation == "" {
		t.Error("Expected recommendation to fix value range")
	}

	details, ok := result.Details.(ConfigCheckDetails)
	if !ok {
		t.Fatal("Expected ConfigCheckDetails type")
	}
	if !details.Exists {
		t.Error("Expected Exists to be true")
	}
	if len(details.Errors) == 0 {
		t.Error("Expected validation errors for clipboard_timeout")
	}

	// Check that the error mentions clipboard_timeout
	foundClipboardError := false
	for _, err := range details.Errors {
		if err.Key == "clipboard_timeout" {
			foundClipboardError = true
			if err.Problem == "" {
				t.Error("Expected problem description")
			}
			if err.CurrentValue != "500" {
				t.Errorf("Expected current value 500, got %s", err.CurrentValue)
			}
		}
	}
	if !foundClipboardError {
		t.Error("Expected error for clipboard_timeout key")
	}
}

// T015: TestConfigCheck_UnknownKeys - Typo in config key → Warning status
func TestConfigCheck_UnknownKeys(t *testing.T) {
	// Create temporary config file with unknown keys (typos)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	typoConfig := `vault_path: ~/.pass-cli/vault.enc
clipbaord_timeout: 30
auidt_enabled: true
`
	if err := os.WriteFile(configPath, []byte(typoConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create config checker
	checker := NewConfigChecker(configPath)

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	if result.Status != CheckWarning {
		t.Errorf("Expected status %s, got %s", CheckWarning, result.Status)
	}
	if result.Message == "" {
		t.Error("Expected message about unknown keys")
	}

	details, ok := result.Details.(ConfigCheckDetails)
	if !ok {
		t.Fatal("Expected ConfigCheckDetails type")
	}
	if len(details.UnknownKeys) == 0 {
		t.Error("Expected unknown keys to be detected")
	}

	// Check for the typo keys
	hasClipbaordTypo := false
	hasAuidtTypo := false
	for _, key := range details.UnknownKeys {
		if key == "clipbaord_timeout" {
			hasClipbaordTypo = true
		}
		if key == "auidt_enabled" {
			hasAuidtTypo = true
		}
	}
	if !hasClipbaordTypo {
		t.Error("Expected 'clipbaord_timeout' in unknown keys")
	}
	if !hasAuidtTypo {
		t.Error("Expected 'auidt_enabled' in unknown keys")
	}
}
