package pgtestserver

import (
	"fmt"
	"net"
	"os"
	"pgtestserver/utils"
	"sync"

	"github.com/Apollo-group-io/pgtest"
)

var (
	templateDb      *pgtest.PG = nil
	templateDBMutex sync.Mutex
)

func createTemplateDbDump() error {
	socketPath := utils.GetSockFolderPathForDB(_TEMPLATE_DB_PATH)
	return utils.CreatePgDump(socketPath, _TEMPLATE_DB_DUMP_FILE_LOCATION, "postgres", "test")
}

func GetOrStartTemplateDb() (string, error) {
	templateDBMutex.Lock()
	defer templateDBMutex.Unlock()
	/*
		We are going to check if we have a db, if not we will start it,
		and from that point on all incoming connections will be sent to this
		database.
	*/
	if templateDb != nil {
		return _TEMPLATE_DB_PATH, nil
	}

	// start the database in the templateDbDir
	db, err := utils.StartTemplateDB(_TEMPLATE_DB_PATH)
	if err != nil {
		os.RemoveAll(_TEMPLATE_DB_PATH)
		return "", fmt.Errorf("error creating updater db: %w", err)
	}
	// run a query to block until the database is ready
	// instead of sleeping for unknown time.
	db.DB.Query("SELECT 1")
	// take the db dump when starting the first time
	err = createTemplateDbDump()
	if err != nil {
		fmt.Println("error while creating a dump from template db: ", err)
	}
	templateDb = db
	return _TEMPLATE_DB_PATH, nil
}

func startDbAndPipeUntilConnectionClosed(incomingClientSocket net.Conn) error {
	// start the database in a temporary directory
	dbRootDir, err := GetOrStartTemplateDb() // enable fsync
	if err != nil {
		return fmt.Errorf("error starting database: %s", err)
	}
	fmt.Printf("created new pgtest-updates database\n")

	// get a connection to the database via the unix socket
	unixSocketConnectionToDatabase, err := utils.GetUnixSocketConnectionToDatabase(dbRootDir)
	if err != nil {
		return fmt.Errorf("error getting unix socket connection to database: %s", err)
	}
	defer unixSocketConnectionToDatabase.Close()

	// pipe the connection between the client and the database
	utils.PipeClientSocketToDb(incomingClientSocket, unixSocketConnectionToDatabase)
	fmt.Println("db or client disconnected.. closing database connection for temporary directory: ", dbRootDir)

	// take a dump after closing the connection
	createTemplateDbDump()

	return nil
}

func HandleDataUpdaterConnection(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()

	err := startDbAndPipeUntilConnectionClosed(incomingClientSocket)
	if err != nil {
		fmt.Println("error while processing connection: ", err)
	}

}
