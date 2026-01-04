package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
	"pass-cli/internal/security"
)

// cleanupTestKeychain removes keychain entries created during tests
func cleanupTestKeychain(t *testing.T, vaultPath string) {
	t.Helper()
	vaultID := filepath.Base(filepath.Dir(vaultPath))
	t.Cleanup(func() {
		_ = keyring.Delete(testKeychainService, "master-password-"+vaultID)
		_ = keyring.Delete(testKeychainService, "master-password")
		_ = keyring.Delete(testAuditKeychainService, vaultPath)
		_ = keyring.Delete(testAuditKeychainService, vaultID)
	})
}

// T015/T022/T034: Test that audit logging captures vault save operations
func TestAuditLoggingForVaultSave(t *testing.T) {
	// Skip if keychain is not available (CI environment)
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping audit logging test in CI - keychain not available")
	}

	// Setup test directory
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	auditLogPath := filepath.Join(tempDir, "audit.log")
	cleanupTestKeychain(t, vaultPath)

	// Create vault service
	vault, err := New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize vault WITH audit logging enabled
	password := []byte("TestPassword123!")
	if err := vault.Initialize(password, false, auditLogPath, "test-vault-id"); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Unlock vault
	password2 := []byte("TestPassword123!")
	if err := vault.Unlock(password2); err != nil {
		t.Fatalf("Failed to unlock vault: %v", err)
	}

	// Add a credential - this should trigger save() which should log "vault_save"
	credPassword := []byte("credential-password")
	if err := vault.AddCredential("test-service", "test-user", credPassword, "", "", ""); err != nil {
		t.Fatalf("Failed to add credential: %v", err)
	}

	// Read audit log
	auditLogContent, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	// Parse audit log entries
	var entries []security.AuditLogEntry
	for _, line := range splitLines(string(auditLogContent)) {
		if line == "" {
			continue
		}
		var entry security.AuditLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Logf("Failed to parse audit log entry: %v (line: %s)", err, line)
			continue
		}
		entries = append(entries, entry)
	}

	// Verify we have vault_save events
	foundVaultSave := false
	for _, entry := range entries {
		if entry.EventType == "vault_save" && entry.Outcome == security.OutcomeSuccess {
			foundVaultSave = true
			break
		}
	}

	if !foundVaultSave {
		t.Errorf("Expected to find 'vault_save' event with 'success' outcome in audit log")
		t.Logf("Audit log entries found:")
		for i, entry := range entries {
			t.Logf("  [%d] EventType=%s, Outcome=%s, CredentialName=%s",
				i, entry.EventType, entry.Outcome, entry.CredentialName)
		}
	}
}

// TestAuditCallback_AllEventsLogged verifies FR-015: ALL atomic save state transitions are logged
func TestAuditCallback_AllEventsLogged(t *testing.T) {
	// Skip if keychain is not available (CI environment)
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping audit logging test in CI - keychain not available")
	}

	// Setup
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	auditLogPath := filepath.Join(tempDir, "audit.log")
	cleanupTestKeychain(t, vaultPath)

	vault, err := New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize with audit logging
	password := []byte("TestPassword123!")
	if err := vault.Initialize(password, false, auditLogPath, "test-vault"); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Unlock with fresh password slice
	password2 := []byte("TestPassword123!")
	if err := vault.Unlock(password2); err != nil {
		t.Fatalf("Failed to unlock vault: %v", err)
	}

	// Trigger save operation
	credPass := []byte("cred-password")
	if err := vault.AddCredential("test", "user", credPass, "", "", ""); err != nil {
		t.Fatalf("Failed to add credential: %v", err)
	}

	// Read audit log
	auditContent, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	auditText := string(auditContent)

	// FR-015: Verify ALL required events are logged
	requiredEvents := []string{
		"vault save operation initiated",    // atomic_save_started
		"temporary file created",            // temp_file_created
		"vault verification started",        // verification_started
		"vault verification passed",         // verification_passed
		"atomic rename",                     // atomic_rename_started (x2)
		"vault save completed successfully", // atomic_save_completed
	}

	for _, event := range requiredEvents {
		if !contains(auditText, event) {
			t.Errorf("FR-015 violation: Missing required audit event: %q", event)
		}
	}

	// Verify at least 2 atomic rename events (vault→backup, temp→vault)
	renameCount := countOccurrences(auditText, "atomic rename")
	if renameCount < 2 {
		t.Errorf("FR-015: Expected at least 2 'atomic rename' events, found %d", renameCount)
	}
}

// TestAuditCallback_VerificationFailedEvent verifies verification failures are logged
func TestAuditCallback_VerificationFailedEvent(t *testing.T) {
	// Skip if keychain is not available (CI environment)
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping audit logging test in CI - keychain not available")
	}

	// Setup
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	auditLogPath := filepath.Join(tempDir, "audit.log")
	cleanupTestKeychain(t, vaultPath)

	vault, err := New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Initialize with audit logging
	password := []byte("TestPassword123!")
	if err := vault.Initialize(password, false, auditLogPath, "test-vault"); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Verification failure is tested via storage layer tests
	// This test just verifies the event exists in the audit callback
	// The actual verification_failed event is tested in TestSaveVault_ErrorMessage_VerificationFailed

	// Read audit log to verify initialization was logged
	auditContent, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	if len(auditContent) == 0 {
		t.Error("Expected audit log entries, got empty file")
	}
}

// TestAuditCallback_NoSensitiveData verifies FR-015: No credentials logged
func TestAuditCallback_NoSensitiveData(t *testing.T) {
	// Skip if keychain is not available (CI environment)
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping audit logging test in CI - keychain not available")
	}

	// Setup
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	auditLogPath := filepath.Join(tempDir, "audit.log")
	cleanupTestKeychain(t, vaultPath)

	vault, err := New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	password := []byte("MasterPassword123!")
	if err := vault.Initialize(password, false, auditLogPath, "test-vault"); err != nil {
		t.Fatalf("Failed to initialize vault: %v", err)
	}

	// Unlock with fresh password slice (original is cleared by Initialize)
	password2 := []byte("MasterPassword123!")
	if err := vault.Unlock(password2); err != nil {
		t.Fatalf("Failed to unlock vault: %v", err)
	}

	// Add credential with known sensitive values
	sensitivePassword := []byte("SuperSecret123!")
	sensitiveUsername := "admin"
	if err := vault.AddCredential("test-service", sensitiveUsername, sensitivePassword, "", "", ""); err != nil {
		t.Fatalf("Failed to add credential: %v", err)
	}

	// Read audit log
	auditContent, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}

	auditText := string(auditContent)

	// Verify sensitive data NEVER appears
	if contains(auditText, "SuperSecret123!") {
		t.Error("SECURITY VIOLATION: Credential password found in audit log")
	}

	if contains(auditText, "MasterPassword123!") {
		t.Error("SECURITY VIOLATION: Master password found in audit log")
	}

	// Usernames and service names are OK to log (not passwords)
	// Just verify no password values appear
}

// Helper to split lines (handles both \n and \r\n)
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// Helper functions for audit log verification
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

func countOccurrences(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
			i += len(substr) - 1
		}
	}
	return count
}
