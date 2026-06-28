package server

import "log"

func broadcast(message string, sender *Client) {
	mu.Lock()
	defer mu.Unlock()
	for _, client := range clients {
		if client == sender {
			continue
		}
		_, err := client.writer.WriteString(message + "\n")
		if err != nil {
			log.Printf("broadcast write error to %q (%s): %v", client.name, client.conn.RemoteAddr(), err)
			continue
		}
		if err := client.writer.Flush(); err != nil {
			log.Printf("broadcast flush error to %q (%s): %v", client.name, client.conn.RemoteAddr(), err)
		}
	}
}
