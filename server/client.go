package server

import (
	"bufio"
	"net"
)

// Client represents one connected chat participant.
type Client struct {
	conn   net.Conn
	name   string
	writer *bufio.Writer
}

func addClient(client *Client) {
	mu.Lock()
	clients = append(clients, client)
	mu.Unlock()
}

func removeClient(client *Client) {
	mu.Lock()
	for i, existingClient := range clients {
		if existingClient == client {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	mu.Unlock()
}
