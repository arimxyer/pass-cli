package vault

import (
	"errors"
	"testing"

	"pass-cli/internal/config"
	intsync "pass-cli/internal/sync"
)

// mockSyncService wraps a real sync.Service but tracks calls via a mock executor.
// We use this to verify vault-layer sync integration without real rclone.
type syncTestRecorder struct {
	smartPushCalled bool
	smartPullCalled bool
	smartPushErr    error
	smartPullErr    error
}

func (r *syncTestRecorder) recordPush() {
	r.smartPushCalled = true
}

func (r *syncTestRecorder) recordPull() {
	r.smartPullCalled = true
}

// setupVaultWithMockSync creates a minimal VaultService with a mock sync service
// that records SmartPush/SmartPull calls. It creates a real vault on disk for save operations.
func setupVaultWithMockSync(t *testing.T, recorder *syncTestRecorder) *VaultService {
	t.Helper()

	tmpDir := t.TempDir()
	vs, err := New(tmpDir + "/vault.enc")
	if err != nil {
		t.Fatalf("failed to create VaultService: %v", err)
	}

	// Create a mock executor that records calls
	mock := &vaultSyncMockExecutor{recorder: recorder}

	// Inject sync service with mock executor
	cfg := config.SyncConfig{Enabled: true, Remote: "mock-remote:bucket"}
	vs.syncService = intsync.NewServiceWithExecutor(cfg, mock)

	return vs
}

// vaultSyncMockExecutor implements sync.CommandExecutor to track calls.
type vaultSyncMockExecutor struct {
	recorder *syncTestRecorder
}

func (m *vaultSyncMockExecutor) Run(name string, args ...string) ([]byte, error) {
	// lsjson calls during SmartPush/SmartPull
	if len(args) > 0 && args[0] == "lsjson" {
		m.recorder.recordPull() // lsjson is called during pull check
		return []byte("[]"), m.recorder.smartPullErr
	}
	return []byte("[]"), nil
}

func (m *vaultSyncMockExecutor) RunNoOutput(name string, args ...string) error {
	if len(args) > 0 && args[0] == "sync" {
		m.recorder.recordPush()
		return m.recorder.smartPushErr
	}
	return nil
}

func TestVaultSave_TriggersSmartPush(t *testing.T) {
	recorder := &syncTestRecorder{}
	vs := setupVaultWithMockSync(t, recorder)

	// Initialize vault with a valid password
	masterPass := []byte("TestP@ssw0rd123!")
	err := vs.Initialize(masterPass, false, "", "")
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Unlock vault
	err = vs.Unlock([]byte("TestP@ssw0rd123!"))
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
	defer vs.Lock()

	// Add a credential (which triggers save → SmartPush)
	err = vs.AddCredential("test-service", "user", []byte("s3cretP@ss!"), "", "", "")
	if err != nil {
		t.Fatalf("AddCredential failed: %v", err)
	}

	// SmartPush calls rclone sync via RunNoOutput
	if !recorder.smartPushCalled {
		t.Error("expected SmartPush to be called after save, but it was not")
	}
}

func TestVaultSyncPull_CalledBeforeUnlock(t *testing.T) {
	recorder := &syncTestRecorder{}
	vs := setupVaultWithMockSync(t, recorder)

	// Initialize vault
	masterPass := []byte("TestP@ssw0rd123!")
	err := vs.Initialize(masterPass, false, "", "")
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Call SyncPull (simulates pre-unlock sync)
	err = vs.SyncPull()
	if err != nil {
		t.Fatalf("SyncPull failed: %v", err)
	}

	// SmartPull calls lsjson via Run
	if !recorder.smartPullCalled {
		t.Error("expected SmartPull to be called via SyncPull, but it was not")
	}
}

func TestVaultSave_SkipsPushOnConflict(t *testing.T) {
	recorder := &syncTestRecorder{}
	vs := setupVaultWithMockSync(t, recorder)

	// Initialize and unlock
	masterPass := []byte("TestP@ssw0rd123!")
	err := vs.Initialize(masterPass, false, "", "")
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = vs.Unlock([]byte("TestP@ssw0rd123!"))
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
	defer vs.Lock()

	// Simulate conflict detected
	vs.syncConflictDetected = true

	// Reset recorder to check only the AddCredential save
	recorder.smartPushCalled = false

	// Add credential (triggers save, but push should be skipped)
	err = vs.AddCredential("conflict-test", "user", []byte("s3cretP@ss!"), "", "", "")
	if err != nil {
		t.Fatalf("AddCredential failed: %v", err)
	}

	if recorder.smartPushCalled {
		t.Error("expected SmartPush to be skipped when syncConflictDetected=true, but it was called")
	}
}

// Test that SyncPull sets syncConflictDetected on ErrSyncConflict
func TestVaultSyncPull_SetsConflictFlag(t *testing.T) {
	tmpDir := t.TempDir()
	vs, err := New(tmpDir + "/vault.enc")
	if err != nil {
		t.Fatalf("failed to create VaultService: %v", err)
	}

	// Create a mock that returns conflict from lsjson to trigger conflict path
	// We need a more targeted approach: set up state so SmartPull detects conflict
	cfg := config.SyncConfig{Enabled: true, Remote: "mock-remote:bucket"}
	mock := &conflictMockExecutor{}
	vs.syncService = intsync.NewServiceWithExecutor(cfg, mock)

	// SyncPull should set the conflict flag but not return error
	err = vs.SyncPull()
	if err != nil {
		t.Fatalf("SyncPull should not return error on conflict, got: %v", err)
	}

	if !vs.syncConflictDetected {
		t.Error("expected syncConflictDetected=true after conflict, but it was false")
	}
}

// conflictMockExecutor is designed to make SmartPull return ErrSyncConflict.
// It returns remote metadata that differs from state, and the local file also differs.
type conflictMockExecutor struct{}

func (m *conflictMockExecutor) Run(name string, args ...string) ([]byte, error) {
	// Return metadata that will differ from any saved state
	return []byte(`[{"Path":"vault.enc","Name":"vault.enc","Size":999,"ModTime":"2026-01-29T20:00:00Z","IsDir":false}]`), nil
}

func (m *conflictMockExecutor) RunNoOutput(name string, args ...string) error {
	return errors.New("should not be called during conflict")
}
