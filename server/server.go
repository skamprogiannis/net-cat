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
	clients         []*Client
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
	"     `-'       `--'\n"

const namePrompt = "[ENTER YOUR NAME]: "

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
			mu.Lock()
			connectionCount--
			mu.Unlock()
			conn.Close()
			continue
		}
		name, err := readClientName(conn)
		if err != nil {
			log.Println("Name read error:", err)
			mu.Lock()
			connectionCount--
			mu.Unlock()
			conn.Close()
			continue
		}
		go handleClient(conn, name)
	}
}

func sendWelcomePrompt(conn net.Conn) error {
	_, err := fmt.Fprint(conn, welcomeMessage, namePrompt)
	return err
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
