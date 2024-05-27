package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type ServerInfo struct {
	OS           string `json:"os"`
	RAM          string `json:"ram"`
	CPU          string `json:"cpu"`
	ComputerType string `json:"computerType"`
}

type Client struct {
	Conn     net.Conn
	Addr     string
	LastSeen time.Time
}

type Hub struct {
	addClientChan    chan *Client
	removeClientChan chan string
	clients          map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		addClientChan:    make(chan *Client),
		removeClientChan: make(chan string),
		clients:          make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.addClientChan:
			h.clients[client.Addr] = client
			fmt.Printf("Added client: %s\n", client.Addr)
		case addr := <-h.removeClientChan:
			delete(h.clients, addr)
			fmt.Printf("Removed client: %s\n", addr)
		}
	}
}

func (h *Hub) AddClient(client *Client) {
	h.addClientChan <- client
}

func (h *Hub) RemoveClient(addr string) {
	h.removeClientChan <- addr
}

func handleConnection(hub *Hub, client *Client, password string) {
	defer client.Conn.Close()

	// Authentication
	buf := make([]byte, 1024)
	n, err := client.Conn.Read(buf)
	if err != nil {
		hub.RemoveClient(client.Addr)
		return
	}

	if string(buf[:n]) != password {
		fmt.Printf("Client %s provided wrong password\n", client.Addr)
		hub.RemoveClient(client.Addr)
		return
	}

	client.Conn.Write([]byte("Authenticated"))

	addr := client.Conn.RemoteAddr().String()

	for {
		n, err := client.Conn.Read(buf)
		if err != nil {
			hub.RemoveClient(addr)
			return
		}

		client.LastSeen = time.Now()
		message := string(buf[:n])
		fmt.Printf("Received data from %s: %s\n", addr, message)

		var response []byte
		switch message {
		case "ping":
			serverInfo := ServerInfo{
				OS:           "Linux",
				RAM:          "16GB",
				CPU:          "4 cores",
				ComputerType: "data",
			}
			response, err = json.Marshal(serverInfo)
			if err != nil {
				fmt.Printf("Error marshaling JSON: %v\n", err)
				return
			}
		default:
			response = []byte(`{"error": "unknown command"}`)
		}

		client.Conn.Write(response)
	}
}

func startTcpServer(password string) {
	hub := NewHub()
	go hub.Run()

	tlsConfig, err := generateTLSConfig()
	if err != nil {
		fmt.Printf("Error generating TLS config: %v\n", err)
		return
	}

	listener, err := tls.Listen("tcp", ":12345", tlsConfig)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 12345...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		client := &Client{
			Conn: conn,
			Addr: conn.RemoteAddr().String(),
		}
		go handleConnection(hub, client, password)
	}
}