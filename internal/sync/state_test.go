package sync

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadState_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	state, err := LoadState(tmpDir)
	if err != nil {
		t.Fatalf("LoadState with no file returned error: %v", err)
	}
	if state.LastPushHash != "" {
		t.Errorf("expected empty LastPushHash, got %q", state.LastPushHash)
	}
}

func TestSaveAndLoadState(t *testing.T) {
	tmpDir := t.TempDir()
	now := time.Now().Truncate(time.Second)

	original := &SyncState{
		LastPushHash:  "abc123",
		LastPushTime:  now,
		RemoteModTime: now.Add(-time.Hour),
		RemoteSize:    12345,
	}

	if err := SaveState(tmpDir, original); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	loaded, err := LoadState(tmpDir)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded.LastPushHash != original.LastPushHash {
		t.Errorf("LastPushHash = %q, want %q", loaded.LastPushHash, original.LastPushHash)
	}
	if loaded.RemoteSize != original.RemoteSize {
		t.Errorf("RemoteSize = %d, want %d", loaded.RemoteSize, original.RemoteSize)
	}
}

func TestLoadState_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, syncStateFile)
	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadState(tmpDir)
	if err == nil {
		t.Error("expected error for corrupted state file")
	}
}

func TestHashFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.bin")
	if err := os.WriteFile(path, []byte("hello world"), 0600); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}

	// SHA-256 of "hello world"
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("hash = %q, want %q", hash, expected)
	}
}

func TestHashFile_NotExist(t *testing.T) {
	_, err := HashFile("/nonexistent/file")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestHashFile_DifferentContent(t *testing.T) {
	tmpDir := t.TempDir()

	path1 := filepath.Join(tmpDir, "a.bin")
	path2 := filepath.Join(tmpDir, "b.bin")
	_ = os.WriteFile(path1, []byte("content-a"), 0600)
	_ = os.WriteFile(path2, []byte("content-b"), 0600)

	hash1, _ := HashFile(path1)
	hash2, _ := HashFile(path2)

	if hash1 == hash2 {
		t.Error("different files should have different hashes")
	}
}

func TestStatePath(t *testing.T) {
	got := StatePath("/home/user/.pass-cli")
	expected := filepath.Join("/home/user/.pass-cli", ".sync-state")
	if got != expected {
		t.Errorf("StatePath = %q, want %q", got, expected)
	}
}
