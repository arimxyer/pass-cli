package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pass-cli/internal/crypto"
)

// TestRandomHexSuffix tests hex suffix generation
func TestRandomHexSuffix(t *testing.T) {
	// Test normal case
	suffix := randomHexSuffix(12)
	if len(suffix) != 12 {
		t.Errorf("Expected suffix length 12, got %d", len(suffix))
	}

	// Verify it's hex
	for _, c := range suffix {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Suffix contains non-hex character: %c", c)
		}
	}

	// Test uniqueness (very unlikely to collide)
	suffix2 := randomHexSuffix(12)
	if suffix == suffix2 {
		t.Error("Two suffixes should not be identical")
	}
}

// TestWriteToTempFile_Success tests successful temp file writing
func TestWriteToTempFile_Success(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	testData := []byte("test data for temp file")
	tempPath := filepath.Join(tempDir, "test.tmp")

	err = s.writeToTempFile(tempPath, testData)
	if err != nil {
		t.Fatalf("writeToTempFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		t.Error("Temp file was not created")
	}

	// Verify content
	content, err := os.ReadFile(tempPath)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("File content mismatch: expected %s, got %s", testData, content)
	}

	// Verify permissions (Unix-like systems)
	info, err := os.Stat(tempPath)
	if err != nil {
		t.Fatalf("Failed to stat temp file: %v", err)
	}

	mode := info.Mode().Perm()
	// On Windows, permissions work differently, so we just verify file exists
	if mode != 0600 && os.Getenv("GOOS") != "windows" {
		t.Logf("Warning: File permissions are %o, expected 0600", mode)
	}
}

// TestWriteToTempFile_InvalidPath tests error handling for invalid paths
func TestWriteToTempFile_InvalidPath(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	testData := []byte("test data")

	// Try to write to invalid path (non-existent directory)
	invalidPath := filepath.Join(tempDir, "nonexistent", "dir", "test.tmp")

	err = s.writeToTempFile(invalidPath, testData)
	if err == nil {
		t.Error("writeToTempFile should fail for non-existent directory")
	}
}

// TestVerifyTempFile_Success tests successful temp file verification
func TestVerifyTempFile_Success(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault first
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Create a valid encrypted vault file
	testData := []byte(`{"credentials":[]}`)
	encryptedVault, err := s.loadEncryptedVault()
	if err != nil {
		t.Fatalf("loadEncryptedVault failed: %v", err)
	}

	encryptedVault.Metadata.UpdatedAt = encryptedVault.Metadata.CreatedAt
	encryptedData, err := s.prepareEncryptedData(testData, encryptedVault.Metadata, password)
	if err != nil {
		t.Fatalf("prepareEncryptedData failed: %v", err)
	}

	// Write to temp file
	tempPath := filepath.Join(tempDir, "test.tmp")
	if err := os.WriteFile(tempPath, encryptedData, 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify temp file
	err = s.verifyTempFile(tempPath, password)
	if err != nil {
		t.Errorf("verifyTempFile should succeed for valid encrypted file: %v", err)
	}
}

// TestVerifyTempFile_InvalidFile tests verification failure for corrupted files
func TestVerifyTempFile_InvalidFile(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"

	// Initialize vault
	if err := s.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Test 1: Non-existent file
	err = s.verifyTempFile("/nonexistent/file.tmp", password)
	if err == nil {
		t.Error("verifyTempFile should fail for non-existent file")
	}

	// Test 2: Invalid JSON
	invalidPath := filepath.Join(tempDir, "invalid.tmp")
	if err := os.WriteFile(invalidPath, []byte("not valid json"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = s.verifyTempFile(invalidPath, password)
	if err == nil {
		t.Error("verifyTempFile should fail for invalid JSON")
	}

	// Test 3: Wrong password
	validPath := filepath.Join(tempDir, "valid.tmp")
	testData := []byte(`{"credentials":[]}`)
	encryptedVault, _ := s.loadEncryptedVault()
	encryptedData, _ := s.prepareEncryptedData(testData, encryptedVault.Metadata, password)
	if err := os.WriteFile(validPath, encryptedData, 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = s.verifyTempFile(validPath, "WrongPassword!")
	if err == nil {
		t.Error("verifyTempFile should fail with wrong password")
	}
}

// TestAtomicRename_Success tests successful rename operation
func TestAtomicRename_Success(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Create source file
	srcPath := filepath.Join(tempDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("test content"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Rename to destination
	dstPath := filepath.Join(tempDir, "destination.txt")
	err = s.atomicRename(srcPath, dstPath)
	if err != nil {
		t.Fatalf("atomicRename failed: %v", err)
	}

	// Verify source removed
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("Source file should be removed after rename")
	}

	// Verify destination exists
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("Destination file should exist after rename")
	}

	// Verify content
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Content mismatch after rename")
	}
}

// TestAtomicRename_SourceNotExist tests rename when source doesn't exist
func TestAtomicRename_SourceNotExist(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	srcPath := filepath.Join(tempDir, "nonexistent.txt")
	dstPath := filepath.Join(tempDir, "destination.txt")

	err = s.atomicRename(srcPath, dstPath)
	if err == nil {
		t.Error("atomicRename should fail when source doesn't exist")
	}
}

// TestAtomicRename_DestinationExists tests rename when destination exists (should replace)
func TestAtomicRename_DestinationExists(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Create source and destination files
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "destination.txt")

	if err := os.WriteFile(srcPath, []byte("new content"), 0600); err != nil {
		t.Fatalf("WriteFile source failed: %v", err)
	}

	if err := os.WriteFile(dstPath, []byte("old content"), 0600); err != nil {
		t.Fatalf("WriteFile destination failed: %v", err)
	}

	// Rename (should replace destination)
	err = s.atomicRename(srcPath, dstPath)
	if err != nil {
		t.Fatalf("atomicRename failed: %v", err)
	}

	// Verify new content
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination: %v", err)
	}

	if string(content) != "new content" {
		t.Errorf("Destination should have new content after rename")
	}
}

// TestCleanupTempFile tests temporary file cleanup
func TestCleanupTempFile(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Test 1: Cleanup existing file
	tempPath := filepath.Join(tempDir, "test.tmp")
	if err := os.WriteFile(tempPath, []byte("test"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = s.cleanupTempFile(tempPath)
	if err != nil {
		t.Errorf("cleanupTempFile should not fail: %v", err)
	}

	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("Temp file should be deleted")
	}

	// Test 2: Cleanup non-existent file (should succeed silently)
	err = s.cleanupTempFile(filepath.Join(tempDir, "nonexistent.tmp"))
	if err != nil {
		t.Errorf("cleanupTempFile should succeed for non-existent file: %v", err)
	}
}

// TestCleanupOrphanedTempFiles tests orphaned temp file cleanup
func TestCleanupOrphanedTempFiles(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Create some orphaned temp files
	tempFiles := []string{
		filepath.Join(tempDir, "vault.enc.tmp.20250109-120000.abc123"),
		filepath.Join(tempDir, "vault.enc.tmp.20250109-120001.def456"),
		filepath.Join(tempDir, "vault.enc.tmp.20250109-120002.ghi789"),
	}

	for _, path := range tempFiles {
		if err := os.WriteFile(path, []byte("orphaned"), 0600); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
	}

	// Create non-temp file that should NOT be deleted
	normalFile := filepath.Join(tempDir, "normal.txt")
	if err := os.WriteFile(normalFile, []byte("keep"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Run cleanup (excluding one specific temp file)
	s.cleanupOrphanedTempFiles(tempFiles[0])

	// Verify excluded file still exists
	if _, err := os.Stat(tempFiles[0]); os.IsNotExist(err) {
		t.Error("Excluded temp file should not be deleted")
	}

	// Verify other temp files are deleted
	for i, path := range tempFiles[1:] {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Orphaned temp file %d should be deleted", i+1)
		}
	}

	// Verify normal file is not touched
	if _, err := os.Stat(normalFile); os.IsNotExist(err) {
		t.Error("Normal file should not be deleted")
	}
}

// TestCleanupOrphanedTempFiles_EmptyDirectory tests cleanup with no temp files
func TestCleanupOrphanedTempFiles_EmptyDirectory(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Cleanup with no temp files (should not error)
	s.cleanupOrphanedTempFiles("")

	// Directory should still exist
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Directory should still exist after cleanup")
	}
}

// TestGenerateTempFileName tests temp filename generation
func TestGenerateTempFileName(t *testing.T) {
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	s, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	// Generate temp filename
	tempPath := s.generateTempFileName()

	// Verify format: vault.enc.tmp.TIMESTAMP.SUFFIX
	if !strings.HasPrefix(tempPath, vaultPath+".tmp.") {
		t.Errorf("Temp filename should start with %s.tmp., got %s", vaultPath, tempPath)
	}

	// Verify uniqueness
	tempPath2 := s.generateTempFileName()
	if tempPath == tempPath2 {
		t.Error("Two generated temp filenames should not be identical")
	}
}
