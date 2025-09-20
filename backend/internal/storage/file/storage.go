package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Storage provides a simple file-based storage backend.
// It stores files under a specified base path on the local filesystem.
type Storage struct {
	basePath string
}

// NewStorage creates a new Storage instance with the given basePath.
// The basePath defines the root directory where files will be stored.
func NewStorage(basePath string) *Storage {
	return &Storage{basePath: basePath}
}

// Save stores the uploaded file in the given subdirectory (e.g. "original" or "processed")
// with the provided filename.
func (s *Storage) Save(subdir, filename string, src io.Reader) (string, error) {
	dir := filepath.Join(s.basePath, subdir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	dstPath := filepath.Join(dir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", dstPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file %s: %w", dstPath, err)
	}

	return dstPath, nil
}

// Load opens the file and returns a reader.
func (s *Storage) Load(subdir, filename string) (*os.File, error) {
	path := filepath.Join(s.basePath, subdir, filename)

	return os.Open(path)
}

// Delete removes the file from storage.
func (s *Storage) Delete(subdir, filename string) error {
	path := filepath.Join(s.basePath, subdir, filename)

	return os.Remove(path)
}
