package pgtestserver

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"pgtestserver/utils"

	"github.com/Apollo-group-io/pgtest"
)

func startTempDatabase() (*pgtest.PG, string, error) {
	// create a new temporary directory for the pgtest database
	temporaryDir, err := os.MkdirTemp("", _TEST_DB_ROOT_DIR_PREFIX)
	if err != nil {
		return nil, "", fmt.Errorf("error creating temporary directory: %w", err)
	}

	dataDirectory := filepath.Join(temporaryDir, _DATA_DIR_NAME)
	err = copyTemplateDbDataFolderTo(dataDirectory)
	if err != nil {
		fmt.Printf("error cloning template db: %s\n", err)
		return nil, "", err
	}

	// start the database in the temporary directory
	db, err := utils.StartPgTempDb(temporaryDir, false)
	if err != nil {
		fmt.Printf("error creating new pgtest database: %s\n", err)
		return nil, "", err
	}
	// run a query to block until the database is ready
	// instead of sleeping for unknown time.
	db.DB.Query("SELECT 1")
	return db, temporaryDir, nil
}

func HandleTempDBConnection(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()

	// start the database in a temporary directory
	db, temporaryDir, err := startTempDatabase()
	if err != nil {
		fmt.Printf("error starting database: %s\n", err)
		return
	}
	defer db.Stop()
	fmt.Printf("created new pgtest database\n")

	// get a connection to the database via the unix socket
	unixSocketConnectionToDatabase, err := utils.GetUnixSocketConnectionToDatabase(temporaryDir)
	if err != nil {
		fmt.Printf("error getting unix socket connection to database: %s\n", err)
		return
	}
	defer unixSocketConnectionToDatabase.Close()

	// pipe the connection between the client and the database
	utils.PipeClientSocketToDb(incomingClientSocket, unixSocketConnectionToDatabase)
	fmt.Println("db or client disconnected.. closing database connection for temporary directory: ", temporaryDir)
}
