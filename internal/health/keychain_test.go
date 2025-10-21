package health

import (
	"context"
	"os"
	"testing"
)

// T016: TestKeychainCheck_Healthy - Password stored, no orphaned entries → Pass status
func TestKeychainCheck_Healthy(t *testing.T) {
	// Create temporary vault file for testing
	tmpDir := t.TempDir()
	vaultPath := tmpDir + "/vault.enc"
	// Create the vault file
	if err := os.WriteFile(vaultPath, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test vault: %v", err)
	}

	// Create mock keyring with password for the vault (healthy state)
	mock := newMockKeyringService(map[string]string{
		"pass-cli:" + vaultPath: "password123", // Password stored for existing vault
	})

	// Create keychain checker with mock
	checker := &KeychainChecker{
		defaultVaultPath: vaultPath,
		keyring:          mock,
	}

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	if result.Status != CheckPass {
		t.Errorf("Expected status %s, got %s", CheckPass, result.Status)
	}

	details, ok := result.Details.(KeychainCheckDetails)
	if !ok {
		t.Fatal("Expected KeychainCheckDetails type")
	}
	if !details.Available {
		t.Error("Expected Available to be true")
	}
	if len(details.OrphanedEntries) > 0 {
		t.Errorf("Expected no orphaned entries, got %d", len(details.OrphanedEntries))
	}
	if details.Backend == "" {
		t.Error("Expected backend name to be populated")
	}
	if details.CurrentVault == nil {
		t.Error("Expected CurrentVault to be populated")
	}
}

// T017: TestKeychainCheck_OrphanedEntries - Keychain entry for deleted vault → Error status with orphan list
func TestKeychainCheck_OrphanedEntries(t *testing.T) {
	// Create mock keyring with orphaned entries
	// Note: In production, go-keyring doesn't support List(), so orphaned detection won't work.
	// But we can test the business logic with a mock that does support List().
	mock := newMockKeyringService(map[string]string{
		"pass-cli:/tmp/deleted-vault.enc":  "password1", // Orphaned (vault doesn't exist)
		"pass-cli:/tmp/deleted-vault2.enc": "password2", // Orphaned (vault doesn't exist)
	})

	checker := &KeychainChecker{
		defaultVaultPath: "/home/user/.pass-cli/vault.enc", // Not in mock, so no current password
		keyring:          mock,
	}

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	if result.Status != CheckError {
		t.Errorf("Expected status %s when orphaned entries exist, got %s", CheckError, result.Status)
	}
	if result.Message == "" {
		t.Error("Expected message about orphaned entries")
	}
	if !contains(result.Message, "orphaned") {
		t.Errorf("Expected message to mention 'orphaned', got: %s", result.Message)
	}
	if result.Recommendation == "" {
		t.Error("Expected recommendation to clean up orphaned entries")
	}

	details, ok := result.Details.(KeychainCheckDetails)
	if !ok {
		t.Fatal("Expected KeychainCheckDetails type")
	}
	if len(details.OrphanedEntries) != 2 {
		t.Errorf("Expected 2 orphaned entries, got %d", len(details.OrphanedEntries))
	}

	// Verify orphaned entry structure
	for _, entry := range details.OrphanedEntries {
		if entry.Key == "" {
			t.Error("Expected Key to be populated")
		}
		if entry.VaultPath == "" {
			t.Error("Expected VaultPath to be populated")
		}
		if entry.Exists {
			t.Error("Expected Exists to be false for orphaned entries")
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
