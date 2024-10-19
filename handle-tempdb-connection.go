package pgtestserver

import (
	"fmt"
	"net"
	"os"
	"pgtestserver/utils"

	"github.com/Apollo-group-io/pgtest"
)

func startTempDatabase() (*pgtest.PG, string, error) {
	tempDir, err := os.MkdirTemp("", _TEMP_DB_ROOT_DIR_PREFIX)
	if err != nil {
		return nil, "", fmt.Errorf("error creating temp folder for temp db: %s", err)
	}
	// start the database in the temporary directory
	db, err := utils.StartTempDb(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("error creating new pgtest database: %s", err)
	}
	// use pg_restore to backup from existing template db.
	return db, tempDir, nil
}

func HandleTempDBConnection(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()

	// start the database in a temporary directory
	db, temporaryDir, err := startTempDatabase()
	if err != nil {
		fmt.Println("error starting database: ", err)
		return
	}
	defer db.Stop()
	fmt.Println("created new pgtest database")

	// get a connection to the database via the unix socket
	unixSocketConnectionToDatabase, err := utils.GetUnixSocketConnectionToDatabase(temporaryDir)
	if err != nil {
		fmt.Println("error getting unix socket connection to database: ", err)
		return
	}
	defer unixSocketConnectionToDatabase.Close()

	// pipe the connection between the client and the database
	utils.PipeClientSocketToDb(incomingClientSocket, unixSocketConnectionToDatabase)
	fmt.Println("db or client disconnected.. closing database connection for temporary directory: ", temporaryDir)
}
