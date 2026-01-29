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

func TestVaultSave_DoesNotTriggerPush(t *testing.T) {
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

	// Reset recorder after unlock (Initialize calls save)
	recorder.syncCalled = false

	// Add a credential (triggers save but should NOT push)
	err = vs.AddCredential("test-service", "user", []byte("s3cretP@ss!"), "", "", "")
	if err != nil {
		t.Fatalf("AddCredential failed: %v", err)
	}

	// save() should no longer call SmartPush — push is command-layer only
	if recorder.syncCalled {
		t.Error("expected save() to NOT call rclone sync, but it did")
	}
}

func TestVaultSyncPush_CallsSmartPush(t *testing.T) {
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

	// Call SyncPush explicitly (as command layer would)
	vs.SyncPush()

	// SmartPush should hash the file and push if changed
	// Since Initialize created the vault, the hash differs from empty state → push happens
	if !recorder.syncCalled {
		t.Error("expected SyncPush to call rclone sync, but it did not")
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

func TestVaultSyncPush_SkipsOnConflict(t *testing.T) {
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

	// Reset recorder
	recorder.syncCalled = false

	// SyncPush should skip when conflict is detected
	vs.SyncPush()

	if recorder.syncCalled {
		t.Error("expected SyncPush to skip when syncConflictDetected=true, but it was called")
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

// --- Benchmarks ---
// These demonstrate that save() is purely local I/O (no network delay),
// while SyncPush() incurs the sync cost exactly once.

// slowMockExecutor simulates network latency for rclone operations.
type slowMockExecutor struct {
	delay time.Duration
}

func (m *slowMockExecutor) Run(name string, args ...string) ([]byte, error) {
	time.Sleep(m.delay)
	return []byte("[]"), nil
}

func (m *slowMockExecutor) RunNoOutput(name string, args ...string) error {
	time.Sleep(m.delay)
	return nil
}

func setupBenchVault(b *testing.B, delay time.Duration) *VaultService {
	b.Helper()
	tmpDir := b.TempDir()
	vs, err := New(tmpDir + "/vault.enc")
	if err != nil {
		b.Fatalf("failed to create VaultService: %v", err)
	}

	mock := &slowMockExecutor{delay: delay}
	cfg := config.SyncConfig{Enabled: true, Remote: "mock-remote:bucket"}
	vs.syncService = intsync.NewServiceWithExecutor(cfg, mock)

	masterPass := []byte("BenchP@ssw0rd123!")
	if err := vs.Initialize(masterPass, false, "", ""); err != nil {
		b.Fatalf("Initialize failed: %v", err)
	}
	if err := vs.Unlock([]byte("BenchP@ssw0rd123!")); err != nil {
		b.Fatalf("Unlock failed: %v", err)
	}

	return vs
}

// BenchmarkSave_NoNetworkCost benchmarks save() which should be purely local.
// With 200ms simulated network delay, save() should still be fast because
// it no longer calls SmartPush.
func BenchmarkSave_NoNetworkCost(b *testing.B) {
	vs := setupBenchVault(b, 200*time.Millisecond)
	defer vs.Lock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vs.AddCredential("bench-svc", "user", []byte("s3cretP@ss!"), "", "", "")
	}
}

// BenchmarkSyncPush_IncursNetworkCost benchmarks SyncPush() which does
// call rclone (with simulated 200ms delay per call).
func BenchmarkSyncPush_IncursNetworkCost(b *testing.B) {
	vs := setupBenchVault(b, 200*time.Millisecond)
	defer vs.Lock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vs.SyncPush()
	}
}
