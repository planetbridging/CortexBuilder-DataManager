package main

import (
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

func (h *Hub) ReconnectClient(client *Client) {
	for {
		conn, err := net.Dial("tcp", client.Addr)
		if err == nil {
			client.Conn = conn
			client.LastSeen = time.Now()
			h.AddClient(client)
			go handleConnection(h, client)
			fmt.Printf("Reconnected to client: %s\n", client.Addr)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (h *Hub) MonitorClients() {
	for {
		time.Sleep(1 * time.Second)
		for addr, client := range h.clients {
			if time.Since(client.LastSeen) > 5*time.Second {
				client.Conn.Close()
				h.RemoveClient(addr)
				go h.ReconnectClient(client)
			}
		}
	}
}

func handleConnection(hub *Hub, client *Client) {
	defer client.Conn.Close()

	addr := client.Conn.RemoteAddr().String()

	buf := make([]byte, 1024)
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

func startTcpServer() {
	hub := NewHub()
	go hub.Run()

	listener, err := net.Listen("tcp", ":12345")
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
		go handleConnection(hub, client)
	}
}
