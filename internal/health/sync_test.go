package health

import (
	"context"
	"testing"

	"pass-cli/internal/config"
)

// ARI-53: Tests for SyncChecker

func TestSyncCheck_Disabled(t *testing.T) {
	checker := NewSyncChecker(config.SyncConfig{
		Enabled: false,
		Remote:  "",
	})

	result := checker.Run(context.Background())

	if result.Name != "sync" {
		t.Errorf("Expected name 'sync', got %q", result.Name)
	}
	if result.Status != CheckPass {
		t.Errorf("Expected status CheckPass when sync disabled, got %s", result.Status)
	}
	if result.Message != "Cloud sync is disabled" {
		t.Errorf("Unexpected message: %s", result.Message)
	}

	details, ok := result.Details.(SyncCheckDetails)
	if !ok {
		t.Fatal("Expected SyncCheckDetails in details")
	}
	if details.Enabled {
		t.Error("Expected Enabled=false in details")
	}
}

func TestSyncCheck_EnabledNoRemote(t *testing.T) {
	checker := NewSyncChecker(config.SyncConfig{
		Enabled: true,
		Remote:  "",
	})

	result := checker.Run(context.Background())

	if result.Status != CheckWarning {
		t.Errorf("Expected status CheckWarning when no remote, got %s", result.Status)
	}
	if result.Recommendation == "" {
		t.Error("Expected recommendation when remote not configured")
	}
}

func TestSyncCheck_EnabledWithRemote(t *testing.T) {
	checker := NewSyncChecker(config.SyncConfig{
		Enabled: true,
		Remote:  "gdrive:.pass-cli",
	})

	result := checker.Run(context.Background())

	details, ok := result.Details.(SyncCheckDetails)
	if !ok {
		t.Fatal("Expected SyncCheckDetails in details")
	}

	if !details.Enabled {
		t.Error("Expected Enabled=true in details")
	}
	if details.Remote != "gdrive:.pass-cli" {
		t.Errorf("Expected remote 'gdrive:.pass-cli', got %q", details.Remote)
	}

	// Result depends on whether rclone is installed
	// If installed: CheckPass with version
	// If not installed: CheckWarning
	if result.Status != CheckPass && result.Status != CheckWarning {
		t.Errorf("Expected CheckPass or CheckWarning, got %s", result.Status)
	}

	if result.Status == CheckWarning {
		// Should recommend installing rclone
		if result.Recommendation == "" {
			t.Error("Expected recommendation when rclone not installed")
		}
	}

	if result.Status == CheckPass {
		// Should have rclone version if installed
		if details.RcloneInstalled && details.RcloneVersion == "" {
			t.Error("Expected rclone version when installed")
		}
	}
}

func TestSyncCheck_Name(t *testing.T) {
	checker := NewSyncChecker(config.SyncConfig{})
	if checker.Name() != "sync" {
		t.Errorf("Expected name 'sync', got %q", checker.Name())
	}
}
