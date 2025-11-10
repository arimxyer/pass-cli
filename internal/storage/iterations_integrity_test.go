package storage

import (
	"bytes"
	"testing"

	"pass-cli/internal/crypto"
)

// TestSaveVaultWithIterations_NoOpChange tests SaveVaultWithIterations with current iterations (no-op)
func TestSaveVaultWithIterations_NoOpChange(t *testing.T) {
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

	currentIterations := s.GetIterations()
	t.Logf("Current iterations: %d", currentIterations)

	// Save data
	data := []byte(`{"credentials":[{"service":"test","username":"user","password":"pass"}]}`)
	if err := s.SaveVault(data, password, nil); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Save with SAME iteration count (no-op migration)
	err = s.SaveVaultWithIterations(data, password, currentIterations)
	if err != nil {
		t.Fatalf("SaveVaultWithIterations with current iterations should succeed: %v", err)
	}

	// Verify iterations unchanged
	if s.GetIterations() != currentIterations {
		t.Errorf("Iterations should remain unchanged for no-op migration")
	}

	// Verify data accessible and unchanged
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed after no-op migration: %v", err)
	}

	if !bytes.Equal(loadedData, data) {
		t.Error("Data mismatch after no-op iteration migration")
	}
}

// TestSaveVaultWithIterations_SequentialChanges tests multiple iteration changes
func TestSaveVaultWithIterations_SequentialChanges(t *testing.T) {
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

	// Test data
	data := []byte(`{"credentials":[{"service":"bank","username":"alice","password":"secret123"}]}`)

	// Test sequence of iteration changes
	iterationSequence := []int{700000, 800000, 600000, 750000}

	for _, targetIterations := range iterationSequence {
		t.Logf("Changing iterations to %d", targetIterations)

		err = s.SaveVaultWithIterations(data, password, targetIterations)
		if err != nil {
			t.Fatalf("SaveVaultWithIterations(%d) failed: %v", targetIterations, err)
		}

		// Verify iterations were updated
		currentIterations := s.GetIterations()
		if currentIterations != targetIterations {
			t.Errorf("Expected iterations %d, got %d", targetIterations, currentIterations)
		}

		// Verify data is still accessible
		loadedData, err := s.LoadVault(password)
		if err != nil {
			t.Fatalf("LoadVault failed after iteration change to %d: %v", targetIterations, err)
		}

		if !bytes.Equal(loadedData, data) {
			t.Errorf("Data mismatch after iteration change to %d", targetIterations)
		}
	}
}

// TestSaveVaultWithIterations_LargeDataIntegrity tests data integrity with large data
func TestSaveVaultWithIterations_LargeDataIntegrity(t *testing.T) {
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

	// Create large data (simulating vault with many credentials)
	largeData := []byte(`{"credentials":[`)
	for i := 0; i < 100; i++ {
		if i > 0 {
			largeData = append(largeData, []byte(",")...)
		}
		credential := []byte(`{"service":"service` + string(rune('0'+i%10)) + `","username":"user` + string(rune('0'+i%10)) + `","password":"pass` + string(rune('0'+i%10)) + `"}`)
		largeData = append(largeData, credential...)
	}
	largeData = append(largeData, []byte("]}")...)

	t.Logf("Large data size: %d bytes", len(largeData))

	// Save initial data
	if err := s.SaveVault(largeData, password, nil); err != nil {
		t.Fatalf("SaveVault failed: %v", err)
	}

	// Change iterations
	newIterations := 700000
	err = s.SaveVaultWithIterations(largeData, password, newIterations)
	if err != nil {
		t.Fatalf("SaveVaultWithIterations failed with large data: %v", err)
	}

	// Verify data integrity
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if !bytes.Equal(loadedData, largeData) {
		t.Error("Large data mismatch after iteration change")
		t.Logf("Expected size: %d, got: %d", len(largeData), len(loadedData))
	}
}

// TestSaveVaultWithIterations_MinIterations tests boundary condition with MinIterations
func TestSaveVaultWithIterations_MinIterations(t *testing.T) {
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

	// Save with exactly MinIterations (should succeed)
	minIterations := crypto.MinIterations
	err = s.SaveVaultWithIterations(data, password, minIterations)
	if err != nil {
		t.Fatalf("SaveVaultWithIterations should accept MinIterations: %v", err)
	}

	// Verify iterations were set
	if s.GetIterations() != minIterations {
		t.Errorf("Expected iterations %d, got %d", minIterations, s.GetIterations())
	}

	// Verify data accessible
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if !bytes.Equal(loadedData, data) {
		t.Error("Data mismatch after setting MinIterations")
	}
}

// TestSaveVaultWithIterations_JustBelowMin tests validation with iterations just below minimum
func TestSaveVaultWithIterations_JustBelowMin(t *testing.T) {
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

	// Try to save with iterations just below minimum
	belowMin := crypto.MinIterations - 1
	err = s.SaveVaultWithIterations(data, password, belowMin)
	if err == nil {
		t.Error("SaveVaultWithIterations should fail for iterations below minimum")
	}

	// Verify vault is still accessible with original iterations
	_, err = s.LoadVault(password)
	if err != nil {
		t.Error("Vault should remain accessible after failed iteration change")
	}
}

// TestSaveVaultWithIterations_EmptyData tests iteration change with empty vault
func TestSaveVaultWithIterations_EmptyData(t *testing.T) {
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

	// Empty vault data
	emptyData := []byte(`{"credentials":[]}`)

	// Change iterations on empty vault
	newIterations := 700000
	err = s.SaveVaultWithIterations(emptyData, password, newIterations)
	if err != nil {
		t.Fatalf("SaveVaultWithIterations should work with empty data: %v", err)
	}

	// Verify iterations were updated
	if s.GetIterations() != newIterations {
		t.Errorf("Expected iterations %d, got %d", newIterations, s.GetIterations())
	}

	// Verify empty vault is accessible
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed: %v", err)
	}

	if !bytes.Equal(loadedData, emptyData) {
		t.Error("Empty data mismatch")
	}
}

// TestSaveVaultWithIterations_DowngradeIterations tests downgrading iterations (security regression)
func TestSaveVaultWithIterations_DowngradeIterations(t *testing.T) {
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

	// Upgrade to high iterations
	highIterations := 800000
	data := []byte(`{"credentials":[{"service":"test"}]}`)
	err = s.SaveVaultWithIterations(data, password, highIterations)
	if err != nil {
		t.Fatalf("SaveVaultWithIterations failed: %v", err)
	}

	// Verify high iterations set
	if s.GetIterations() != highIterations {
		t.Errorf("Expected %d iterations, got %d", highIterations, s.GetIterations())
	}

	// Downgrade to lower (but still valid) iterations
	lowerIterations := 600000
	err = s.SaveVaultWithIterations(data, password, lowerIterations)
	if err != nil {
		t.Fatalf("SaveVaultWithIterations should allow downgrade to valid iterations: %v", err)
	}

	// Verify downgrade successful
	if s.GetIterations() != lowerIterations {
		t.Errorf("Expected %d iterations after downgrade, got %d", lowerIterations, s.GetIterations())
	}

	// Verify data still accessible
	loadedData, err := s.LoadVault(password)
	if err != nil {
		t.Fatalf("LoadVault failed after iteration downgrade: %v", err)
	}

	if !bytes.Equal(loadedData, data) {
		t.Error("Data mismatch after iteration downgrade")
	}
}
