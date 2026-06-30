package server

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

func formatMessage(name, message string, t time.Time) string {
	return fmt.Sprintf("[%s][%s]: %s", t.Format("2006-01-02 15:04:05"), name, message)
}

// formatPrompt builds the simple input marker shown after live chat output.
func formatPrompt() string {
	return "> "
}

// sendPrompt writes a fresh input prompt to a single client.
func sendPrompt(client *Client) error {
	mu.Lock()
	defer mu.Unlock()

	if _, err := client.writer.WriteString(formatPrompt()); err != nil {
		return err
	}
	return client.writer.Flush()
}

func sendLiveLine(client *Client, message string, leadingNewline bool) error {
	mu.Lock()
	defer mu.Unlock()

	return sendLiveLineLocked(client, message, leadingNewline)
}

func sendLiveLineLocked(client *Client, message string, leadingNewline bool) error {
	if leadingNewline {
		if _, err := client.writer.WriteString("\n"); err != nil {
			return err
		}
	}
	if _, err := client.writer.WriteString(message + "\n" + formatPrompt()); err != nil {
		return err
	}
	return client.writer.Flush()
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

// handleClientMessages reads raw client input, filters empty lines at the input
// boundary, then records and broadcasts formatted chat messages to every
// connected client, including the sender.
func handleClientMessages(client *Client) error {
	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}

		formatted := formatMessage(client.name, message, time.Now())
		addMessageToHistory(formatted)

		if err := sendLiveLine(client, formatted, false); err != nil {
			return err
		}
		broadcast(formatted, client)
	}
	return scanner.Err()
}
