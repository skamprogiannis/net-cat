package main

import (
	"fmt"
	"net-cat/server"
	"os"
)

func main() {
	port := "8989"

	switch len(os.Args) {
	case 1:
		// no port specified, keep default 8989
	case 2:
		port = os.Args[1]
	default:
		fmt.Println("[USAGE]: ./TCPChat $port")
		os.Exit(1)
	}

	server.Start(port)
}
