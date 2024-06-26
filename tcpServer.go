package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
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

		//var response []byte

		var js map[string]interface{}
		errCheckJson := json.Unmarshal([]byte(message), &js)
		fmt.Println(js)

		if errCheckJson != nil {
			switch message {
			case "ping":

				// Get OS info
				osInfo := runtime.GOOS

				// Get RAM info
				vmStat, err := mem.VirtualMemory()
				if err != nil {
					log.Fatalf("Error getting memory info: %v", err)
				}
				ramInfo := fmt.Sprintf("%.2fGB", float64(vmStat.Total)/(1024*1024*1024))

				// Get CPU info
				cpuInfo, err := cpu.Info()
				if err != nil {
					log.Fatalf("Error getting CPU info: %v", err)
				}
				cpuModel := ""
				if len(cpuInfo) > 0 {
					cpuModel = fmt.Sprintf("%d cores %s", cpuInfo[0].Cores, cpuInfo[0].ModelName)
				}

				serverInfo := make(map[string]interface{})
				serverInfo["OS"] = osInfo
				serverInfo["RAM"] = ramInfo
				serverInfo["CPU"] = cpuModel
				serverInfo["ComputerType"] = "data"
				serverInfo["type"] = "serverInfo"

				response, err := json.Marshal(serverInfo)
				if err != nil {
					fmt.Printf("Error marshaling JSON: %v\n", err)
					return
				}

				client.Conn.Write(response)
				//default:
				//	response = []byte(`{"error": "unknown command"}`)
			}
		} else {
			if value, ok := js["type"]; ok {
				fmt.Println("Key exists, value: ", value)
				switch js["type"] {
				case "mountData":
					fmt.Println("mount", js)
					if path, ok := js["path"].(string); ok {
						if _, ok := js["clientID"]; ok {
							mountStatus := mountFile(path)
							fmt.Println(mountStatus)
							mountRes := make(map[string]interface{})
							mountRes["type"] = "mountStatus"
							mountRes["clientID"] = js["clientID"]
							mountRes["path"] = js["path"]
							mountRes["status"] = mountStatus
							response, err := json.Marshal(mountRes)
							if err != nil {
								fmt.Printf("Error marshaling JSON: %v\n", err)
								return
							}

							client.Conn.Write(response)
						} else {
							fmt.Println("clientID does not exist")
						}
					} else {
						fmt.Println("path does not exist or is not a string")
					}
					break
				}
			} else {
				fmt.Println("Key does not exist")
			}
		}

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
