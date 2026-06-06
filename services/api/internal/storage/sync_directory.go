package storage

import (
	"fmt"
	"os"
)

func syncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open directory %q: %w", path, err)
	}

	if err := dir.Sync(); err != nil {
		_ = dir.Close()
		return fmt.Errorf("sync directory %q: %w", path, err)
	}
	if err := dir.Close(); err != nil {
		return fmt.Errorf("close directory %q: %w", path, err)
	}
	return nil
}
