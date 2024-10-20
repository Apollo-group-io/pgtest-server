package pgtestserver

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"pgtestserver/utils"
	"sync"
)

var (
	baseDBMutex sync.Mutex
)

func writeBaseDbDataToDumpFile() error {
	baseDbRootDir := _BASE_DB_PATH
	dumpFilePath := _BASE_DB_DUMP_FILE_LOCATION
	// create dump file with postfix ".new"
	socketFolderPath := utils.GetSockFolderPathForDB(baseDbRootDir)
	err := utils.CreatePgDump(socketFolderPath, dumpFilePath+".new", "postgres", "test")
	if err != nil {
		return fmt.Errorf("error while taking basedb dump: %s", err)
	}

	// overwrite the old file with the new file in one atomic file system operation
	os.Rename(dumpFilePath+".new", dumpFilePath)
	return nil
}

func initializeBaseDbDataFromDumpFile(baseDbRootDir, dumpFilePath string) error {
	socketFolderPath := utils.GetSockFolderPathForDB(baseDbRootDir)
	return utils.RestorePgDump(socketFolderPath, dumpFilePath, "postgres", "test")
}

func isBaseDbUp() bool {
	dataFolderExists := utils.FolderExistsAndNotEmpty(filepath.Join(_BASE_DB_PATH, "data"))
	sockFolderExists := utils.FolderExistsAndNotEmpty(filepath.Join(_BASE_DB_PATH, "sock"))
	return dataFolderExists && sockFolderExists
}

func StartBaseDbIfNotUp() error {
	baseDBMutex.Lock()
	defer baseDBMutex.Unlock()

	if isBaseDbUp() {
		return nil
	}

	// start the database in the baseDbDir
	db, err := utils.StartPGTestDB(_BASE_DB_PATH, true)
	if err != nil {
		os.RemoveAll(_BASE_DB_PATH) // remove basedb folder
		return fmt.Errorf("error creating updater db: %s", err)
	}
	// when starting the first time, restore from the .dump file
	if utils.FileExists(_BASE_DB_DUMP_FILE_LOCATION) {
		fmt.Println("base-db: dump file exists, restoring from base file: ", _BASE_DB_DUMP_FILE_LOCATION)
		err = initializeBaseDbDataFromDumpFile(_BASE_DB_PATH, _BASE_DB_DUMP_FILE_LOCATION)
		if err != nil {
			db.Stop() // stop removes basedb folder
			return fmt.Errorf("error restoring basedb from dump: %s", err)
		}
	} else {
		fmt.Println("base-db: dump file does not exist, starting empty...")
	}

	return nil
}

func startDbAndPipeUntilConnectionClosed(incomingClientSocket net.Conn) error {
	// start the database in a temporary directory
	err := StartBaseDbIfNotUp() // enable fsync
	if err != nil {
		return fmt.Errorf("error starting database: %s", err)
	}
	fmt.Printf("created new pgtest-updates database\n")

	// get a connection to the database via the unix socket
	unixSocketConnectionToDatabase, err := utils.GetUnixSocketConnectionToDatabase(_BASE_DB_PATH)
	if err != nil {
		return fmt.Errorf("error getting unix socket connection to database: %s", err)
	}
	defer unixSocketConnectionToDatabase.Close()

	// pipe the connection between the client and the database
	utils.PipeClientSocketToDb(incomingClientSocket, unixSocketConnectionToDatabase)
	fmt.Println("db or client disconnected.. closing database connection for basedb: ", _BASE_DB_PATH)

	return nil
}

func HandleBaseDBConnection(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()

	err := startDbAndPipeUntilConnectionClosed(incomingClientSocket)
	if err != nil {
		fmt.Println("error while processing connection: ", err)
	}

	// take a dump after closing the connection
	err = writeBaseDbDataToDumpFile()
	if err != nil {
		fmt.Println("error while creating a dump from basedb: ", err)
	}

}
