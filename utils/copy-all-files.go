package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Helper function to copy all files from src to dst
func CopySrcDirToDstDir(src, dst string) error {
	// Get information about the source directory
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error getting source info: %w", err)
	}

	// Create the destination directory with the same permissions
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	// Walk through the source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through source: %w", err)
		}

		// Calculate the relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("error calculating relative path: %w", err)
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create the directory in the destination
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy the file
		return copyFile(path, dstPath, info.Mode())
	})
}

// Helper function to copy a single file
func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("error copying file contents: %w", err)
	}

	return nil
}
