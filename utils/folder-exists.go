package utils

import "os"

func FolderExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		// For other errors, we'll assume the folder doesn't exist
		return false
	}
	return info.IsDir()
}
