package storage

import (
	"errors"
	"fmt"
)

// Atomic save error types (FR-011)
var (
	// ErrVerificationFailed indicates temp file failed decryption verification
	ErrVerificationFailed = errors.New("verification failed")

	// ErrDiskSpaceExhausted indicates insufficient disk space for temp file
	ErrDiskSpaceExhausted = errors.New("insufficient disk space")

	// ErrPermissionDenied indicates cannot write to vault directory
	ErrPermissionDenied = errors.New("permission denied")

	// ErrFilesystemNotAtomic indicates rename operation not supported on filesystem
	ErrFilesystemNotAtomic = errors.New("filesystem does not support atomic operations")
)

// actionableErrorMessage wraps an error with FR-011 compliant message
// Format: "save failed: {reason}. Your vault was not modified. {action}."
func actionableErrorMessage(err error) error {
	if err == nil {
		return nil
	}

	// Check for custom error types and provide actionable guidance
	switch {
	case errors.Is(err, ErrVerificationFailed):
		return fmt.Errorf("save failed during verification. Your vault was not modified. Check your master password and try again: %w", err)

	case errors.Is(err, ErrDiskSpaceExhausted):
		return fmt.Errorf("save failed: insufficient disk space. Your vault was not modified. Free up at least 50 MB and try again: %w", err)

	case errors.Is(err, ErrPermissionDenied):
		return fmt.Errorf("save failed: permission denied. Your vault was not modified. Check file permissions for your vault directory and try again: %w", err)

	case errors.Is(err, ErrFilesystemNotAtomic):
		return fmt.Errorf("save failed: filesystem does not support atomic operations. Your vault was not modified. Move your vault to a local filesystem (not NFS/SMB): %w", err)

	default:
		// Generic fallback for unexpected errors
		return fmt.Errorf("save failed. Your vault was not modified: %w", err)
	}
}

// criticalErrorMessage creates FR-011 compliant message for critical failures
// during the final vault commit step (temp → vault rename)
func criticalErrorMessage(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ErrPermissionDenied):
		return fmt.Errorf("CRITICAL: save failed during final commit (permission denied). Automatic restore attempted. Check file permissions and verify vault integrity: %w", err)

	case errors.Is(err, ErrFilesystemNotAtomic):
		return fmt.Errorf("CRITICAL: save failed during final commit (filesystem error). Automatic restore attempted. If vault is corrupted, restore from backup file manually. verify vault integrity: %w", err)

	default:
		return fmt.Errorf("CRITICAL: save failed during final commit. Automatic restore attempted. verify vault integrity: %w", err)
	}
}
