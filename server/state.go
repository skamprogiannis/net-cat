package server

import "sync"

var (
	clients         []*Client
	messageHistory []string
	connectionCount int
	mu              sync.Mutex
)
