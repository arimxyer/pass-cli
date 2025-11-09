package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"pass-cli/internal/security"
)

// T015/T022/T034: Test that audit logging captures vault save operations
func TestAuditLoggingForVaultSave(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")
	auditLogPath := filepath.Join(tempDir, "audit.log")

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
