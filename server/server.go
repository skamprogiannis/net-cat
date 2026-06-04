package server

import (
	"log"
	"net"
	"sync"
)

var (
	connectionCount int
	mu              sync.Mutex
)

func Start(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
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

		_ = conn
	}
}
