package utils

import (
	"fmt"
	"io"
	"net"
	"path/filepath"
)

const (
	_PG_UNIX_SOCKET_FILE_NAME = ".s.PGSQL.5432"
	_SOCK_DIR_NAME            = "sock"
)

func GetSockFilePathForDB(dbRootDir string) string {
	// the db creates a temporary directory in which two folders exist 'data' and 'sock'.
	// the one of interest is 'sock' which contains a unix socket file '.s.PGSQL.5432'.
	return filepath.Join(dbRootDir, _SOCK_DIR_NAME, _PG_UNIX_SOCKET_FILE_NAME)
}

func GetUnixSocketConnectionToDatabase(dbRootDir string) (net.Conn, error) {
	socketPath := GetSockFilePathForDB(dbRootDir)

	unixSocketConnectionToDatabase, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	fmt.Printf("connected to database via unix socket %s\n", socketPath)
	return unixSocketConnectionToDatabase, nil
}

// doesn't return until one of the connections is closed
func PipeClientSocketToDb(clientSocket net.Conn, databaseSocket net.Conn) {
	go io.Copy(databaseSocket, clientSocket) // clientSocket -> databaseSocket
	io.Copy(clientSocket, databaseSocket)    // databaseSocket -> clientSocket
}
