package main

import "pgtestserver"

func main() {
	go pgtestserver.StartTCPServer()
	// keep the main thread alive
	select {}
}
