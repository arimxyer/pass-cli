package storage

import (
	"os"
	"path/filepath"
)

// FileSystem abstracts file system operations for testability
type FileSystem interface {
	// File operations
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	Remove(name string) error
	Rename(oldpath, newpath string) error

	// Directory operations
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)

	// Glob for pattern matching
	Glob(pattern string) ([]string, error)
}

// osFileSystem implements FileSystem using real os package
type osFileSystem struct{}

// NewOSFileSystem creates a FileSystem backed by the real os package
func NewOSFileSystem() FileSystem {
	return &osFileSystem{}
}

func (f *osFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	// #nosec G304 -- filesystem abstraction layer, file paths are validated by callers (vault, config, etc.)
	return os.OpenFile(name, flag, perm)
}

func (f *osFileSystem) ReadFile(name string) ([]byte, error) {
	// #nosec G304 -- filesystem abstraction layer, file paths are validated by callers (vault, config, etc.)
	return os.ReadFile(name)
}

func (f *osFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (f *osFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (f *osFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (f *osFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *osFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f *osFileSystem) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}
