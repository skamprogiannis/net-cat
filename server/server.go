package server

import (
	"log"
	"net"
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

		if err := sendWelcomePrompt(conn); err != nil {
			log.Println("Welcome message error:", err)
			mu.Lock()
			connectionCount--
			mu.Unlock()
			conn.Close()
			continue
		}
		name, err := readClientName(conn)
		if err != nil {
			log.Println("Name read error:", err)
			mu.Lock()
			connectionCount--
			mu.Unlock()
			conn.Close()
			continue
		}
		go handleClient(conn, name)
	}
}
