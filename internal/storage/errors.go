package storage

import "errors"

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
