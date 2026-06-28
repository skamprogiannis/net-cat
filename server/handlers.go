package server

import (
	"bufio"
	"log"
	"net"
)

func notifyJoin(c *Client) {
	broadcast(c.name+" has joined our chat...", c)
}

func notifyLeave(c *Client) {
	broadcast(c.name+" has left our chat...", c)
}

func disconnectClient(c *Client, joined bool) {
	if joined {
		notifyLeave(c)
	}
	removeClient(c)
}

func handleClient(conn net.Conn) {
	var client *Client
	joined := false

	defer func() {
		if client != nil {
			disconnectClient(client, joined)
		}

		// Start() already incremented for every accepted conn, so always
		// decrement here — even if the client never finished joining (client==nil).
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

	notifyJoin(client)
	joined = true

	if err := handleClientMessages(client); err != nil {
		log.Println("Client read error:", err)
	}
}
