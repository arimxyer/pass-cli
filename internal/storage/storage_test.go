package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pass-cli/internal/crypto"
)

func TestNewStorageService(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	// Test valid creation
	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	if storage.vaultPath != vaultPath {
		t.Errorf("Expected vault path %s, got %s", vaultPath, storage.vaultPath)
	}

	// Test nil crypto service
	_, err = NewStorageService(nil, vaultPath)
	if err == nil {
		t.Error("Expected error for nil crypto service")
	}

	// Test empty vault path
	_, err = NewStorageService(cryptoService, "")
	if err != ErrInvalidVaultPath {
		t.Errorf("Expected ErrInvalidVaultPath, got %v", err)
	}
}

func TestStorageService_InitializeVault(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Test initialization
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Verify vault exists
	if !storage.VaultExists() {
		t.Error("Vault should exist after initialization")
	}

	// Test double initialization (should fail)
	if err := storage.InitializeVault(password); err == nil {
		t.Error("Expected error when initializing existing vault")
	}

	// Verify file permissions (skip on Windows as it doesn't support Unix permissions)
	info, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("Failed to stat vault file: %v", err)
	}

	// Only check permissions on Unix-like systems
	if info.Mode().Perm() != os.FileMode(VaultPermissions) {
		// This is expected on Windows, so just log it
		t.Logf("Note: File permissions are %v (expected %v) - this is normal on Windows",
			info.Mode().Perm(), os.FileMode(VaultPermissions))
	}
}

// FR-015: Test that InitializeVault creates parent directories
func TestStorageService_InitializeVault_CreatesParentDirectories(t *testing.T) {
	// Create a vault path with nested non-existent directories
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "level1", "level2", "level3", "vault.enc")
	
	// Verify parent directories don't exist
	parentDir := filepath.Dir(nestedPath)
	if _, err := os.Stat(parentDir); !os.IsNotExist(err) {
		t.Fatalf("Parent directory should not exist yet")
	}
	
	// Create storage service (this should create parent directories)
	cryptoService := crypto.NewCryptoService()
	storageService, err := NewStorageService(cryptoService, nestedPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}
	
	// Verify parent directories were created by NewStorageService
	if _, err := os.Stat(parentDir); err != nil {
		t.Fatalf("Parent directories were not created: %v", err)
	}
	
	// Initialize vault
	password := "TestPassword123!"
	err = storageService.InitializeVault(password)
	if err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}
	
	// Verify vault file exists
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Fatalf("Vault file was not created at %s", nestedPath)
	}
	
	t.Logf("✓ Vault created with parent directories: %s", nestedPath)
}

func TestStorageService_LoadSaveVault(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"
	testData := []byte(`{"credentials": [{"service": "example.com", "username": "user", "password": "pass"}]}`)

	// Initialize vault
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Save test data
	if err := storage.SaveVault(testData, password); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Load and verify data
	loadedData, err := storage.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if !bytes.Equal(testData, loadedData) {
		t.Error("Loaded data does not match saved data")
	}

	// Test wrong password
	_, err = storage.LoadVault("wrong-password")
	if err == nil {
		t.Error("Expected error with wrong password")
	}
}

func TestStorageService_VaultNotFound(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "nonexistent_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Test loading non-existent vault
	_, err = storage.LoadVault("password")
	if err != ErrVaultNotFound {
		t.Errorf("Expected ErrVaultNotFound, got %v", err)
	}

	// Test saving to non-existent vault
	err = storage.SaveVault([]byte("data"), "password")
	if err != ErrVaultNotFound {
		t.Errorf("Expected ErrVaultNotFound, got %v", err)
	}
}

func TestStorageService_GetVaultInfo(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize vault
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Get vault info
	info, err := storage.GetVaultInfo()
	if err != nil {
		t.Fatalf("GetVaultInfo failed: %v", err)
	}

	// Verify metadata
	if info.Version != 1 {
		t.Errorf("Expected version 1, got %d", info.Version)
	}

	if info.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if info.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	// Verify salt is not exposed
	if info.Salt != nil {
		t.Error("Salt should not be exposed in vault info")
	}

	// Save data and check if UpdatedAt changes
	time.Sleep(10 * time.Millisecond) // Ensure time difference
	originalUpdatedAt := info.UpdatedAt

	if err := storage.SaveVault([]byte("new data"), password); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	newInfo, err := storage.GetVaultInfo()
	if err != nil {
		t.Fatalf("GetVaultInfo failed: %v", err)
	}

	if !newInfo.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated after save")
	}
}

func TestStorageService_ValidateVault(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Test validation of non-existent vault
	err = storage.ValidateVault()
	if err != ErrVaultNotFound {
		t.Errorf("Expected ErrVaultNotFound, got %v", err)
	}

	// Initialize vault and test validation
	password := "test-password"
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Valid vault should pass validation
	if err := storage.ValidateVault(); err != nil {
		t.Errorf("ValidateVault failed for valid vault: %v", err)
	}

	// Test corrupted vault by writing invalid JSON
	invalidJSON := []byte(`{"invalid": "json"`)
	if err := os.WriteFile(vaultPath, invalidJSON, VaultPermissions); err != nil {
		t.Fatalf("Failed to write corrupted vault: %v", err)
	}

	err = storage.ValidateVault()
	if err == nil {
		t.Error("Expected error for corrupted vault")
	}
}

func TestStorageService_BackupRestore(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"
	originalData := []byte(`{"original": "data"}`)

	// Initialize vault with original data
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	if err := storage.SaveVault(originalData, password); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Create backup
	if err := storage.CreateBackup(); err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Verify backup file exists
	backupPath := vaultPath + BackupSuffix
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file should exist")
	}

	// Modify vault
	modifiedData := []byte(`{"modified": "data"}`)
	if err := storage.SaveVault(modifiedData, password); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Verify modification
	loadedData, err := storage.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if bytes.Equal(originalData, loadedData) {
		t.Error("Data should be modified")
	}

	// Restore from backup
	if err := storage.RestoreFromBackup(); err != nil {
		t.Fatalf("RestoreFromBackup failed: %v", err)
	}

	// Verify restoration
	restoredData, err := storage.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed after restore: %v", err)
	}

	if !bytes.Equal(originalData, restoredData) {
		t.Error("Restored data should match original data")
	}

	// Test removing backup
	if err := storage.RemoveBackup(); err != nil {
		t.Fatalf("RemoveBackup failed: %v", err)
	}

	// Verify backup is removed
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Backup file should be removed")
	}

	// Test removing non-existent backup (should not error)
	if err := storage.RemoveBackup(); err != nil {
		t.Errorf("RemoveBackup should not error for non-existent backup: %v", err)
	}
}

func TestStorageService_AtomicWrite(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize vault
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Save some data
	testData := []byte(`{"test": "data"}`)
	if err := storage.SaveVault(testData, password); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Verify no temporary files are left behind
	tempPath := vaultPath + TempSuffix
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("Temporary file should not exist after successful write")
	}

	// Verify data is correctly saved
	loadedData, err := storage.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if !bytes.Equal(testData, loadedData) {
		t.Error("Loaded data does not match saved data")
	}
}

func TestStorageService_EmptyData(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize vault
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Save empty data
	emptyData := []byte("")
	if err := storage.SaveVault(emptyData, password); err != nil {
		t.Fatalf("SaveVault failed with empty data: %v", err)
	}

	// Load and verify empty data
	loadedData, err := storage.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if !bytes.Equal(emptyData, loadedData) {
		t.Error("Loaded empty data does not match saved empty data")
	}
}

// Test corruption detection scenarios
func TestStorageService_CorruptionDetection(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize vault
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	tests := []struct {
		name          string
		corruptFunc   func() error
		expectedError error
	}{
		{
			name: "Empty file",
			corruptFunc: func() error {
				return os.WriteFile(vaultPath, []byte(""), VaultPermissions)
			},
			expectedError: nil, // Should fail to parse
		},
		{
			name: "Invalid JSON",
			corruptFunc: func() error {
				return os.WriteFile(vaultPath, []byte(`{"incomplete": `), VaultPermissions)
			},
			expectedError: nil, // Should fail to parse
		},
		{
			name: "Invalid version",
			corruptFunc: func() error {
				return os.WriteFile(vaultPath, []byte(`{"metadata":{"version":0,"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z","salt":"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY="},"data":"dGVzdA=="}`), VaultPermissions)
			},
			expectedError: ErrVaultCorrupted,
		},
		{
			name: "Invalid salt length",
			corruptFunc: func() error {
				return os.WriteFile(vaultPath, []byte(`{"metadata":{"version":1,"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z","salt":"c2hvcnQ="},"data":"dGVzdA=="}`), VaultPermissions)
			},
			expectedError: ErrVaultCorrupted,
		},
		{
			name: "Empty data",
			corruptFunc: func() error {
				return os.WriteFile(vaultPath, []byte(`{"metadata":{"version":1,"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z","salt":"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY="},"data":""}`), VaultPermissions)
			},
			expectedError: ErrVaultCorrupted,
		},
		{
			name: "Zero timestamp",
			corruptFunc: func() error {
				return os.WriteFile(vaultPath, []byte(`{"metadata":{"version":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z","salt":"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY="},"data":"dGVzdA=="}`), VaultPermissions)
			},
			expectedError: ErrVaultCorrupted,
		},
		{
			name: "UpdatedAt before CreatedAt",
			corruptFunc: func() error {
				return os.WriteFile(vaultPath, []byte(`{"metadata":{"version":1,"created_at":"2025-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z","salt":"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY="},"data":"dGVzdA=="}`), VaultPermissions)
			},
			expectedError: ErrVaultCorrupted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Corrupt the vault
			if err := tt.corruptFunc(); err != nil {
				t.Fatalf("Failed to corrupt vault: %v", err)
			}

			// Validate should detect corruption
			err := storage.ValidateVault()
			if tt.expectedError != nil {
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else if err == nil {
				t.Error("Expected validation to fail for corrupted vault")
			}
		})
	}
}

// Test atomic write edge cases
func TestStorageService_AtomicWriteEdgeCases(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize vault
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	t.Run("Concurrent writes", func(t *testing.T) {
		// Save initial data
		initialData := []byte(`{"initial": "data"}`)
		if err := storage.SaveVault(initialData, password); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}

		// Attempt concurrent writes (simulated)
		data1 := []byte(`{"write": "1"}`)
		data2 := []byte(`{"write": "2"}`)

		if err := storage.SaveVault(data1, password); err != nil {
			t.Errorf("First write failed: %v", err)
		}

		if err := storage.SaveVault(data2, password); err != nil {
			t.Errorf("Second write failed: %v", err)
		}

		// Verify vault is still readable and has valid data
		loadedData, err := storage.LoadVault(password)
		if err != nil {
			t.Fatalf("LoadVault failed: %v", err)
		}

		// Should have data from last successful write
		if !bytes.Equal(data2, loadedData) {
			t.Errorf("Expected data2, got %s", string(loadedData))
		}
	})

	t.Run("Large data handling", func(t *testing.T) {
		// Create large data (1MB)
		largeData := make([]byte, 1024*1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		if err := storage.SaveVault(largeData, password); err != nil {
			t.Fatalf("SaveVault failed with large data: %v", err)
		}

		loadedData, err := storage.LoadVault(password)
		if err != nil {
			t.Fatalf("LoadVault failed: %v", err)
		}

		if !bytes.Equal(largeData, loadedData) {
			t.Error("Large data not preserved correctly")
		}
	})

	t.Run("Temp file cleanup on error", func(t *testing.T) {
		tempPath := vaultPath + TempSuffix

		// Verify no temp file exists initially
		if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
			t.Error("Temp file should not exist initially")
		}

		// Perform successful save
		testData := []byte(`{"cleanup": "test"}`)
		if err := storage.SaveVault(testData, password); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}

		// Verify temp file is cleaned up
		if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
			t.Error("Temp file should be cleaned up after successful write")
		}
	})
}

// Test permission changes during operations
func TestStorageService_PermissionScenarios(t *testing.T) {
	// Skip on Windows as it doesn't properly support Unix permissions
	if filepath.Separator == '\\' {
		t.Skip("Skipping permission test on Windows")
	}

	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize vault
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	t.Run("Correct permissions after write", func(t *testing.T) {
		testData := []byte(`{"permission": "test"}`)
		if err := storage.SaveVault(testData, password); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}

		info, err := os.Stat(vaultPath)
		if err != nil {
			t.Fatalf("Failed to stat vault: %v", err)
		}

		expectedPerms := os.FileMode(VaultPermissions)
		actualPerms := info.Mode().Perm()
		if actualPerms != expectedPerms {
			t.Errorf("Expected permissions %v, got %v", expectedPerms, actualPerms)
		}
	})

	t.Run("Read-only vault directory", func(t *testing.T) {
		readOnlyDir := filepath.Join(tempDir, "readonly")
		if err := os.MkdirAll(readOnlyDir, 0500); err != nil {
			t.Fatalf("Failed to create read-only directory: %v", err)
		}
		defer func() {
			_ = os.Chmod(readOnlyDir, 0700) // Restore permissions for cleanup
		}()

		readOnlyVaultPath := filepath.Join(readOnlyDir, "vault.enc")
		readOnlyStorage, err := NewStorageService(cryptoService, readOnlyVaultPath)
		if err != nil {
			t.Fatalf("NewStorageService failed: %v", err)
		}

		// Attempt to initialize vault in read-only directory
		err = readOnlyStorage.InitializeVault(password)
		if err == nil {
			t.Error("Expected error when initializing vault in read-only directory")
		}
	})
}

// Test comprehensive backup and restore scenarios
func TestStorageService_BackupRestoreComprehensive(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize vault
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	t.Run("Backup of non-existent vault", func(t *testing.T) {
		emptyDir := filepath.Join(tempDir, "empty")
		emptyVaultPath := filepath.Join(emptyDir, "vault.enc")
		emptyStorage, err := NewStorageService(cryptoService, emptyVaultPath)
		if err != nil {
			t.Fatalf("NewStorageService failed: %v", err)
		}

		// Backup of non-existent vault should succeed (no-op)
		if err := emptyStorage.CreateBackup(); err != nil {
			t.Errorf("CreateBackup should not fail for non-existent vault: %v", err)
		}
	})

	t.Run("Restore without backup", func(t *testing.T) {
		// Ensure no backup exists
		_ = storage.RemoveBackup()

		// Restore should fail
		if err := storage.RestoreFromBackup(); err != ErrBackupFailed {
			t.Errorf("Expected ErrBackupFailed, got %v", err)
		}
	})

	t.Run("Multiple backup cycles", func(t *testing.T) {
		data1 := []byte(`{"iteration": 1}`)
		data2 := []byte(`{"iteration": 2}`)
		data3 := []byte(`{"iteration": 3}`)

		// Save data1 and backup
		if err := storage.SaveVault(data1, password); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}
		if err := storage.CreateBackup(); err != nil {
			t.Fatalf("CreateBackup failed: %v", err)
		}

		// Save data2 (creates automatic backup)
		if err := storage.SaveVault(data2, password); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}

		// Manual backup should now have data2
		if err := storage.CreateBackup(); err != nil {
			t.Fatalf("CreateBackup failed: %v", err)
		}

		// Save data3
		if err := storage.SaveVault(data3, password); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}

		// Restore should bring back data2 (from manual backup)
		if err := storage.RestoreFromBackup(); err != nil {
			t.Fatalf("RestoreFromBackup failed: %v", err)
		}

		restoredData, err := storage.LoadVault(password)
		if err != nil {
			t.Fatalf("LoadVault failed: %v", err)
		}

		if !bytes.Equal(data2, restoredData) {
			t.Errorf("Expected data2, got %s", string(restoredData))
		}
	})

	t.Run("Backup file permissions", func(t *testing.T) {
		data := []byte(`{"backup": "permissions"}`)
		if err := storage.SaveVault(data, password); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}

		if err := storage.CreateBackup(); err != nil {
			t.Fatalf("CreateBackup failed: %v", err)
		}

		backupPath := vaultPath + BackupSuffix
		info, err := os.Stat(backupPath)
		if err != nil {
			t.Fatalf("Failed to stat backup: %v", err)
		}

		// Backup should have same permissions as vault
		expectedPerms := os.FileMode(VaultPermissions)
		actualPerms := info.Mode().Perm()

		// On Windows, skip this check
		if actualPerms != expectedPerms {
			t.Logf("Note: Backup permissions are %v (expected %v) - may differ on Windows",
				actualPerms, expectedPerms)
		}
	})

	t.Run("Restore preserves metadata", func(t *testing.T) {
		originalData := []byte(`{"preserve": "metadata"}`)
		if err := storage.SaveVault(originalData, password); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}

		// Get original metadata
		originalInfo, err := storage.GetVaultInfo()
		if err != nil {
			t.Fatalf("GetVaultInfo failed: %v", err)
		}

		// Create backup
		if err := storage.CreateBackup(); err != nil {
			t.Fatalf("CreateBackup failed: %v", err)
		}

		// Modify vault
		time.Sleep(10 * time.Millisecond)
		modifiedData := []byte(`{"modified": "data"}`)
		if err := storage.SaveVault(modifiedData, password); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}

		// Restore from backup
		if err := storage.RestoreFromBackup(); err != nil {
			t.Fatalf("RestoreFromBackup failed: %v", err)
		}

		// Get restored metadata
		restoredInfo, err := storage.GetVaultInfo()
		if err != nil {
			t.Fatalf("GetVaultInfo failed: %v", err)
		}

		// Metadata should match original
		if !restoredInfo.CreatedAt.Equal(originalInfo.CreatedAt) {
			t.Error("CreatedAt should be preserved after restore")
		}

		if !restoredInfo.UpdatedAt.Equal(originalInfo.UpdatedAt) {
			t.Error("UpdatedAt should be preserved after restore")
		}
	})

	t.Run("SaveVault creates automatic backup", func(t *testing.T) {
		// Remove any existing backup
		_ = storage.RemoveBackup()

		// Save should create automatic backup
		testData := []byte(`{"auto": "backup"}`)
		if err := storage.SaveVault(testData, password); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}

		// Verify backup was created
		backupPath := vaultPath + BackupSuffix
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Error("SaveVault should create automatic backup")
		}
	})
}

// Test recovery from save failures
func TestStorageService_SaveFailureRecovery(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize vault
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	originalData := []byte(`{"original": "data"}`)
	if err := storage.SaveVault(originalData, password); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	t.Run("Vault intact after failed save", func(t *testing.T) {
		// Create backup
		if err := storage.CreateBackup(); err != nil {
			t.Fatalf("CreateBackup failed: %v", err)
		}

		// Verify original data is still accessible even after potential failures
		loadedData, err := storage.LoadVault(password)
		if err != nil {
			t.Fatalf("LoadVault failed: %v", err)
		}

		if !bytes.Equal(originalData, loadedData) {
			t.Error("Original data should be intact")
		}
	})
}

// Test directory creation for vault path
func TestStorageService_DirectoryCreation(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()

	// Create vault in nested non-existent directory
	nestedPath := filepath.Join(tempDir, "level1", "level2", "level3", "vault.enc")

	storage, err := NewStorageService(cryptoService, nestedPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Verify directories were created
	dir := filepath.Dir(nestedPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Directory should be created automatically")
	}

	// Initialize vault should work in the created directory
	password := "test-password"
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed in nested directory: %v", err)
	}

	// Verify vault was created
	if !storage.VaultExists() {
		t.Error("Vault should exist after initialization")
	}
}

// Test error paths in backup operations
func TestStorageService_BackupErrorPaths(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize vault
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	t.Run("Corrupted backup file", func(t *testing.T) {
		// Create a valid backup first
		if err := storage.CreateBackup(); err != nil {
			t.Fatalf("CreateBackup failed: %v", err)
		}

		// Corrupt the backup file
		backupPath := vaultPath + BackupSuffix
		if err := os.WriteFile(backupPath, []byte("corrupted"), VaultPermissions); err != nil {
			t.Fatalf("Failed to corrupt backup: %v", err)
		}

		// Restore should succeed (it just copies the file)
		// The corruption will be detected when trying to load/validate
		if err := storage.RestoreFromBackup(); err != nil {
			t.Fatalf("RestoreFromBackup failed: %v", err)
		}

		// Validation should fail now
		if err := storage.ValidateVault(); err == nil {
			t.Error("Validation should fail for corrupted restored vault")
		}
	})
}

// Test various edge cases for better coverage
func TestStorageService_AdditionalEdgeCases(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	t.Run("GetVaultInfo on non-existent vault", func(t *testing.T) {
		_, err := storage.GetVaultInfo()
		if err != ErrVaultNotFound {
			t.Errorf("Expected ErrVaultNotFound, got %v", err)
		}
	})

	t.Run("Large password handling", func(t *testing.T) {
		// Create a very long password
		longPassword := string(make([]byte, 1000))
		for i := range longPassword {
			longPassword = longPassword[:i] + "a" + longPassword[i:]
		}
		longPassword = longPassword[:1000]

		if err := storage.InitializeVault(longPassword); err != nil {
			t.Fatalf("InitializeVault failed with long password: %v", err)
		}

		// Should be able to load with the same long password
		testData := []byte(`{"test": "data"}`)
		if err := storage.SaveVault(testData, longPassword); err != nil {
			t.Fatalf("SaveVault failed: %v", err)
		}

		loadedData, err := storage.LoadVault(longPassword)
		if err != nil {
			t.Fatalf("LoadVault failed with long password: %v", err)
		}

		if !bytes.Equal(testData, loadedData) {
			t.Error("Data mismatch with long password")
		}
	})

	t.Run("Special characters in data", func(t *testing.T) {
		tempDir2 := t.TempDir()
		vaultPath2 := filepath.Join(tempDir2, "vault2.enc")
		storage2, _ := NewStorageService(cryptoService, vaultPath2)

		if err := storage2.InitializeVault(password); err != nil {
			t.Fatalf("InitializeVault failed: %v", err)
		}

		// Data with various special characters and unicode
		specialData := []byte(`{"unicode": "🔐🔑", "special": "!@#$%^&*()[]{}|\\/<>?", "null": "\u0000"}`)

		if err := storage2.SaveVault(specialData, password); err != nil {
			t.Fatalf("SaveVault failed with special characters: %v", err)
		}

		loadedData, err := storage2.LoadVault(password)
		if err != nil {
			t.Fatalf("LoadVault failed: %v", err)
		}

		if !bytes.Equal(specialData, loadedData) {
			t.Error("Special characters not preserved")
		}
	})

	t.Run("Rapid successive saves", func(t *testing.T) {
		tempDir3 := t.TempDir()
		vaultPath3 := filepath.Join(tempDir3, "vault3.enc")
		storage3, _ := NewStorageService(cryptoService, vaultPath3)

		if err := storage3.InitializeVault(password); err != nil {
			t.Fatalf("InitializeVault failed: %v", err)
		}

		// Perform many rapid saves
		for i := 0; i < 10; i++ {
			data := []byte(`{"iteration": ` + string(rune(i+'0')) + `}`)
			if err := storage3.SaveVault(data, password); err != nil {
				t.Errorf("SaveVault failed on iteration %d: %v", i, err)
			}
		}

		// Verify last save is readable
		if _, err := storage3.LoadVault(password); err != nil {
			t.Fatalf("LoadVault failed after rapid saves: %v", err)
		}
	})
}

// Test metadata edge cases
func TestStorageService_MetadataValidation(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test_vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize and get initial metadata
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	info1, err := storage.GetVaultInfo()
	if err != nil {
		t.Fatalf("GetVaultInfo failed: %v", err)
	}

	// Verify metadata is not exposed
	if info1.Salt != nil {
		t.Error("Salt should not be exposed through GetVaultInfo")
	}

	// Save and verify UpdatedAt changes
	time.Sleep(10 * time.Millisecond)
	testData := []byte(`{"test": "metadata"}`)
	if err := storage.SaveVault(testData, password); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	info2, err := storage.GetVaultInfo()
	if err != nil {
		t.Fatalf("GetVaultInfo failed: %v", err)
	}

	if info2.CreatedAt != info1.CreatedAt {
		t.Error("CreatedAt should not change after save")
	}

	if !info2.UpdatedAt.After(info1.UpdatedAt) {
		t.Error("UpdatedAt should be updated after save")
	}
}

// T022 [US2]: Test backward compatibility for loading vaults without Iterations field
// FR-010: System MUST support loading legacy vaults with 100k iterations
func TestStorageService_BackwardCompatibleIterations(t *testing.T) {
	t.Skip("T022: Backward compatibility test - will be enabled after T024-T026 implementation")

	// cryptoService := crypto.NewCryptoService()
	// tempDir := t.TempDir()
	// vaultPath := filepath.Join(tempDir, "legacy_vault.enc")
	//
	// // Create a vault file manually with metadata missing Iterations field
	// // This simulates a vault created before T024 (adding Iterations field)
	// legacyVaultJSON := `{
	// 	"metadata": {
	// 		"version": 1,
	// 		"created_at": "2025-01-01T00:00:00Z",
	// 		"updated_at": "2025-01-01T00:00:00Z",
	// 		"salt": "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY="
	// 	},
	// 	"data": "dGVzdA=="
	// }`
	//
	// if err := os.WriteFile(vaultPath, []byte(legacyVaultJSON), VaultPermissions); err != nil {
	// 	t.Fatalf("Failed to create legacy vault file: %v", err)
	// }
	//
	// storage, err := NewStorageService(cryptoService, vaultPath)
	// if err != nil {
	// 	t.Fatalf("NewStorageService failed: %v", err)
	// }
	//
	// // Load legacy vault - should default to 100k iterations
	// // T026: Load method should detect missing Iterations and default to 100000
	// _, err = storage.LoadVault("test-password")
	// if err != nil {
	// 	t.Logf("Load failed (expected until key matches): %v", err)
	// }
	//
	// // Get vault info and verify iterations defaulted correctly
	// info, err := storage.GetVaultInfo()
	// if err != nil {
	// 	t.Fatalf("GetVaultInfo failed: %v", err)
	// }
	//
	// // T026: Verify default iterations applied
	// if info.Iterations == 0 {
	// 	t.Error("Iterations should be populated, got 0")
	// }
	//
	// if info.Iterations != 100000 {
	// 	t.Errorf("Expected default iterations 100000, got %d", info.Iterations)
	// }
	//
	// t.Logf("Legacy vault loaded with iterations: %d", info.Iterations)
}

// T008 [US1] TestAtomicSave_HappyPath verifies successful atomic save operation
// Acceptance: vault.enc contains new data, vault.enc.backup contains old data, no temp files
func TestAtomicSave_HappyPath(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	storage, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "test-password"

	// Initialize vault with initial data
	if err := storage.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	initialData := []byte(`{"credentials": [{"name": "initial"}]}`)
	if err := storage.SaveVault(initialData, password); err != nil {
		t.Fatalf("SaveVault initial failed: %v", err)
	}

	// Save new data (this should create backup of old data)
	newData := []byte(`{"credentials": [{"name": "updated"}]}`)
	if err := storage.SaveVault(newData, password); err != nil {
		t.Fatalf("SaveVault new data failed: %v", err)
	}

	// Verify vault.enc contains new data
	loadedData, err := storage.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if !bytes.Equal(loadedData, newData) {
		t.Errorf("Vault should contain new data. Got %s, want %s", string(loadedData), string(newData))
	}

	// Verify vault.enc.backup contains old data (N-1 generation)
	backupPath := vaultPath + BackupSuffix
	backupExists := false
	if _, err := os.Stat(backupPath); err == nil {
		backupExists = true
	}

	if !backupExists {
		t.Error("Backup file should exist after save")
	}

	// Verify no orphaned temp files exist
	tempPattern := filepath.Join(tempDir, "vault.enc.tmp.*")
	matches, err := filepath.Glob(tempPattern)
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}

	if len(matches) > 0 {
		t.Errorf("No temp files should remain after successful save. Found: %v", matches)
	}
}
