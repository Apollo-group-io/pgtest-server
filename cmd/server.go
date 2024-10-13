package main

import (
	"fmt"
	"io"
	"net"
	"path/filepath"
	"reflect"

	"github.com/rubenv/pgtest"
)

// TODO: It would be better if the pgtest package exported the functionality
// to get the socket path, so we didn't have to use reflection.
// maybe contribute to the pgtest package?
func getSocketPathFromPG(p *pgtest.PG) (string, error) {
	// Use reflection to access the private field
	val := reflect.ValueOf(*p)

	// Access the private field 'dir' (first field in the struct)
	dirField := val.FieldByName("dir")

	// Check if the field is valid and can be accessed
	if dirField.IsValid() {
		return filepath.Join(dirField.String(), "sock/.s.PGSQL.5432"), nil
	} else {
		return "", fmt.Errorf("field 'dir' d")
	}
}

func handleConnection(incomingClientSocket net.Conn) {
	defer incomingClientSocket.Close()
	// create a new pgtest database
	db, err := pgtest.Start()
	if err != nil {
		fmt.Printf("error creating new pgtest database: %s\n", err)
		return
	}
	defer db.Stop()
	// run a query to block until the database is ready
	// instead of sleeping for unknown time.
	db.DB.Query("SELECT 1")
	fmt.Printf("created new pgtest database\n")
	// the db creates a temporary directory in which two folders exist 'data' and 'sock'.
	// the one of interest is 'sock' which contains a unix socket file '.s.PGSQL.5432'.
	// get a connection to the database using the socket file
	socketPath, err := getSocketPathFromPG(db)
	if err != nil {
		fmt.Printf("error getting socket path from pgtest database: %s\n", err)
		return
	}
	unixSocketConnectionToDatabase, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Printf("error connecting to database via unix socket %s: %s\n", socketPath, err)
		return
	}
	defer unixSocketConnectionToDatabase.Close()
	fmt.Printf("connected to database via unix socket %s\n", socketPath)

	// pipe the connection between the client and the database
	go io.Copy(unixSocketConnectionToDatabase, incomingClientSocket)
	io.Copy(incomingClientSocket, unixSocketConnectionToDatabase)
	fmt.Println("db or client disconnected..")
	fmt.Printf("closing database connection to socket file %s\n", socketPath)
}

/*
Starts a TCP server, and loops on accepting connections.
For each connection, it spawns a new goroutine to handle the connection.
*/
func startTCPServer() {
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
		go handleConnection(conn)
	}
}

func main() {
	go startTCPServer()
	// keep the main thread alive
	select {}
}
