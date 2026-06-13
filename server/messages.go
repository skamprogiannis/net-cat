package server

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

func formatMessage(name, message string, t time.Time) string {
	return fmt.Sprintf("[%s][%s]:%s", t.Format("2006-01-02 15:04:05"), name, message)
}

func addMessageToHistory(message string) {
	mu.Lock()
	messageHistory = append(messageHistory, message)
	mu.Unlock()
}

func sendHistory(client *Client) error {
	mu.Lock()
	defer mu.Unlock()

	for _, message := range messageHistory {
		if _, err := client.writer.WriteString(message + "\n"); err != nil {
			return err
		}
	}
	return client.writer.Flush()
}

func handleClientMessages(client *Client) error {
	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}

		formatted := formatMessage(client.name, message, time.Now())
		addMessageToHistory(formatted)
		broadcast(formatted, client)
	}
	return scanner.Err()
}
