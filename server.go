package pgtestserver

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/rubenv/pgtest"
)

func getSocketPathFromDir(dir string) string {
	// the db creates a temporary directory in which two folders exist 'data' and 'sock'.
	// the one of interest is 'sock' which contains a unix socket file '.s.PGSQL.5432'.
	return filepath.Join(dir, "sock", ".s.PGSQL.5432")
}

func startDatabaseInTemporaryDirectory() (*pgtest.PG, string, error) {
	// create a new temporary directory for the pgtest database
	temporaryDir, err := os.MkdirTemp("", "pgtest")
	if err != nil {
		fmt.Printf("error creating temporary directory: %s\n", err)
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

func handleClientConnection(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()

	// start the database in a temporary directory
	db, temporaryDir, err := startDatabaseInTemporaryDirectory()
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
	fmt.Println("db or client disconnected.. closing database connection")
}

/*
Starts a TCP server, and loops on accepting connections.
For each connection, it spawns a new goroutine to handle the connection.
*/
func StartTCPServer() {
	listener, err := net.Listen("tcp", ":5432")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		return
	}

	fmt.Printf("listening on port 5432\n")
	for {
		conn, err := listener.Accept()
		fmt.Printf("accepted tcp connection\n")
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleClientConnection(conn)
	}
}
