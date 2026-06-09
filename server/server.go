package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

type Client struct {
	conn   net.Conn
	name   string
	writer *bufio.Writer
}

var (
	connectionCount int
	mu              sync.Mutex
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
	"     `-'       `--'\n" +
	"[ENTER YOUR NAME]: "

func Start(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}

		mu.Lock()
		if connectionCount >= 10 {
			mu.Unlock()
			conn.Write([]byte("Server is full. Maximum 10 connections reached.\n"))
			conn.Close()
			continue
		}
		connectionCount++
		mu.Unlock()

		if err := sendWelcomePrompt(conn); err != nil {
			log.Println("Welcome message error:", err)
			conn.Close()
			continue
		}
		name, err := readClientName(conn)
		if err != nil {
			log.Println("Name read error:", err)
			conn.Close()
			continue
		}
		_ = name
	}
}

func sendWelcomePrompt(conn net.Conn) error {
	_, err := fmt.Fprint(conn, welcomeMessage)
	return err
}
