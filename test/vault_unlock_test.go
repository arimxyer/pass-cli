package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"pass-cli/internal/crypto"
	"pass-cli/internal/storage"
	"pass-cli/internal/vault"
)

// T026: Integration test: unlock v2 vault with correct password
func TestUnlockV2VaultWithCorrectPassword(t *testing.T) {
	// Create temp directory for vault
	tempDir, err := os.MkdirTemp("", "vault-unlock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vaultPath := filepath.Join(tempDir, "vault.enc")
	passwordStr := "Test123!@#Password"

	// Create v2 vault with recovery enabled
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	// Make a fresh copy for initialization (will be cleared)
	initPassword := []byte(passwordStr)
	_, err = vs.InitializeWithRecovery(initPassword, false, "", "", nil)
	if err != nil {
		t.Fatalf("InitializeWithRecovery() error = %v", err)
	}

	// Verify vault is v2
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault file: %v", err)
	}

	var encryptedVault struct {
		Metadata storage.VaultMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(vaultData, &encryptedVault); err != nil {
		t.Fatalf("Failed to parse vault file: %v", err)
	}

	if encryptedVault.Metadata.Version != 2 {
		t.Fatalf("Expected v2 vault, got version %d", encryptedVault.Metadata.Version)
	}

	// Create new vault service to simulate fresh unlock
	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create second vault service: %v", err)
	}

	// Unlock with correct password (make a fresh copy from string)
	unlockPassword := []byte(passwordStr)
	err = vs2.Unlock(unlockPassword)
	if err != nil {
		t.Fatalf("Unlock() with correct password should succeed, got error: %v", err)
	}

	// Verify vault is unlocked
	if !vs2.IsUnlocked() {
		t.Error("Vault should be unlocked after successful Unlock()")
	}
}

// T027: Integration test: unlock v2 vault with wrong password fails
func TestUnlockV2VaultWithWrongPassword(t *testing.T) {
	// Create temp directory for vault
	tempDir, err := os.MkdirTemp("", "vault-unlock-wrong-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vaultPath := filepath.Join(tempDir, "vault.enc")

	// Create v2 vault with recovery enabled
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	initPassword := []byte("Test123!@#Password")
	_, err = vs.InitializeWithRecovery(initPassword, false, "", "", nil)
	if err != nil {
		t.Fatalf("InitializeWithRecovery() error = %v", err)
	}

	// Create new vault service to simulate fresh unlock
	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create second vault service: %v", err)
	}

	// Unlock with wrong password
	wrongPassword := []byte("Wrong123!@#Password")
	err = vs2.Unlock(wrongPassword)
	if err == nil {
		t.Fatal("Unlock() with wrong password should fail")
	}

	// Verify vault is NOT unlocked
	if vs2.IsUnlocked() {
		t.Error("Vault should NOT be unlocked after failed Unlock()")
	}
}

// T028: Integration test: unlock v1 vault still works (backward compat)
func TestUnlockV1VaultBackwardCompatibility(t *testing.T) {
	// Create temp directory for vault
	tempDir, err := os.MkdirTemp("", "vault-unlock-v1-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vaultPath := filepath.Join(tempDir, "vault.enc")
	passwordStr := "Test123!@#Password"

	// Create v1 vault (without recovery)
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	initPassword := []byte(passwordStr)
	err = vs.Initialize(initPassword, false, "", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Verify vault is v1
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault file: %v", err)
	}

	var encryptedVault struct {
		Metadata storage.VaultMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(vaultData, &encryptedVault); err != nil {
		t.Fatalf("Failed to parse vault file: %v", err)
	}

	if encryptedVault.Metadata.Version != 1 {
		t.Fatalf("Expected v1 vault, got version %d", encryptedVault.Metadata.Version)
	}

	// Create new vault service to simulate fresh unlock
	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create second vault service: %v", err)
	}

	// Unlock with correct password (make a fresh copy from string)
	unlockPassword := []byte(passwordStr)
	err = vs2.Unlock(unlockPassword)
	if err != nil {
		t.Fatalf("Unlock() v1 vault with correct password should succeed, got error: %v", err)
	}

	// Verify vault is unlocked
	if !vs2.IsUnlocked() {
		t.Error("V1 vault should be unlocked after successful Unlock()")
	}
}

// T028.1: Integration test: unlock with corrupted/missing WrappedDEK metadata fails gracefully
func TestUnlockV2VaultCorruptedMetadataFailsGracefully(t *testing.T) {
	// Create temp directory for vault
	tempDir, err := os.MkdirTemp("", "vault-unlock-corrupt-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vaultPath := filepath.Join(tempDir, "vault.enc")
	passwordStr := "Test123!@#Password"

	// Create v2 vault with recovery enabled
	vs, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create vault service: %v", err)
	}

	initPassword := []byte(passwordStr)
	_, err = vs.InitializeWithRecovery(initPassword, false, "", "", nil)
	if err != nil {
		t.Fatalf("InitializeWithRecovery() error = %v", err)
	}

	// Read and corrupt the vault file
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("Failed to read vault file: %v", err)
	}

	var encryptedVault storage.EncryptedVault
	if err := json.Unmarshal(vaultData, &encryptedVault); err != nil {
		t.Fatalf("Failed to parse vault file: %v", err)
	}

	// Corrupt the WrappedDEK by truncating it
	encryptedVault.Metadata.WrappedDEK = encryptedVault.Metadata.WrappedDEK[:16] // Should be 48 bytes

	// Write corrupted vault back
	corruptedData, err := json.Marshal(encryptedVault)
	if err != nil {
		t.Fatalf("Failed to marshal corrupted vault: %v", err)
	}
	if err := os.WriteFile(vaultPath, corruptedData, 0600); err != nil {
		t.Fatalf("Failed to write corrupted vault: %v", err)
	}

	// Create new vault service to simulate fresh unlock
	vs2, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("Failed to create second vault service: %v", err)
	}

	// Unlock should fail gracefully (not panic)
	unlockPassword := []byte(passwordStr)
	err = vs2.Unlock(unlockPassword)
	if err == nil {
		t.Fatal("Unlock() with corrupted WrappedDEK should fail")
	}

	// Verify error message doesn't leak key material
	errMsg := err.Error()
	if len(errMsg) > 200 {
		t.Error("Error message should be concise, not leak large amounts of data")
	}
}

// Helper test: verify both v1 and v2 can add and retrieve credentials
func TestV1AndV2VaultsCanManageCredentials(t *testing.T) {
	passwordStr := "Test123!@#Password"

	testCases := []struct {
		name     string
		initFunc func(vs *vault.VaultService, password []byte) error
		version  int
	}{
		{
			name: "v1_vault",
			initFunc: func(vs *vault.VaultService, password []byte) error {
				return vs.Initialize(password, false, "", "")
			},
			version: 1,
		},
		{
			name: "v2_vault",
			initFunc: func(vs *vault.VaultService, password []byte) error {
				_, err := vs.InitializeWithRecovery(password, false, "", "", nil)
				return err
			},
			version: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temp directory
			tempDir, err := os.MkdirTemp("", "vault-cred-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			vaultPath := filepath.Join(tempDir, "vault.enc")

			// Create vault
			vs, err := vault.New(vaultPath)
			if err != nil {
				t.Fatalf("Failed to create vault service: %v", err)
			}

			initPassword := []byte(passwordStr)
			if err := tc.initFunc(vs, initPassword); err != nil {
				t.Fatalf("Init error = %v", err)
			}

			// Re-create vault service and unlock
			vs2, err := vault.New(vaultPath)
			if err != nil {
				t.Fatalf("Failed to create second vault service: %v", err)
			}

			unlockPassword := []byte(passwordStr)
			if err := vs2.Unlock(unlockPassword); err != nil {
				t.Fatalf("Unlock() error = %v", err)
			}

			// Add a credential
			credPassword := []byte("secret123")
			if err := vs2.AddCredential("test-service", "testuser", credPassword, "", "", ""); err != nil {
				t.Fatalf("AddCredential() error = %v", err)
			}

			// Lock and unlock again
			vs2.Lock()

			vs3, err := vault.New(vaultPath)
			if err != nil {
				t.Fatalf("Failed to create third vault service: %v", err)
			}

			unlockPassword2 := []byte(passwordStr)
			if err := vs3.Unlock(unlockPassword2); err != nil {
				t.Fatalf("Second Unlock() error = %v", err)
			}

			// Retrieve credential
			cred, err := vs3.GetCredential("test-service", false)
			if err != nil {
				t.Fatalf("GetCredential() error = %v", err)
			}

			if cred.Username != "testuser" {
				t.Errorf("Username = %q, want %q", cred.Username, "testuser")
			}
			if string(cred.Password) != "secret123" {
				t.Errorf("Password mismatch")
			}

			// Clear credential password after checking
			crypto.ClearBytes(cred.Password)
		})
	}
}
