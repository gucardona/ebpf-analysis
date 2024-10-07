package server

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	DiscoveryPort     = 9999
	HeartbeatInterval = 10 * time.Second
	ClientTimeout     = 30 * time.Second
)

type Client struct {
	Port          int
	LastHeartbeat time.Time
}

type DiscoveryServer struct {
	clients map[int]*Client
	mutex   sync.RWMutex
}

func NewDiscoveryServer() *DiscoveryServer {
	return &DiscoveryServer{
		clients: make(map[int]*Client),
	}
}

func (ds *DiscoveryServer) Start() error {
	addr := net.UDPAddr{
		Port: DiscoveryPort,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("error starting discovery server: %s", err)
	}
	defer conn.Close()

	go ds.cleanupClients()

	fmt.Println("Discovery server started. Listening for messages...")

	buf := make([]byte, 2048)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving discovery message:", err)
			continue
		}

		var message struct {
			Type string `json:"type"`
			Port int    `json:"port"`
		}

		if err := json.Unmarshal(buf[:n], &message); err != nil {
			fmt.Println("Error parsing message:", err)
			continue
		}

		switch message.Type {
		case "register":
			ds.registerClient(message.Port)
			ds.sendClientList(conn, remoteAddr)
		case "heartbeat":
			ds.updateHeartbeat(message.Port)
		}
	}
}

func (ds *DiscoveryServer) registerClient(port int) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.clients[port] = &Client{
		Port:          port,
		LastHeartbeat: time.Now(),
	}

	fmt.Printf("New client registered: %d\n", port)
}

func (ds *DiscoveryServer) updateHeartbeat(port int) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if client, ok := ds.clients[port]; ok {
		client.LastHeartbeat = time.Now()
	}
}

func (ds *DiscoveryServer) sendClientList(conn *net.UDPConn, addr *net.UDPAddr) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	clientList := make([]int, 0, len(ds.clients))
	for port := range ds.clients {
		clientList = append(clientList, port)
	}

	message := struct {
		Type    string `json:"type"`
		Clients []int  `json:"clients"`
	}{
		Type:    "client_list",
		Clients: clientList,
	}

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling client list:", err)
		return
	}

	_, err = conn.WriteToUDP(data, addr)
	if err != nil {
		fmt.Println("Error sending client list:", err)
	}
}

func (ds *DiscoveryServer) cleanupClients() {
	for {
		time.Sleep(HeartbeatInterval)

		ds.mutex.Lock()
		for port, client := range ds.clients {
			if time.Since(client.LastHeartbeat) > ClientTimeout {
				delete(ds.clients, port)
				fmt.Printf("Client timed out and removed: %d\n", port)
			}
		}
		ds.mutex.Unlock()
	}
}
