package main

import (
	"fmt"
	"net-cat/server"
	"os"
)

// resolvePort picks the listening port from the command-line arguments
// (args is os.Args, so args[0] is the program name). No port argument uses the
// default 8989; exactly one uses that port; more than one is a usage error.
func resolvePort(args []string) (string, bool) {
	switch len(args) {
	case 1:
		return "8989", true
	case 2:
		return args[1], true
	default:
		return "", false
	}
}

func main() {
	port, ok := resolvePort(os.Args)
	if !ok {
		fmt.Println("[USAGE]: ./TCPChat $port")
		os.Exit(1)
	}

	server.Start(port)
}
