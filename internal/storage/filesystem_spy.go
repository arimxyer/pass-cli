package storage

import (
	"os"
	"path/filepath"
)

// spyFileSystem wraps the real OS filesystem but can simulate failures for testing
// It delegates all operations to the real OS except where configured to fail
type spyFileSystem struct {
	realFS   *osFileSystem
	delegate FileSystem // Alias for realFS to support test compatibility

	// Rename failure configuration
	renameCallCount   int
	failRenameAt      int              // Fail rename on Nth call (0 = don't fail)
	failRenameAtSet   map[int]bool     // Fail rename on multiple specific calls
	failRenameWithErr error            // Custom error to return on rename failure
	failAllRenames    bool             // Fail all rename calls (alias for compatibility)
	failRename        bool             // Alias for failAllRenames

	// ReadFile failure configuration
	readFileCallCount  int
	failReadFileAtCall int // Fail ReadFile on Nth call (0 = don't fail)

	// OpenFile failure configuration
	failOpenFileWithErrFunc func(path string) error // Custom error function for OpenFile
}

// NewSpyFileSystem creates a filesystem that delegates to real OS but can fail on demand
func NewSpyFileSystem() *spyFileSystem {
	fs := &osFileSystem{}
	return &spyFileSystem{
		realFS:          fs,
		delegate:        fs, // Populate delegate alias
		failRenameAtSet: make(map[int]bool),
	}
}

func (s *spyFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	// Check if we should fail with custom error
	if s.failOpenFileWithErrFunc != nil {
		if err := s.failOpenFileWithErrFunc(name); err != nil {
			return nil, err
		}
	}
	return s.realFS.OpenFile(name, flag, perm)
}

func (s *spyFileSystem) ReadFile(name string) ([]byte, error) {
	s.readFileCallCount++

	// Check if we should fail this specific ReadFile call
	if s.failReadFileAtCall > 0 && s.readFileCallCount == s.failReadFileAtCall {
		return nil, os.ErrPermission
	}

	return s.realFS.ReadFile(name)
}

func (s *spyFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return s.realFS.WriteFile(name, data, perm)
}

func (s *spyFileSystem) Remove(name string) error {
	return s.realFS.Remove(name)
}

func (s *spyFileSystem) Rename(oldpath, newpath string) error {
	s.renameCallCount++

	// Determine which error to return (custom or default)
	errToReturn := s.failRenameWithErr
	if errToReturn == nil {
		errToReturn = os.ErrPermission // Default error
	}

	// Check if we should fail this specific rename call (single)
	if s.failRenameAt > 0 && s.renameCallCount == s.failRenameAt {
		return errToReturn
	}

	// Check if we should fail this specific rename call (multiple)
	if s.failRenameAtSet[s.renameCallCount] {
		return errToReturn
	}

	if s.failAllRenames || s.failRename {
		return errToReturn
	}

	// Delegate to real filesystem
	return s.realFS.Rename(oldpath, newpath)
}

func (s *spyFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return s.realFS.MkdirAll(path, perm)
}

func (s *spyFileSystem) Stat(name string) (os.FileInfo, error) {
	return s.realFS.Stat(name)
}

func (s *spyFileSystem) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// ResetRenameCount resets the rename call counter
func (s *spyFileSystem) ResetRenameCount() {
	s.renameCallCount = 0
}

// GetRenameCallCount returns the current rename call count
func (s *spyFileSystem) GetRenameCallCount() int {
	return s.renameCallCount
}
