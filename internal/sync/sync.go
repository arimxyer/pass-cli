// Package sync provides rclone-based vault synchronization for cross-device access.
package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"pass-cli/internal/config"
)

// Service provides vault synchronization using rclone.
type Service struct {
	config config.SyncConfig
}

// NewService creates a new sync service with the given configuration.
func NewService(cfg config.SyncConfig) *Service {
	return &Service{
		config: cfg,
	}
}

// IsEnabled returns true if sync is enabled in the configuration.
func (s *Service) IsEnabled() bool {
	return s.config.Enabled && s.config.Remote != ""
}

// IsRcloneInstalled checks if rclone is available in PATH.
func (s *Service) IsRcloneInstalled() bool {
	_, err := exec.LookPath("rclone")
	return err == nil
}

// Pull syncs the vault from the remote to the local directory.
// This should be called before unlocking the vault to ensure we have the latest version.
func (s *Service) Pull(vaultDir string) error {
	if !s.IsEnabled() {
		return nil
	}

	if !s.IsRcloneInstalled() {
		fmt.Fprintf(os.Stderr, "Warning: sync enabled but rclone not found in PATH\n")
		return nil
	}

	// Ensure local directory exists
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		return fmt.Errorf("failed to create vault directory: %w", err)
	}

	// Run: rclone sync <remote> <local>
	// #nosec G204 -- remote is user-configured in config file
	cmd := exec.Command("rclone", "sync", s.config.Remote, vaultDir)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: sync pull failed: %v\n", err)
		// Don't return error - allow offline operation
		return nil
	}

	return nil
}

// Push syncs the vault from the local directory to the remote.
// This should be called after any write operation (add, update, delete).
func (s *Service) Push(vaultDir string) error {
	if !s.IsEnabled() {
		return nil
	}

	if !s.IsRcloneInstalled() {
		fmt.Fprintf(os.Stderr, "Warning: sync enabled but rclone not found in PATH\n")
		return nil
	}

	// Verify vault directory exists
	if _, err := os.Stat(vaultDir); os.IsNotExist(err) {
		return fmt.Errorf("vault directory does not exist: %s", vaultDir)
	}

	// Run: rclone sync <local> <remote>
	// #nosec G204 -- remote is user-configured in config file
	cmd := exec.Command("rclone", "sync", vaultDir, s.config.Remote)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: sync push failed: %v\n", err)
		// Don't return error - local operation succeeded
		return nil
	}

	return nil
}

// GetVaultDir returns the directory containing the vault file.
func GetVaultDir(vaultPath string) string {
	return filepath.Dir(vaultPath)
}
