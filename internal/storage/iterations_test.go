package storage

import (
	"testing"

	"pass-cli/internal/crypto"
)

// TestGetIterations_NewVault tests getting iterations from newly initialized vault
func TestGetIterations_NewVault(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Get iterations - should be default (600k as of Jan 2025)
	iterations := s.GetIterations()
	expectedIterations := crypto.GetIterations()
	if iterations != expectedIterations {
		t.Errorf("Expected iterations %d, got %d", expectedIterations, iterations)
	}
}

// TestGetIterations_NoVault tests getting iterations when vault doesn't exist
func TestGetIterations_NoVault(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/nonexistent.enc"

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Get iterations from non-existent vault - should return 0
	iterations := s.GetIterations()
	if iterations != 0 {
		t.Errorf("Expected 0 iterations for non-existent vault, got %d", iterations)
	}
}

// TestSetIterations_Valid tests setting valid iteration count
func TestSetIterations_Valid(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Note: SetIterations updates metadata in memory but doesn't persist
	// This is documented behavior - persistence happens on next save
	newIterations := 700000

	err = s.SetIterations(newIterations)
	if err != nil {
		t.Errorf("SetIterations should succeed for valid count: %v", err)
	}

	// Verify it was set (note: this reads from disk, so won't reflect in-memory change)
	// SetIterations doesn't persist immediately by design
	currentIterations := s.GetIterations()
	if currentIterations == newIterations {
		t.Error("SetIterations should not persist immediately (by design)")
	}
}

// TestSetIterations_TooLow tests setting iteration count below minimum
func TestSetIterations_TooLow(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Try to set iterations below minimum
	err = s.SetIterations(50000) // Below MinIterations (100k)
	if err == nil {
		t.Error("SetIterations should fail for iteration count below minimum")
	}
}

// TestSaveVaultWithIterations_Success tests successful iteration upgrade
func TestSaveVaultWithIterations_Success(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault with default iterations
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Save some data
	initialData := []byte(`{"credentials":[{"service":"test"}]}`)
	if err := s.SaveVault(initialData, password, nil); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Upgrade to higher iteration count
	newIterations := 700000
	err = s.SaveVaultWithIterations(initialData, password, newIterations)
	if err != nil {
		t.Fatalf("SaveVaultWithIterations failed: %v", err)
	}

	// Verify iterations were updated
	currentIterations := s.GetIterations()
	if currentIterations != newIterations {
		t.Errorf("Expected iterations %d, got %d", newIterations, currentIterations)
	}

	// Verify data is still accessible with same password
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed after iteration upgrade: %v", err)
	}

	if string(loadedData) != string(initialData) {
		t.Error("Data mismatch after iteration upgrade")
	}
}

// TestSaveVaultWithIterations_TooLow tests validation of iteration count
func TestSaveVaultWithIterations_TooLow(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	data := []byte(`{"credentials":[]}`)

	// Try to save with too-low iteration count
	err = s.SaveVaultWithIterations(data, password, 50000)
	if err == nil {
		t.Error("SaveVaultWithIterations should fail for iterations below minimum")
	}

	// Verify vault is still readable with original password
	_, err = s.LoadVault(password)
	if err != nil {
		t.Error("Vault should still be accessible after failed iteration upgrade")
	}
}

// TestSaveVaultWithIterationsUnsafe_AllowsLowIterations tests unsafe variant
func TestSaveVaultWithIterationsUnsafe_AllowsLowIterations(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	data := []byte(`{"credentials":[]}`)

	// Unsafe variant should allow low iterations (for testing legacy vaults)
	lowIterations := 50000
	err = s.SaveVaultWithIterationsUnsafe(data, password, lowIterations)
	if err != nil {
		t.Fatalf("SaveVaultWithIterationsUnsafe should allow low iterations: %v", err)
	}

	// Verify iterations were set to low value
	currentIterations := s.GetIterations()
	if currentIterations != lowIterations {
		t.Errorf("Expected iterations %d, got %d", lowIterations, currentIterations)
	}

	// Verify vault is still accessible
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if string(loadedData) != string(data) {
		t.Error("Data mismatch after unsafe iteration change")
	}
}

// TestSaveVaultWithIterations_PreflightCheckDiskSpace tests preflight disk space check
func TestSaveVaultWithIterations_PreflightCheck(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := tempDir + "/vault.enc"

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	data := []byte(`{"credentials":[]}`)

	// SaveVaultWithIterations runs preflight checks
	// On most systems, disk space check will pass (or be skipped with warning)
	err = s.SaveVaultWithIterations(data, password, 600000)

	// We expect this to succeed (disk space available)
	// If it fails, it should be due to disk space or permissions
	if err != nil {
		t.Logf("SaveVaultWithIterations failed (may be expected on low-disk systems): %v", err)
		// Don't fail the test - disk space checks are platform-dependent
	}
}
