package server

import (
	"bufio"
	"net"
)

type Client struct {
	conn   net.Conn
	name   string
	writer *bufio.Writer
}

func addClient(c *Client) {
	mu.Lock()
	clients = append(clients, c)
	mu.Unlock()
}

func removeClient(c *Client) {
	mu.Lock()
	for i, client := range clients {
		if client == c {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	mu.Unlock()
}
