package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func handleClient(conn net.Conn, name string) {
	client := &Client{
		conn:   conn,
		name:   name,
		writer: bufio.NewWriter(conn),
	}
	addClient(client)
	defer func() {
		removeClient(client)
		mu.Lock()
		connectionCount--
		mu.Unlock()
		conn.Close()
	}()

}

func readClientName(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	for {
		name, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		name = strings.TrimSpace(name)
		if name != "" {
			return name, nil
		}
		fmt.Fprint(conn, "[ENTER YOUR NAME]: ")
	}
}
