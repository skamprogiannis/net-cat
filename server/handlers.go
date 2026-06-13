package server

import (
	"bufio"
	"log"
	"net"
	"strings"
)

func handleClient(conn net.Conn) {
	var client *Client

	defer func() {
		if client != nil {
			removeClient(client)
		}
		mu.Lock()
		connectionCount--
		mu.Unlock()
		conn.Close()
	}()

	if err := sendWelcomePrompt(conn); err != nil {
		log.Println("Welcome message error:", err)
		return
	}

	name, err := readClientName(conn)
	if err != nil {
		log.Println("Name read error:", err)
		return
	}

	client = &Client{
		conn:   conn,
		name:   name,
		writer: bufio.NewWriter(conn),
	}
	addClient(client)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}
		broadcast(message, client)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Client read error:", err)
	}
}
