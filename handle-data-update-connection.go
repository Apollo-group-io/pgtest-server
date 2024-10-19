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
		os.RemoveAll(temporaryDir)
		return nil, "", fmt.Errorf("error creating updater db: %w", err)
	}
	// run a query to block until the database is ready
	// instead of sleeping for unknown time.
	db.DB.Query("SELECT 1")
	return db, temporaryDir, nil
}

func startDbAndPipeUntilConnectionClosed(incomingClientSocket net.Conn) (string, error) {
	// start the database in a temporary directory
	db, temporaryDir, err := startDataUpdaterDb() // enable fsync
	if err != nil {
		fmt.Printf("error starting database: %s\n", err)
		return "", err
	}
	defer db.Stop()
	fmt.Printf("created new pgtest-updates database\n")

	// get a connection to the database via the unix socket
	unixSocketConnectionToDatabase, err := utils.GetUnixSocketConnectionToDatabase(temporaryDir)
	if err != nil {
		fmt.Printf("error getting unix socket connection to database: %s\n", err)
		return "", err
	}
	defer unixSocketConnectionToDatabase.Close()

	// pipe the connection between the client and the database
	utils.PipeClientSocketToDb(incomingClientSocket, unixSocketConnectionToDatabase)
	fmt.Println("db or client disconnected.. closing database connection for temporary directory: ", temporaryDir)

	return temporaryDir, nil
}

func HandleDataUpdaterConnection(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()

	temporaryDir, err := startDbAndPipeUntilConnectionClosed(incomingClientSocket)
	if err != nil {
		fmt.Println("error while processing connection: ", err)
	}
	defer (func() {
		// Step 4: delete old-data
		fmt.Println("removing data directory left behind by updater")
		err = os.RemoveAll(temporaryDir)
		if err != nil {
			fmt.Println("warning: failed to remove temp directory: ", temporaryDir, "->", err)
		}
	})()

	dataFolderPath := path.Join(temporaryDir, _DATA_DIR_NAME)
	// copy data from temporaryDir into /tmp/templatedb/data
	err = updateTemplateDbDataFolder(dataFolderPath)
	if err != nil {
		fmt.Println("error copying updated database: ", err)
		return
	}

	fmt.Println("Successfully updated template database")
}
