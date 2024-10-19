package pgtestserver

import (
	"fmt"
	"net"
	"os"
	"path"
	"pgtestserver/utils"

	"github.com/Apollo-group-io/pgtest"
)

func startDataUpdaterDb() (*pgtest.PG, string, error) {
	// create a new temporary directory for the pgtest database
	temporaryDir, err := os.MkdirTemp("", _SNAPSHOT_DB_ROOT_DIR_PREFIX)
	if err != nil {
		return nil, "", fmt.Errorf("error creating temporary directory: %w", err)
	}

	// start the database in the temporary directory
	db, err := utils.StartPgTempDb(temporaryDir, true)
	if err != nil {
		fmt.Printf("error creating new pgtest-updates database: %s\n", err)
		return nil, "", err
	}
	// run a query to block until the database is ready
	// instead of sleeping for unknown time.
	db.DB.Query("SELECT 1")
	return db, temporaryDir, nil
}

func HandleDataUpdaterConnection(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()

	// start the database in a temporary directory
	db, temporaryDir, err := startDataUpdaterDb() // enable fsync
	if err != nil {
		fmt.Printf("error starting database: %s\n", err)
		return
	}
	fmt.Printf("created new pgtest-updates database\n")

	// get a connection to the database via the unix socket
	unixSocketConnectionToDatabase, err := utils.GetUnixSocketConnectionToDatabase(temporaryDir)
	if err != nil {
		fmt.Printf("error getting unix socket connection to database: %s\n", err)
		return
	}

	// pipe the connection between the client and the database
	utils.PipeClientSocketToDb(incomingClientSocket, unixSocketConnectionToDatabase)
	fmt.Println("db or client disconnected.. closing database connection for temporary directory: ", temporaryDir)

	unixSocketConnectionToDatabase.Close()
	db.Stop()

	dataFolderPath := path.Join(temporaryDir, _DATA_DIR_NAME)
	// copy data from temporaryDir into /tmp/templatedb/data
	err = updateTemplateDbDataFolder(dataFolderPath)
	if err != nil {
		fmt.Printf("error copying updated database: %s\n", err)
	} else {
		fmt.Println("Successfully updated template database")
	}
}
