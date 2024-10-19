package pgtestserver

import (
	"fmt"
	"os"
	"path/filepath"
	"pgtestserver/utils"
	"sync"
)

var templateDBMutex sync.Mutex

func copyTemplateDbDataFolderTo(destinationDir string) error {
	templateDBMutex.Lock()
	defer templateDBMutex.Unlock()

	templateDbDataFolderPath := filepath.Join(_TEMPLATE_DB_PATH, _DATA_DIR_NAME)
	if !utils.FolderExistsAndNotEmpty(templateDbDataFolderPath) {
		return fmt.Errorf("template db directory does not exist or is empty: %s", templateDbDataFolderPath)
	}

	// copy the /tmp/templatedb/data folder into the destination
	err := utils.CopySrcDirToDstDir(templateDbDataFolderPath, destinationDir)
	if err != nil {
		return fmt.Errorf("error copying template database: %w", err)
	}

	return nil
}

// Add this new function at the end of the file
func updateTemplateDbDataFolder(sourceDir string) error {
	// Acquire the mutex lock
	templateDBMutex.Lock()
	defer templateDBMutex.Unlock()

	if !utils.FolderExistsAndNotEmpty(sourceDir) {
		fmt.Println("skipping update of template db because updater db left an empty folder")
		return nil
	}

	// Paths for the template database directories
	oldDataDir := filepath.Join(_TEMPLATE_DB_PATH, _DATA_DIR_NAME)
	newDataDir := filepath.Join(_TEMPLATE_DB_PATH, "new-data")
	tempOldDataDir := filepath.Join(_TEMPLATE_DB_PATH, "old-data")

	// Step 1: move source dir to new data dir
	// we're not using .Rename here because the destination
	// might be mounted somewhere. This could end up in a
	// invalid cross device link error
	err := utils.CopySrcDirToDstDir(sourceDir, newDataDir)
	if err != nil {
		return fmt.Errorf("error copying updated database: %w", err)
	}

	// Step 2: move data to old-data if `data` exists
	// previously
	if utils.FolderExists(oldDataDir) {
		err = os.Rename(oldDataDir, tempOldDataDir)
		if err != nil {
			return fmt.Errorf("error renaming old data directory: %w", err)
		}
	}

	// Step 3: move new-data to data
	err = os.Rename(newDataDir, oldDataDir)
	if err != nil {
		// If this fails, try to restore the old data
		os.Rename(tempOldDataDir, oldDataDir)
		return fmt.Errorf("error renaming new data directory: %w", err)
	}

	// Step 4: delete old-data
	err = os.RemoveAll(tempOldDataDir)
	if err != nil {
		return fmt.Errorf("warning: failed to remove old data directory: %v", err)
	}

	return nil
}
