package server

import (
	"errors"
	"log"
	"net"
)

func Start(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	if err := serve(listener); err != nil {
		log.Fatal(err)
	}
}

func serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Println("Connection error:", err)
			continue
		}

		mu.Lock()
		if connectionCount >= 10 {
			mu.Unlock()
			conn.Write([]byte("Server is full. Maximum 10 connections reached.\n"))
			conn.Close()
			continue
		}
		connectionCount++
		mu.Unlock()

		go handleClient(conn)
	}
}
