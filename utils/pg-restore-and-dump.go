package utils

import (
	"fmt"
	"os/exec"
	"sync"
)

var dumperMutex sync.Mutex

// CreatePgDump creates a backup of the PostgreSQL database using pg_dump.
func CreatePgDump(sockFolderPath, dumpFilePath, username, dbName string) error {
	dumperMutex.Lock()
	defer dumperMutex.Unlock()

	fmt.Println("start basedb backup to:", dumpFilePath)
	cmd := exec.Command("pg_dump", "-U", username, "-h", sockFolderPath, "-F", "c", "-b", "-v", "-f", dumpFilePath, dbName)
	cmd.Env = append(cmd.Env, "PGPASSWORD=passwordnotimportant")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating dump: %v, output: %s", err, output)
	}

	fmt.Println("done basedb backup to:", dumpFilePath)
	return nil

}

// RestoreDump restores a PostgreSQL database using pg_restore.
func RestorePgDump(sockFolderPath, dumpFilePath, username, dbName string) error {
	dumperMutex.Lock()
	defer dumperMutex.Unlock()

	fmt.Println("start basedb restore from: ", dumpFilePath)

	// Check if the dump file exists
	if !FileExists(dumpFilePath) {
		return fmt.Errorf("dump file does not exist at path: %s", dumpFilePath)
	}

	cmd := exec.Command("pg_restore", "-U", username, "-h", sockFolderPath, "-d", dbName, "-v", dumpFilePath)
	cmd.Env = append(cmd.Env, "PGPASSWORD=passwordnotimportant")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error restoring dump: %v, output: %s", err, output)
	}

	fmt.Println("done basedb restore from: ", dumpFilePath)

	return nil
}
