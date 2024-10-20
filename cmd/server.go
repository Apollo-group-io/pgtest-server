package main

import (
	"fmt"
	"net"
	"pgtestserver"
)

const (
	_TEMPDB_PORT       = 5432
	_DATE_UPDATER_PORT = 5433
)

func main() {
	pgtestserver.GetOrStartBaseDb()
	go startServer(_TEMPDB_PORT, pgtestserver.HandleTempDBConnection)
	go startServer(_DATE_UPDATER_PORT, pgtestserver.HandleDataUpdaterConnection)

	// Keep the main goroutine running
	select {}
}

func startServer(port int, handler func(net.Conn)) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Error starting server on port %d: %v\n", port, err)
		return
	}

	fmt.Printf("Server listening on port %d\n", port)

	for {
		conn, err := listener.Accept()
		fmt.Printf("accepted tcp connection on port %d\n", port)
		if err != nil {
			fmt.Printf("Error accepting connection on port %d: %v\n", port, err)
			continue
		}

		go handler(conn)
	}
}
