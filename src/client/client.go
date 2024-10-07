package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	DiscoveryPort     = 9999
	HeartbeatInterval = 10 * time.Second
)

type Client struct {
	serverPort      int
	messageInterval time.Duration
	discoveryConn   *net.UDPConn
	serverConn      *net.UDPConn
	metricType      string
}

func StartClient(serverPort int, messageInterval time.Duration) error {
	client := &Client{
		serverPort:      serverPort,
		messageInterval: messageInterval,
	}

	if err := client.connectToDiscoveryServer(); err != nil {
		return err
	}

	if err := client.connectToMainServer(); err != nil {
		return err
	}

	if err := client.selectMetricType(); err != nil {
		return err
	}

	go client.sendHeartbeats()

	return client.runMetricLoop()
}

func (c *Client) connectToDiscoveryServer() error {
	discoveryAddr := net.UDPAddr{
		Port: DiscoveryPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &discoveryAddr)
	if err != nil {
		return fmt.Errorf("error connecting to discovery server: %s", err)
	}
	c.discoveryConn = conn

	return c.registerWithDiscoveryServer()
}

func (c *Client) registerWithDiscoveryServer() error {
	registerMessage := struct {
		Type string `json:"type"`
		Port int    `json:"port"`
	}{
		Type: "register",
		Port: c.serverPort,
	}

	data, err := json.Marshal(registerMessage)
	if err != nil {
		return fmt.Errorf("error marshaling register message: %s", err)
	}

	_, err = c.discoveryConn.Write(data)
	if err != nil {
		return fmt.Errorf("error sending register message: %s", err)
	}

	fmt.Println("Registered with discovery server")
	return nil
}

func (c *Client) connectToMainServer() error {
	serverAddr := net.UDPAddr{
		Port: c.serverPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &serverAddr)
	if err != nil {
		return fmt.Errorf("error connecting to main server: %s", err)
	}
	c.serverConn = conn

	fmt.Println("Connected to main server")
	return nil
}

func (c *Client) selectMetricType() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select a metric to send (cpu/mem/gpu):")
	fmt.Println(" - cpu: The command is continuously counting how many times different processes are scheduled by the Linux kernel.")
	metricType, _ := reader.ReadString('\n')
	c.metricType = strings.TrimSpace(metricType)

	return nil
}

func (c *Client) sendHeartbeats() {
	heartbeatMessage := struct {
		Type string `json:"type"`
		Port int    `json:"port"`
	}{
		Type: "heartbeat",
		Port: c.serverPort,
	}

	for {
		data, _ := json.Marshal(heartbeatMessage)
		c.discoveryConn.Write(data)
		time.Sleep(HeartbeatInterval)
	}
}

func (c *Client) runMetricLoop() error {
	for {
		metrics, err := c.collectMetrics()
		if err != nil {
			fmt.Println("Error collecting metrics:", err)
			continue
		}

		if err := c.sendMetrics(metrics); err != nil {
			fmt.Println("Error sending metrics:", err)
		}

		time.Sleep(c.messageInterval)
	}
}

func (c *Client) collectMetrics() ([]byte, error) {
	switch c.metricType {
	case "cpu":
		out, err := exec.Command(
			"sudo",
			"bpftrace",
			"-e",
			"kprobe:schedule { @[comm] = count(); } interval:s:1 { print(@); clear(@); exit(); }").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to exec command: %s", err)
		}
		return out, nil
	case "mem":
		// Implement memory metric collection
		return []byte("Memory metrics not implemented yet"), nil
	case "gpu":
		// Implement GPU metric collection
		return []byte("GPU metrics not implemented yet"), nil
	default:
		return nil, fmt.Errorf("unknown metric type: %s", c.metricType)
	}
}

func (c *Client) sendMetrics(metrics []byte) error {
	message := struct {
		Type    string `json:"type"`
		Metrics string `json:"metrics"`
	}{
		Type:    "metrics",
		Metrics: string(metrics),
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling metrics message: %s", err)
	}

	_, err = c.serverConn.Write(data)
	if err != nil {
		return fmt.Errorf("error sending metrics: %s", err)
	}

	return nil
}
