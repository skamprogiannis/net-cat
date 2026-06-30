package server

import "log"

func broadcast(message string, sender *Client) {
	mu.Lock()
	defer mu.Unlock()
	for _, client := range clients {
		if client == sender {
			continue
		}
		if err := sendLiveLineLocked(client, message, true); err != nil {
			log.Printf("broadcast write error to %q (%s): %v", client.name, client.conn.RemoteAddr(), err)
			continue
		}
	}
}
