package sync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/donghun/wm/internal/config"
)

// SyncFile syncs a single file from srcDir to dstDir based on SyncItem config
func SyncFile(srcDir, dstDir string, item config.SyncItem) error {
	srcPath := filepath.Join(srcDir, item.Src)
	dstPath := filepath.Join(dstDir, item.Dst)

	// Check if source exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return nil // Skip if source doesn't exist
	}

	// Check "when" condition
	if item.When == "missing" {
		if _, err := os.Stat(dstPath); err == nil {
			return nil // Skip if dest already exists
		}
	}

	// Ensure destination directory exists
	dstParent := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstParent, 0755); err != nil {
		return fmt.Errorf("failed to create dest directory: %w", err)
	}

	// Remove existing destination if it exists
	os.Remove(dstPath)

	switch item.Mode {
	case "symlink":
		return createSymlink(srcPath, dstPath)
	default: // "copy"
		return copyFile(srcPath, dstPath)
	}
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create dest: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return nil
}

func createSymlink(src, dst string) error {
	// Use absolute path for symlink target
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := os.Symlink(absSrc, dst); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// SyncAll syncs all files from config
func SyncAll(srcDir, dstDir string, items []config.SyncItem) error {
	for _, item := range items {
		// Handle glob patterns
		matches, err := filepath.Glob(filepath.Join(srcDir, item.Src))
		if err != nil {
			return fmt.Errorf("invalid glob pattern %s: %w", item.Src, err)
		}

		if len(matches) == 0 {
			// No glob match, try as literal path
			if err := SyncFile(srcDir, dstDir, item); err != nil {
				return err
			}
			continue
		}

		// Process each glob match
		for _, match := range matches {
			relPath, _ := filepath.Rel(srcDir, match)
			itemCopy := item
			itemCopy.Src = relPath
			if item.Dst == item.Src || item.Dst == "" {
				itemCopy.Dst = relPath
			}
			if err := SyncFile(srcDir, dstDir, itemCopy); err != nil {
				return err
			}
		}
	}

	return nil
}
