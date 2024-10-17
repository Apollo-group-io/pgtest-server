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

func handleClientConnectionTestRunner(incomingClientSocket net.Conn) {
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
	fmt.Println("db or client disconnected.. closing database connection for temporary directory: ", temporaryDir)
}

func handleClientConnectionSnapshotUpdater(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()

	// start the database in a temporary directory
	db, temporaryDir, err := startDatabaseInTemporaryDirectory() // enable fsync
	if err != nil {
		fmt.Printf("error starting database: %s\n", err)
		return
	}
	fmt.Printf("created new pgtest database\n")

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

}

/*
Starts a TCP server, and loops on accepting connections.
For each connection, it spawns a new goroutine to handle the connection.
*/
func StartTCPServerForTestRunners() {
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
		go handleClientConnectionTestRunner(conn)
	}
}

func StartTCPServerForSnapshotUpdater() {
	listener, err := net.Listen("tcp", ":5433")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		return
	}

	fmt.Printf("listening on port 5433\n")
	for {
		conn, err := listener.Accept()
		fmt.Printf("accepted tcp connection\n")
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleClientConnectionSnapshotUpdater(conn)
	}
}
