package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arimxyer/pass-cli/internal/crypto"
)

// TestSaveVault_ErrorMessage_VerificationFailed verifies FR-011 compliance
// for verification failures: must include actionable guidance
func TestSaveVault_ErrorMessage_VerificationFailed(t *testing.T) {
	// Setup: Create vault that will fail verification
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	service, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"
	if err := service.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Create spy filesystem that will cause verification to fail on temp files only
	spy := NewSpyFileSystem()
	wrapper := &verificationFailingSpy{spy: spy}
	service.fs = wrapper

	// Execute: Try to save with bad verification
	data := []byte(`{"credentials":[]}`)
	err = service.SaveVault(data, password, nil)

	// Verify: Error message must contain FR-011 required elements
	if err == nil {
		t.Fatal("Expected error for verification failure, got nil")
	}

	errMsg := err.Error()

	// FR-011 requirement 1: Specific failure reason
	if !strings.Contains(errMsg, "verification") {
		t.Errorf("Error missing 'verification' keyword: %v", errMsg)
	}

	// FR-011 requirement 2: Vault state confirmation
	if !strings.Contains(errMsg, "Your vault was not modified") {
		t.Errorf("Error missing vault status confirmation: %v", errMsg)
	}

	// FR-011 requirement 3: Actionable guidance
	if !strings.Contains(errMsg, "password") {
		t.Errorf("Error missing actionable guidance about password: %v", errMsg)
	}
}

// TestSaveVault_ErrorMessage_DiskSpaceExhausted verifies FR-011 for disk space errors
func TestSaveVault_ErrorMessage_DiskSpaceExhausted(t *testing.T) {
	// Setup
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	service, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"
	if err := service.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Create spy that simulates disk full when writing temp file
	// Note: SaveVault loads the vault first (using ReadFile), then writes temp file (using OpenFile)
	spy := NewSpyFileSystem()
	spy.failOpenFileWithErrFunc = func(path string) error {
		// Always fail with disk space error - this should cause writeToTempFile to fail
		return ErrDiskSpaceExhausted
	}
	service.fs = spy

	// Execute
	data := []byte(`{"credentials":[]}`)
	err = service.SaveVault(data, password, nil)

	// Verify FR-011 requirements
	if err == nil {
		t.Fatal("Expected error for disk space exhaustion, got nil")
	}

	errMsg := err.Error()

	// Must mention disk space
	if !strings.Contains(errMsg, "disk space") {
		t.Errorf("Error missing 'disk space' keyword: %v", errMsg)
	}

	// Must confirm vault unchanged
	if !strings.Contains(errMsg, "Your vault was not modified") {
		t.Errorf("Error missing vault status: %v", errMsg)
	}

	// Must provide actionable guidance (free up space)
	if !strings.Contains(errMsg, "Free up") || !strings.Contains(errMsg, "MB") {
		t.Errorf("Error missing actionable guidance to free space: %v", errMsg)
	}
}

// TestSaveVault_ErrorMessage_PermissionDenied verifies FR-011 for permission errors
func TestSaveVault_ErrorMessage_PermissionDenied(t *testing.T) {
	// Setup
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	service, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"
	if err := service.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Spy that fails rename due to permissions
	spy := NewSpyFileSystem()
	spy.failRenameWithErr = ErrPermissionDenied
	spy.failRenameAt = 1 // Fail first rename (vault → backup)
	service.fs = spy

	// Execute
	data := []byte(`{"credentials":[]}`)
	err = service.SaveVault(data, password, nil)

	// Verify FR-011
	if err == nil {
		t.Fatal("Expected error for permission denied, got nil")
	}

	errMsg := err.Error()

	if !strings.Contains(errMsg, "permission") {
		t.Errorf("Error missing 'permission' keyword: %v", errMsg)
	}

	if !strings.Contains(errMsg, "Your vault was not modified") {
		t.Errorf("Error missing vault status: %v", errMsg)
	}

	// Must tell user to check permissions
	if !strings.Contains(errMsg, "Check file permissions") {
		t.Errorf("Error missing actionable guidance to check permissions: %v", errMsg)
	}
}

// TestSaveVault_ErrorMessage_FilesystemNotAtomic verifies FR-011 for filesystem errors
func TestSaveVault_ErrorMessage_FilesystemNotAtomic(t *testing.T) {
	// Setup
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	service, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"
	if err := service.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Spy that fails rename with filesystem error
	spy := NewSpyFileSystem()
	spy.failRenameWithErr = ErrFilesystemNotAtomic
	spy.failRenameAt = 1
	service.fs = spy

	// Execute
	data := []byte(`{"credentials":[]}`)
	err = service.SaveVault(data, password, nil)

	// Verify FR-011
	if err == nil {
		t.Fatal("Expected error for filesystem error, got nil")
	}

	errMsg := err.Error()

	if !strings.Contains(errMsg, "filesystem") {
		t.Errorf("Error missing 'filesystem' keyword: %v", errMsg)
	}

	if !strings.Contains(errMsg, "Your vault was not modified") {
		t.Errorf("Error missing vault status: %v", errMsg)
	}

	// Must guide user to use local filesystem
	if !strings.Contains(errMsg, "local filesystem") {
		t.Errorf("Error missing actionable guidance about local filesystem: %v", errMsg)
	}
}

// TestSaveVault_ErrorMessage_CriticalRenameFail verifies FR-011 for critical failures
// during the second rename (temp → vault), which triggers rollback
func TestSaveVault_ErrorMessage_CriticalRenameFail(t *testing.T) {
	// Setup
	cryptoService := crypto.NewCryptoService()
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.enc")

	service, err := NewStorageService(cryptoService, vaultPath)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	password := "TestPassword123!"
	if err := service.InitializeVault(password); err != nil {
		t.Fatalf("InitializeVault failed: %v", err)
	}

	// Spy that fails SECOND rename (temp → vault) but succeeds on rollback
	spy := NewSpyFileSystem()
	spy.failRenameAt = 2 // Second rename fails
	spy.failRenameWithErr = errors.New("filesystem error during final commit")
	service.fs = spy

	// Execute
	data := []byte(`{"credentials":[]}`)
	err = service.SaveVault(data, password, nil)

	// Verify FR-011 for critical path
	if err == nil {
		t.Fatal("Expected error for critical rename failure, got nil")
	}

	errMsg := err.Error()

	// Must be marked as CRITICAL
	if !strings.Contains(errMsg, "CRITICAL") {
		t.Errorf("Error missing 'CRITICAL' severity marker: %v", errMsg)
	}

	// Must mention automatic restore attempt
	if !strings.Contains(errMsg, "restore") || !strings.Contains(errMsg, "backup") {
		t.Errorf("Error missing information about automatic restore: %v", errMsg)
	}

	// Must provide guidance for recovery
	if !strings.Contains(errMsg, "verify") || !strings.Contains(errMsg, "integrity") {
		t.Errorf("Error missing actionable guidance to verify vault: %v", errMsg)
	}
}

// verificationFailingSpy wraps a spy and makes ReadFile fail for temp files
type verificationFailingSpy struct {
	spy *spyFileSystem
}

func (v *verificationFailingSpy) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return v.spy.OpenFile(name, flag, perm)
}

func (v *verificationFailingSpy) ReadFile(name string) ([]byte, error) {
	if strings.Contains(name, ".tmp.") {
		return nil, ErrVerificationFailed
	}
	return v.spy.ReadFile(name)
}

func (v *verificationFailingSpy) WriteFile(name string, data []byte, perm os.FileMode) error {
	return v.spy.WriteFile(name, data, perm)
}

func (v *verificationFailingSpy) Remove(name string) error {
	return v.spy.Remove(name)
}

func (v *verificationFailingSpy) Rename(oldpath, newpath string) error {
	return v.spy.Rename(oldpath, newpath)
}

func (v *verificationFailingSpy) MkdirAll(path string, perm os.FileMode) error {
	return v.spy.MkdirAll(path, perm)
}

func (v *verificationFailingSpy) Stat(name string) (os.FileInfo, error) {
	return v.spy.Stat(name)
}

func (v *verificationFailingSpy) Glob(pattern string) ([]string, error) {
	return v.spy.Glob(pattern)
}
