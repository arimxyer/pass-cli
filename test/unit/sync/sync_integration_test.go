package sync_test

import (
	"os"
	"path/filepath"
	"testing"

	"pass-cli/internal/config"
	"pass-cli/internal/security"
	"pass-cli/internal/sync"
	"pass-cli/internal/vault"
)

// TestSyncConfigValidation tests that sync configuration is properly validated
func TestSyncConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError bool
		errorField  string
	}{
		{
			name: "valid sync config",
			config: `
sync:
  enabled: true
  remote: "gdrive:.pass-cli"
`,
			expectError: false,
		},
		{
			name: "sync enabled without remote",
			config: `
sync:
  enabled: true
  remote: ""
`,
			expectError: true,
			errorField:  "sync.remote",
		},
		{
			name: "sync disabled - no validation needed",
			config: `
sync:
  enabled: false
  remote: ""
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yml")
			if err := os.WriteFile(configPath, []byte(tt.config), 0600); err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}

			// Load and validate config
			cfg, result := config.LoadFromPath(configPath)
			if cfg == nil {
				t.Fatalf("Failed to load config")
			}

			if tt.expectError {
				if result.Valid {
					t.Errorf("Expected validation error for %s, got none", tt.errorField)
				}
				found := false
				for _, e := range result.Errors {
					if e.Field == tt.errorField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error for field %s, but got errors: %v", tt.errorField, result.Errors)
				}
			} else {
				if !result.Valid {
					t.Errorf("Expected valid config, got errors: %v", result.Errors)
				}
			}
		})
	}
}

// TestSyncServiceIsEnabled tests the IsEnabled logic
func TestSyncServiceIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		remote   string
		expected bool
	}{
		{"enabled with remote", true, "gdrive:.pass-cli", true},
		{"disabled with remote", false, "gdrive:.pass-cli", false},
		{"enabled without remote", true, "", false},
		{"disabled without remote", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.SyncConfig{
				Enabled: tt.enabled,
				Remote:  tt.remote,
			}
			service := sync.NewService(cfg)
			got := service.IsEnabled()
			if got != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSyncServiceNoOpWhenDisabled tests that sync operations are no-ops when disabled
func TestSyncServiceNoOpWhenDisabled(t *testing.T) {
	cfg := config.SyncConfig{
		Enabled: false,
		Remote:  "",
	}
	service := sync.NewService(cfg)

	// Pull should succeed (no-op)
	if err := service.Pull("/tmp/test"); err != nil {
		t.Errorf("Pull() should succeed when disabled, got error: %v", err)
	}

	// Push should succeed (no-op)
	if err := service.Push("/tmp/test"); err != nil {
		t.Errorf("Push() should succeed when disabled, got error: %v", err)
	}
}

// TestPortableAuditKeyDerivation tests that audit keys can be derived from password+salt
func TestPortableAuditKeyDerivation(t *testing.T) {
	password := []byte("test-master-password")

	// Generate salt
	salt, err := security.GenerateAuditSalt()
	if err != nil {
		t.Fatalf("GenerateAuditSalt() failed: %v", err)
	}

	if len(salt) != 32 {
		t.Errorf("Salt length = %d, want 32", len(salt))
	}

	// Derive key
	key1, err := security.DeriveAuditKey(password, salt)
	if err != nil {
		t.Fatalf("DeriveAuditKey() failed: %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("Key length = %d, want 32", len(key1))
	}

	// Derive again with same inputs - should get same key
	key2, err := security.DeriveAuditKey(password, salt)
	if err != nil {
		t.Fatalf("DeriveAuditKey() second call failed: %v", err)
	}

	if string(key1) != string(key2) {
		t.Error("Same password+salt should produce same key")
	}

	// Different password should produce different key
	key3, err := security.DeriveAuditKey([]byte("different-password"), salt)
	if err != nil {
		t.Fatalf("DeriveAuditKey() with different password failed: %v", err)
	}

	if string(key1) == string(key3) {
		t.Error("Different passwords should produce different keys")
	}

	// Different salt should produce different key
	salt2, _ := security.GenerateAuditSalt()
	key4, err := security.DeriveAuditKey(password, salt2)
	if err != nil {
		t.Fatalf("DeriveAuditKey() with different salt failed: %v", err)
	}

	if string(key1) == string(key4) {
		t.Error("Different salts should produce different keys")
	}
}

// TestPortableAuditKeyDerivationErrors tests error cases
func TestPortableAuditKeyDerivationErrors(t *testing.T) {
	salt, _ := security.GenerateAuditSalt()

	// Empty password should fail
	_, err := security.DeriveAuditKey([]byte{}, salt)
	if err == nil {
		t.Error("DeriveAuditKey() should fail with empty password")
	}

	// Wrong salt length should fail
	_, err = security.DeriveAuditKey([]byte("password"), []byte("short"))
	if err == nil {
		t.Error("DeriveAuditKey() should fail with wrong salt length")
	}
}

// TestGetOrCreateAuditKeyPortable tests the portable key creation flow
func TestGetOrCreateAuditKeyPortable(t *testing.T) {
	password := []byte("test-password")

	// With password but no salt - should generate new salt
	key1, salt1, err := security.GetOrCreateAuditKeyPortable("test-vault", password, nil)
	if err != nil {
		t.Fatalf("GetOrCreateAuditKeyPortable() failed: %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("Key length = %d, want 32", len(key1))
	}
	if len(salt1) != 32 {
		t.Errorf("Salt length = %d, want 32", len(salt1))
	}

	// With password and existing salt - should use existing salt
	key2, salt2, err := security.GetOrCreateAuditKeyPortable("test-vault", password, salt1)
	if err != nil {
		t.Fatalf("GetOrCreateAuditKeyPortable() with existing salt failed: %v", err)
	}

	if string(key1) != string(key2) {
		t.Error("Same password+salt should produce same key")
	}
	if string(salt1) != string(salt2) {
		t.Error("Should return the same salt that was passed in")
	}
}

// TestMetadataAuditSalt tests that audit salt is properly stored in metadata
func TestMetadataAuditSalt(t *testing.T) {
	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// Create metadata with audit salt
	salt, _ := security.GenerateAuditSalt()
	meta := &vault.Metadata{
		Version:      "1.0",
		AuditEnabled: true,
		AuditSalt:    salt,
	}

	// Save metadata
	if err := vault.SaveMetadata(vaultPath, meta); err != nil {
		t.Fatalf("SaveMetadata() failed: %v", err)
	}

	// Load metadata
	loaded, err := vault.LoadMetadata(vaultPath)
	if err != nil {
		t.Fatalf("LoadMetadata() failed: %v", err)
	}

	if !loaded.AuditEnabled {
		t.Error("AuditEnabled should be true")
	}

	if len(loaded.AuditSalt) != 32 {
		t.Errorf("AuditSalt length = %d, want 32", len(loaded.AuditSalt))
	}

	if string(loaded.AuditSalt) != string(salt) {
		t.Error("AuditSalt should match what was saved")
	}
}

// TestNewAuditLoggerPortable tests creating a portable audit logger
func TestNewAuditLoggerPortable(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")
	password := []byte("test-password")

	// Create portable logger (generates new salt)
	logger, salt, err := security.NewAuditLoggerPortable(logPath, "test-vault", password, nil)
	if err != nil {
		t.Fatalf("NewAuditLoggerPortable() failed: %v", err)
	}

	if logger == nil {
		t.Error("Logger should not be nil")
	}
	if len(salt) != 32 {
		t.Errorf("Salt length = %d, want 32", len(salt))
	}

	// Create another logger with same salt - should work
	logger2, salt2, err := security.NewAuditLoggerPortable(logPath, "test-vault", password, salt)
	if err != nil {
		t.Fatalf("NewAuditLoggerPortable() with existing salt failed: %v", err)
	}

	if logger2 == nil {
		t.Error("Logger2 should not be nil")
	}
	if string(salt) != string(salt2) {
		t.Error("Salt should be preserved when passed in")
	}
}

// TestAuditLogEntryWithPortableKey tests signing/verifying with derived key
func TestAuditLogEntryWithPortableKey(t *testing.T) {
	password := []byte("test-password")
	salt, _ := security.GenerateAuditSalt()
	key, _ := security.DeriveAuditKey(password, salt)

	entry := &security.AuditLogEntry{
		EventType:      "test_event",
		Outcome:        "success",
		CredentialName: "test-service",
		MachineID:      "test-machine",
	}

	// Sign with derived key
	if err := entry.Sign(key); err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	if len(entry.HMACSignature) == 0 {
		t.Error("Signature should not be empty after signing")
	}

	// Verify with same key
	if err := entry.Verify(key); err != nil {
		t.Errorf("Verify() failed: %v", err)
	}

	// Verify with different key should fail
	differentKey, _ := security.DeriveAuditKey([]byte("different"), salt)
	if err := entry.Verify(differentKey); err == nil {
		t.Error("Verify() should fail with different key")
	}
}

// TestCrossOSAuditKeyDerivation simulates cross-OS scenario
func TestCrossOSAuditKeyDerivation(t *testing.T) {
	// Simulate: Windows creates vault with audit
	password := []byte("shared-master-password")
	salt, _ := security.GenerateAuditSalt()

	// Windows derives key and signs entry
	windowsKey, _ := security.DeriveAuditKey(password, salt)
	entry := &security.AuditLogEntry{
		EventType:      "credential_access",
		Outcome:        "success",
		CredentialName: "github",
		MachineID:      "windows-pc",
	}
	if err := entry.Sign(windowsKey); err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	// Simulate: Linux receives vault (with salt in metadata) and verifies
	// Using same password + salt should derive same key
	linuxKey, _ := security.DeriveAuditKey(password, salt)

	// Keys should be identical
	if string(windowsKey) != string(linuxKey) {
		t.Error("Cross-OS key derivation should produce identical keys")
	}

	// Linux should be able to verify Windows-signed entries
	if err := entry.Verify(linuxKey); err != nil {
		t.Errorf("Cross-OS verification failed: %v", err)
	}
}
