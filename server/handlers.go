package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

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
