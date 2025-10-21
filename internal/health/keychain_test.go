package health

import (
	"context"
	"testing"
)

// T016: TestKeychainCheck_Healthy - Password stored, no orphaned entries → Pass status
func TestKeychainCheck_Healthy(t *testing.T) {
	// Note: This test will need to mock the keychain service
	// For now, we'll test the logic assuming keychain returns valid data

	// Create keychain checker
	checker := &KeychainChecker{
		defaultVaultPath: "/home/user/.pass-cli/vault.enc",
		// Will need to inject mock keychain service in actual implementation
	}

	// Execute check
	result := checker.Run(context.Background())

	// Assertions
	// This test will fail until we implement the actual keychain check
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
}

// T017: TestKeychainCheck_OrphanedEntries - Keychain entry for deleted vault → Error status with orphan list
func TestKeychainCheck_OrphanedEntries(t *testing.T) {
	// Note: This test will need to mock the keychain service with orphaned entries
	// For now, we'll define the expected behavior

	// Create keychain checker
	checker := &KeychainChecker{
		defaultVaultPath: "/home/user/.pass-cli/vault.enc",
		// Will need to inject mock keychain with orphaned entries
	}

	// Execute check (will fail until implementation)
	result := checker.Run(context.Background())

	// Assertions
	// In a real test with mocked orphaned entries, status should be Error
	// For now, this will fail and guide implementation
	if result.Status != CheckError {
		t.Errorf("Expected status %s when orphaned entries exist, got %s", CheckError, result.Status)
	}
	if result.Message == "" {
		t.Error("Expected message about orphaned entries")
	}
	if result.Recommendation == "" {
		t.Error("Expected recommendation to clean up orphaned entries")
	}

	details, ok := result.Details.(KeychainCheckDetails)
	if !ok {
		t.Fatal("Expected KeychainCheckDetails type")
	}
	if len(details.OrphanedEntries) == 0 {
		t.Error("Expected orphaned entries to be detected")
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
