package pgtestserver

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/Apollo-group-io/pgtest"
)

func getSocketPathFromDir(dir string) string {
	// the db creates a temporary directory in which two folders exist 'data' and 'sock'.
	// the one of interest is 'sock' which contains a unix socket file '.s.PGSQL.5432'.
	return filepath.Join(dir, "sock", ".s.PGSQL.5432")
}

func startTestRunnerDatabaseInTemporaryDirectory() (*pgtest.PG, string, error) {
	// create a new temporary directory for the pgtest database
	temporaryDir, err := os.MkdirTemp("", "pgtest")
	if err != nil {
		return nil, "", fmt.Errorf("error creating temporary directory: %w", err)
	}

	err = copyTemplateDatabase(temporaryDir)
	if err != nil {
		return nil, "", err
	}

	// start the database in the temporary directory
	db, err := pgtest.New().DataDir(temporaryDir).Start()
	if err != nil {
		fmt.Printf("error creating new pgtest database: %s\n", err)
		return nil, "", err
	}
	// run a query to block until the database is ready
	// instead of sleeping for unknown time.
	db.DB.Query("SELECT 1")
	return db, temporaryDir, nil
}

func startSnapshotUpdaterDatabaseInTemporaryDirectory() (*pgtest.PG, string, error) {
	// create a new temporary directory for the pgtest database
	temporaryDir, err := os.MkdirTemp("", "pgtest-updates")
	if err != nil {
		return nil, "", fmt.Errorf("error creating temporary directory: %w", err)
	}

	// start the database in the temporary directory
	db, err := pgtest.New().DataDir(temporaryDir).Persistent().EnableFSync().Start()
	if err != nil {
		fmt.Printf("error creating new pgtest-updates database: %s\n", err)
		return nil, "", err
	}
	// run a query to block until the database is ready
	// instead of sleeping for unknown time.
	db.DB.Query("SELECT 1")
	return db, temporaryDir, nil
}

func copyTemplateDatabase(temporaryDir string) error {
	// FOR TESTING PURPOSES - Create an empty /tmp/templatedb/data directory
	err := os.MkdirAll("/tmp/templatedb/data", 0755)
	if err != nil {
		return fmt.Errorf("error creating template database directory: %w", err)
	}

	// copy the /tmp/templatedb/data folder into the newly created temporaryDir
	err = filepath.Walk("/tmp/templatedb/data", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel("/tmp/templatedb/data", path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(temporaryDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, info.Mode())
	})
	if err != nil {
		return fmt.Errorf("error copying template database: %w", err)
	}

	return nil
}

func getUnixSocketConnectionToDatabase(temporaryDir string) (net.Conn, error) {
	socketPath := getSocketPathFromDir(temporaryDir)
	unixSocketConnectionToDatabase, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	fmt.Printf("connected to database via unix socket %s\n", socketPath)
	return unixSocketConnectionToDatabase, nil
}

// doesn't return until one of the connections is closed
func pipeClientSocketToDb(clientSocket net.Conn, databaseSocket net.Conn) {
	go io.Copy(databaseSocket, clientSocket) // clientSocket -> databaseSocket
	io.Copy(clientSocket, databaseSocket)    // databaseSocket -> clientSocket
}

func HandleClientConnectionTestRunner(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()

	// start the database in a temporary directory
	db, temporaryDir, err := startTestRunnerDatabaseInTemporaryDirectory()
	if err != nil {
		fmt.Printf("error starting database: %s\n", err)
		return
	}
	defer db.Stop()
	fmt.Printf("created new pgtest database\n")

	// get a connection to the database via the unix socket
	unixSocketConnectionToDatabase, err := getUnixSocketConnectionToDatabase(temporaryDir)
	if err != nil {
		fmt.Printf("error getting unix socket connection to database: %s\n", err)
		return
	}
	defer unixSocketConnectionToDatabase.Close()

	// pipe the connection between the client and the database
	pipeClientSocketToDb(incomingClientSocket, unixSocketConnectionToDatabase)
	fmt.Println("db or client disconnected.. closing database connection for temporary directory: ", temporaryDir)
}

func HandleClientConnectionSnapshotUpdater(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()

	// start the database in a temporary directory
	db, temporaryDir, err := startSnapshotUpdaterDatabaseInTemporaryDirectory() // enable fsync
	if err != nil {
		fmt.Printf("error starting database: %s\n", err)
		return
	}
	fmt.Printf("created new pgtest-updates database\n")

	// get a connection to the database via the unix socket
	unixSocketConnectionToDatabase, err := getUnixSocketConnectionToDatabase(temporaryDir)
	if err != nil {
		fmt.Printf("error getting unix socket connection to database: %s\n", err)
		return
	}

	// pipe the connection between the client and the database
	pipeClientSocketToDb(incomingClientSocket, unixSocketConnectionToDatabase)
	fmt.Println("db or client disconnected.. closing database connection for temporary directory: ", temporaryDir)

	unixSocketConnectionToDatabase.Close()
	db.Stop()

	// copy data from temporaryDir into /tmp/templatedb/data
	err = copyUpdatedDatabase(temporaryDir, "/tmp/templatedb")
	if err != nil {
		fmt.Printf("error copying updated database: %s\n", err)
	} else {
		fmt.Println("Successfully updated template database")
	}
}

// Add this new function at the end of the file
func copyUpdatedDatabase(sourceDir, templateDir string) error {
	// Paths for the template database directories
	oldDataDir := filepath.Join(templateDir, "data")
	newDataDir := filepath.Join(templateDir, "new-data")
	tempOldDataDir := filepath.Join(templateDir, "old-data")

	// Step 1: Copy all files from sourceDir to new-data
	err := copyAllFiles(sourceDir, newDataDir)
	if err != nil {
		return fmt.Errorf("error copying updated database: %w", err)
	}

	// Step 2: Rename data to old-data
	err = os.Rename(oldDataDir, tempOldDataDir)
	if err != nil {
		return fmt.Errorf("error renaming old data directory: %w", err)
	}

	// Step 3: Rename new-data to data
	err = os.Rename(newDataDir, oldDataDir)
	if err != nil {
		// If this fails, try to restore the old data
		os.Rename(tempOldDataDir, oldDataDir)
		return fmt.Errorf("error renaming new data directory: %w", err)
	}

	// Step 4: Delete old-data
	err = os.RemoveAll(tempOldDataDir)
	if err != nil {
		fmt.Printf("warning: failed to remove old data directory: %v\n", err)
		// Continue execution even if this fails
	}

	return nil
}

// Helper function to copy all files from src to dst
func copyAllFiles(src, dst string) error {
	// Ensure the destination directory exists
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	// Read all entries in the source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("error reading source directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			srcPath := filepath.Join(src, entry.Name())
			dstPath := filepath.Join(dst, entry.Name())

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return fmt.Errorf("error copying file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// Helper function to copy a single file with metadata
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("error getting source file info: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("error copying file contents: %w", err)
	}

	err = dstFile.Sync()
	if err != nil {
		return fmt.Errorf("error syncing destination file: %w", err)
	}

	return nil
}
