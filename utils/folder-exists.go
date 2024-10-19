package utils

import (
	"io"
	"os"
)

func FolderExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		// For other errors, we'll assume the folder doesn't exist
		return false
	}
	if !info.IsDir() {
		return false
	}

	return true
}

func folderNotEmpty(path string) bool {
	// Check if the folder is not empty
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Try to read one entry
	return err != io.EOF       // io.EOF means the directory is empty
}

func FolderExistsAndNotEmpty(path string) bool {
	return FolderExists(path) && folderNotEmpty(path)
}
