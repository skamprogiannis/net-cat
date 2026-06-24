package server

import (
	"bufio"
	"log"
	"net"
)

func notifyJoin(c *Client) {
	broadcast(c.name+" has joined our chat...", c)
}

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

	if err := sendHistory(client); err != nil {
		log.Println("History write error:", err)
		return
	}

	if err := handleClientMessages(client); err != nil {
		log.Println("Client read error:", err)
	}
}
