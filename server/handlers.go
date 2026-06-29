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

// disconnectClient removes the client and only announces a leave if the join
// flow completed. Connections can fail before joining, and those should not
// produce misleading leave messages.
func disconnectClient(c *Client, joined bool) {
	if joined {
		log.Printf("%q (%s) left the chat", c.name, c.conn.RemoteAddr())
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

	addr := conn.RemoteAddr()

	if err := sendWelcomePrompt(conn); err != nil {
		log.Printf("failed to send welcome prompt to %s: %v", addr, err)
		return
	}

	name, err := readClientName(conn)
	if err != nil {
		log.Printf("failed to read client name from %s: %v", addr, err)
		return
	}

	client = &Client{
		conn:   conn,
		name:   name,
		writer: bufio.NewWriter(conn),
	}
	addClient(client)

	if err := sendHistory(client); err != nil {
		log.Printf("failed to send history to %q (%s): %v", name, addr, err)
		return
	}

	notifyJoin(client)
	joined = true
	log.Printf("%q (%s) joined the chat", name, addr)

	if err := handleClientMessages(client); err != nil {
		log.Printf("read error for %q (%s): %v", name, addr, err)
	}
}
