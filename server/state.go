package server

import "sync"

var (
	clients         []*Client
	connectionCount int
	mu              sync.Mutex
)
