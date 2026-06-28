package main

import (
	"fmt"
	"net-cat/server"
	"os"
)

func main() {
	port := "8989"

	if len(os.Args) == 2 {
		port = os.Args[1]
	} else {
		fmt.Println("[USAGE]: ./TCPChat $port")
		os.Exit(1)
	}

	server.Start(port)
}
