package utils

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

var dumperMutex sync.Mutex

// CreatePgDump creates a backup of the PostgreSQL database using pg_dump.
func CreatePgDump(socketPath, dumpFilePath, username, dbName string) error {
	dumperMutex.Lock()
	defer dumperMutex.Unlock()
	cmd := exec.Command("pg_dump", "-U", username, "-h", socketPath, "-F", "c", "-b", "-v", "-f", dumpFilePath, dbName)
	cmd.Env = append(cmd.Env, "PGPASSWORD=passwordnotimportant")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating dump: %v, output: %s", err, output)
	}
	return nil
}

// RestoreDump restores a PostgreSQL database using pg_restore.
func RestorePgDump(socketPath, dumpFilePath, username, dbName string) error {
	dumperMutex.Lock()
	defer dumperMutex.Unlock()

	// Check if the dump file exists
	if _, err := os.Stat(dumpFilePath); os.IsNotExist(err) {
		return fmt.Errorf("dump file does not exist at path: %s", dumpFilePath)
	}

	cmd := exec.Command("pg_restore", "-U", username, "-h", socketPath, "-d", dbName, "-v", dumpFilePath)
	cmd.Env = append(cmd.Env, "PGPASSWORD=passwordnotimportant")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error restoring dump: %v, output: %s", err, output)
	}
	return nil
}
