package vault

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"pass-cli/internal/config"
	intsync "pass-cli/internal/sync"
)

// mockSyncService wraps a real sync.Service but tracks calls via a mock executor.
// We use this to verify vault-layer sync integration without real rclone.
type syncTestRecorder struct {
	// lsjsonCalled tracks whether Run was called with "lsjson" (used by both SmartPull and SmartPush)
	lsjsonCalled bool
	// syncCalled tracks whether RunNoOutput was called with "sync" (the actual rclone sync)
	syncCalled   bool
	lsjsonErr    error
	syncErr      error
}

func (r *syncTestRecorder) recordLsjson() {
	r.lsjsonCalled = true
}

func (r *syncTestRecorder) recordSync() {
	r.syncCalled = true
}

func skipIfNoRclone(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("rclone"); err != nil {
		t.Skip("rclone not installed, skipping vault sync test")
	}
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
	// lsjson calls during SmartPush/SmartPull (CheckRemoteMetadata)
	if len(args) > 0 && args[0] == "lsjson" {
		m.recorder.recordLsjson()
		return []byte("[]"), m.recorder.lsjsonErr
	}
	return []byte("[]"), nil
}

func (m *vaultSyncMockExecutor) RunNoOutput(name string, args ...string) error {
	if len(args) > 0 && args[0] == "sync" {
		m.recorder.recordSync()
		return m.recorder.syncErr
	}
	return nil
}

func TestVaultSave_TriggersSmartPush(t *testing.T) {
	skipIfNoRclone(t)

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
	if !recorder.syncCalled {
		t.Error("expected rclone sync to be called after save, but it was not")
	}
}

func TestVaultSyncPull_CalledBeforeUnlock(t *testing.T) {
	skipIfNoRclone(t)

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

	// SmartPull calls lsjson via Run to check remote metadata
	if !recorder.lsjsonCalled {
		t.Error("expected lsjson to be called via SyncPull, but it was not")
	}
}

func TestVaultSave_SkipsPushOnConflict(t *testing.T) {
	skipIfNoRclone(t)

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
	recorder.syncCalled = false

	// Add credential (triggers save, but push should be skipped)
	err = vs.AddCredential("conflict-test", "user", []byte("s3cretP@ss!"), "", "", "")
	if err != nil {
		t.Fatalf("AddCredential failed: %v", err)
	}

	if recorder.syncCalled {
		t.Error("expected rclone sync to be skipped when syncConflictDetected=true, but it was called")
	}
}

// Test that SyncPull sets syncConflictDetected on ErrSyncConflict
func TestVaultSyncPull_SetsConflictFlag(t *testing.T) {
	skipIfNoRclone(t)

	tmpDir := t.TempDir()
	vaultPath := filepath.Join(tmpDir, "vault.enc")

	// 1. Create a vault file on disk with known content
	vaultContent := []byte("existing vault data - local version")
	if err := os.WriteFile(vaultPath, vaultContent, 0600); err != nil {
		t.Fatalf("failed to write vault file: %v", err)
	}

	// 2. Compute hash of the vault file, then save state with a DIFFERENT LastPushHash
	//    so SmartPull sees local as "changed"
	state := &intsync.SyncState{
		LastPushHash:  "0000000000000000000000000000000000000000000000000000000000000000",
		LastPushTime:  time.Now().Add(-time.Hour),
		RemoteModTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		RemoteSize:    100,
	}
	if err := intsync.SaveState(tmpDir, state); err != nil {
		t.Fatalf("failed to save sync state: %v", err)
	}

	// 3. Create VaultService
	vs, err := New(vaultPath)
	if err != nil {
		t.Fatalf("failed to create VaultService: %v", err)
	}

	// 4. Create mock executor that returns remote metadata differing from saved state
	//    (different ModTime/Size so remote appears "changed")
	remoteModTime := time.Date(2026, 1, 29, 20, 0, 0, 0, time.UTC)
	remoteMetadata := []struct {
		Path    string    `json:"Path"`
		Name    string    `json:"Name"`
		Size    int64     `json:"Size"`
		ModTime time.Time `json:"ModTime"`
		IsDir   bool      `json:"IsDir"`
	}{
		{Path: "vault.enc", Name: "vault.enc", Size: 999, ModTime: remoteModTime, IsDir: false},
	}
	metadataJSON, err := json.Marshal(remoteMetadata)
	if err != nil {
		t.Fatalf("failed to marshal metadata: %v", err)
	}

	mock := &conflictMockExecutor{lsjsonOutput: metadataJSON}
	cfg := config.SyncConfig{Enabled: true, Remote: "mock-remote:bucket"}
	vs.syncService = intsync.NewServiceWithExecutor(cfg, mock)

	// 5. SyncPull should detect conflict (remote changed + local changed) and set flag
	err = vs.SyncPull()
	if err != nil {
		t.Fatalf("SyncPull should not return error on conflict, got: %v", err)
	}

	if !vs.syncConflictDetected {
		t.Error("expected syncConflictDetected=true after conflict, but it was false")
	}
}

// conflictMockExecutor is designed to make SmartPull return ErrSyncConflict.
// It returns remote metadata that differs from saved state, while local file also differs from LastPushHash.
type conflictMockExecutor struct {
	lsjsonOutput []byte
}

func (m *conflictMockExecutor) Run(name string, args ...string) ([]byte, error) {
	return m.lsjsonOutput, nil
}

func (m *conflictMockExecutor) RunNoOutput(name string, args ...string) error {
	return errors.New("should not be called during conflict")
}
