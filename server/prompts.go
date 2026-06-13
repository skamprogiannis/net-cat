package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

const welcomeMessage = "Welcome to TCP-Chat!\n" +
	"         _nnnn_\n" +
	"        dGGGGMMb\n" +
	"       @p~qp~~qMb\n" +
	"       M|@||@) M|\n" +
	"       @,----.JM|\n" +
	"      JS^\\__/  qKL\n" +
	"     dZP        qKRb\n" +
	"    dZP          qKKb\n" +
	"   fZP            SMMb\n" +
	"   HZM            MMMM\n" +
	"   FqM            MMMM\n" +
	" __| \".        |\\dS\"qML\n" +
	" |    `.       | `' \\Zq\n" +
	"_)      \\.___.,|     .'\n" +
	"\\____   )MMMMMP|   .'\n" +
	"     `-'       `--'\n"

const namePrompt = "[ENTER YOUR NAME]: "

func sendWelcomePrompt(conn net.Conn) error {
	_, err := fmt.Fprint(conn, welcomeMessage, namePrompt)
	return err
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
		fmt.Fprint(conn, namePrompt)
	}
}
