package main

import "pgtestserver"

func main() {
	go pgtestserver.StartTCPServerForTestRunners()
	// keep the main thread alive
	select {}
}
