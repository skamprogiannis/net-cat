package server

import (
	"errors"
	"fmt"
	"log"
	"net"
)

// Start listens on port and serves incoming chat connections.
func Start(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}
	defer listener.Close()

	fmt.Printf("Listening on the port :%s\n", port)

	if err := serve(listener); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

// serve runs the accept loop for listener. Tests call it with a listener they
// own so they can close the listener and verify the loop exits cleanly.
func serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		mu.Lock()
		if connectionCount >= 10 {
			mu.Unlock()
			log.Printf("rejected connection from %s: server full (max 10)", conn.RemoteAddr())
			conn.Write([]byte("Server is full. Maximum 10 connections reached.\n"))
			conn.Close()
			continue
		}
		connectionCount++
		mu.Unlock()

		log.Printf("accepted connection from %s", conn.RemoteAddr())
		go handleClient(conn)
	}
}
