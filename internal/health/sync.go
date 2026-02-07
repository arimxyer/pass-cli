package health

import (
	"context"
	"os/exec"
	"strings"

	"github.com/arimxyer/pass-cli/internal/config"
)

// SyncChecker verifies cloud sync configuration and rclone availability
// ARI-53: Added for doctor command sync health checks
type SyncChecker struct {
	syncConfig config.SyncConfig
}

// NewSyncChecker creates a new sync health checker
func NewSyncChecker(syncConfig config.SyncConfig) HealthChecker {
	return &SyncChecker{syncConfig: syncConfig}
}

// Name returns the check identifier
func (s *SyncChecker) Name() string {
	return "sync"
}

// Run executes the sync health check
func (s *SyncChecker) Run(ctx context.Context) CheckResult {
	details := SyncCheckDetails{
		Enabled: s.syncConfig.Enabled,
		Remote:  s.syncConfig.Remote,
	}

	// Check if sync is disabled
	if !s.syncConfig.Enabled {
		return CheckResult{
			Name:    s.Name(),
			Status:  CheckPass,
			Message: "Cloud sync is disabled",
			Details: details,
		}
	}

	// Check if remote is configured
	if s.syncConfig.Remote == "" {
		return CheckResult{
			Name:           s.Name(),
			Status:         CheckWarning,
			Message:        "Sync enabled but no remote configured",
			Recommendation: "Add sync.remote to config (e.g., gdrive:.pass-cli)",
			Details:        details,
		}
	}

	// Check rclone installation
	rclonePath, err := exec.LookPath("rclone")
	details.RcloneInstalled = err == nil

	if !details.RcloneInstalled {
		return CheckResult{
			Name:           s.Name(),
			Status:         CheckWarning,
			Message:        "Sync enabled but rclone not found in PATH",
			Recommendation: "Install rclone: brew install rclone (macOS), scoop install rclone (Windows), or visit https://rclone.org/install/",
			Details:        details,
		}
	}

	// Get rclone version
	// #nosec G204 -- rclonePath is from exec.LookPath, not user input
	if out, err := exec.CommandContext(ctx, rclonePath, "version").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 0 {
			// First line is typically "rclone v1.68.2" or similar
			version := strings.TrimSpace(lines[0])
			version = strings.TrimPrefix(version, "rclone ")
			details.RcloneVersion = version
		}
	}

	return CheckResult{
		Name:    s.Name(),
		Status:  CheckPass,
		Message: "Cloud sync configured and ready",
		Details: details,
	}
}
