//go:build integration

package sync

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arimxyer/pass-cli/internal/config"
)

// setupMockRclone creates a mock rclone script in a temp directory and returns
// the directory (to prepend to PATH), a log file path, and a lsjson response file path.
func setupMockRclone(t *testing.T) (binDir, logFile, responseFile string) {
	t.Helper()
	binDir = t.TempDir()
	logFile = filepath.Join(binDir, "rclone.log")
	responseFile = filepath.Join(binDir, "lsjson_response.json")

	script := `#!/bin/bash
echo "$@" >> "` + logFile + `"
if [[ "$1" == "lsjson" ]]; then
    cat "` + responseFile + `"
fi
`
	scriptPath := filepath.Join(binDir, "rclone")
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write mock rclone: %v", err)
	}

	// Write empty default response
	if err := os.WriteFile(responseFile, []byte("[]"), 0600); err != nil {
		t.Fatalf("failed to write default response: %v", err)
	}

	return binDir, logFile, responseFile
}

func setLsjsonResponse(t *testing.T, responseFile string, files []RemoteFileInfo) {
	t.Helper()
	data, err := json.Marshal(files)
	if err != nil {
		t.Fatalf("failed to marshal lsjson response: %v", err)
	}
	if err := os.WriteFile(responseFile, data, 0600); err != nil {
		t.Fatalf("failed to write lsjson response: %v", err)
	}
}

func readRcloneLog(t *testing.T, logFile string) []string {
	t.Helper()
	data, err := os.ReadFile(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("failed to read rclone log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}

func clearRcloneLog(t *testing.T, logFile string) {
	t.Helper()
	_ = os.Remove(logFile)
}

func TestIntegration_SmartPush_ThenSmartPull(t *testing.T) {
	binDir, logFile, responseFile := setupMockRclone(t)
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+":"+origPath)

	vaultDir := t.TempDir()
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	if err := os.WriteFile(vaultPath, []byte("vault-content-v1"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.SyncConfig{Enabled: true, Remote: "mock-remote:bucket"}
	svc := NewService(cfg)

	// Set initial lsjson response for post-push metadata query
	pushTime := time.Date(2026, 1, 29, 10, 0, 0, 0, time.UTC)
	setLsjsonResponse(t, responseFile, []RemoteFileInfo{
		{Name: "vault.enc", Size: 16, ModTime: pushTime},
	})

	// SmartPush
	if _, err := svc.SmartPush(vaultPath); err != nil {
		t.Fatalf("SmartPush failed: %v", err)
	}

	// Verify state was updated
	state, err := LoadState(vaultDir)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}
	if state.LastPushHash == "" {
		t.Error("expected LastPushHash to be set after push")
	}
	if !state.RemoteModTime.Equal(pushTime) {
		t.Errorf("RemoteModTime = %v, want %v", state.RemoteModTime, pushTime)
	}

	// Clear log and simulate remote change
	clearRcloneLog(t, logFile)
	newRemoteTime := time.Date(2026, 1, 29, 12, 0, 0, 0, time.UTC)
	setLsjsonResponse(t, responseFile, []RemoteFileInfo{
		{Name: "vault.enc", Size: 200, ModTime: newRemoteTime},
	})

	// SmartPull should detect remote change and sync
	if err := svc.SmartPull(vaultPath); err != nil {
		t.Fatalf("SmartPull failed: %v", err)
	}

	// Verify rclone sync was called for pull
	lines := readRcloneLog(t, logFile)
	foundSync := false
	for _, line := range lines {
		if strings.Contains(line, "sync") && strings.Contains(line, "mock-remote:bucket") {
			foundSync = true
			break
		}
	}
	if !foundSync {
		t.Errorf("expected rclone sync call during pull, log: %v", lines)
	}
}

func TestIntegration_SmartPull_NoChanges(t *testing.T) {
	binDir, logFile, responseFile := setupMockRclone(t)
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+":"+origPath)

	vaultDir := t.TempDir()
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	if err := os.WriteFile(vaultPath, []byte("vault-content"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.SyncConfig{Enabled: true, Remote: "mock-remote:bucket"}
	svc := NewService(cfg)

	// Push first to establish state
	remoteTime := time.Date(2026, 1, 29, 10, 0, 0, 0, time.UTC)
	setLsjsonResponse(t, responseFile, []RemoteFileInfo{
		{Name: "vault.enc", Size: 13, ModTime: remoteTime},
	})

	if _, err := svc.SmartPush(vaultPath); err != nil {
		t.Fatalf("SmartPush failed: %v", err)
	}

	// Clear log, keep same remote metadata
	clearRcloneLog(t, logFile)

	// SmartPull with same metadata should skip sync
	if err := svc.SmartPull(vaultPath); err != nil {
		t.Fatalf("SmartPull failed: %v", err)
	}

	// Check log: should have lsjson call but NO sync call
	lines := readRcloneLog(t, logFile)
	for _, line := range lines {
		if strings.HasPrefix(line, "sync ") {
			t.Errorf("expected no rclone sync call, but found: %s", line)
		}
	}
}

func TestIntegration_ConflictDetection(t *testing.T) {
	binDir, _, responseFile := setupMockRclone(t)
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+":"+origPath)

	vaultDir := t.TempDir()
	vaultPath := filepath.Join(vaultDir, "vault.enc")
	if err := os.WriteFile(vaultPath, []byte("original-content"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.SyncConfig{Enabled: true, Remote: "mock-remote:bucket"}
	svc := NewService(cfg)

	// Push to establish baseline
	pushTime := time.Date(2026, 1, 29, 10, 0, 0, 0, time.UTC)
	setLsjsonResponse(t, responseFile, []RemoteFileInfo{
		{Name: "vault.enc", Size: 16, ModTime: pushTime},
	})

	if _, err := svc.SmartPush(vaultPath); err != nil {
		t.Fatalf("SmartPush failed: %v", err)
	}

	// Simulate local change (modify vault file)
	if err := os.WriteFile(vaultPath, []byte("locally-modified-content"), 0600); err != nil {
		t.Fatal(err)
	}

	// Simulate remote change (different metadata)
	newRemoteTime := time.Date(2026, 1, 29, 14, 0, 0, 0, time.UTC)
	setLsjsonResponse(t, responseFile, []RemoteFileInfo{
		{Name: "vault.enc", Size: 300, ModTime: newRemoteTime},
	})

	// SmartPull should detect conflict
	err := svc.SmartPull(vaultPath)
	if !errors.Is(err, ErrSyncConflict) {
		t.Errorf("expected ErrSyncConflict, got: %v", err)
	}
}
