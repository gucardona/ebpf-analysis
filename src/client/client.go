package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	discoveryPort = 9999
)

// StartClient initializes the client, registers it with the server, and starts sending metrics.
func StartClient(serverPort int, messageInterval time.Duration) error {
	discoveryAddr := net.UDPAddr{
		Port: discoveryPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	// Connect to the discovery server
	discoveryConn, err := net.DialUDP("udp", nil, &discoveryAddr)
	if err != nil {
		return fmt.Errorf("error connecting to discovery server: %s", err)
	}
	defer discoveryConn.Close()

	// Register with the discovery server
	portInfo := fmt.Sprintf("REGISTER: %d", serverPort)
	_, err = discoveryConn.Write([]byte(portInfo))
	if err != nil {
		return fmt.Errorf("error sending registration: %s", err)
	}

	// Connect to the main server
	serverAddr := net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &serverAddr)
	if err != nil {
		return fmt.Errorf("error connecting to server: %s", err)
	}
	defer conn.Close()

	// Listen for responses from the discovery server
	go listenForDiscoveryResponses(discoveryConn)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select a metric to send (cpu/mem/gpu):")
	metricType, _ := reader.ReadString('\n')
	metricType = strings.TrimSpace(metricType)

	for {
		metrics, err := collectMetrics(metricType)
		if err != nil {
			fmt.Println("Error collecting metrics:", err)
			continue
		}

		_, err = conn.Write(metrics)
		if err != nil {
			return fmt.Errorf("error sending data: %s", err)
		}

		time.Sleep(messageInterval)
	}
}

// listenForDiscoveryResponses listens for any messages from the discovery server.
func listenForDiscoveryResponses(conn *net.UDPConn) {
	buf := make([]byte, 2048)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving from discovery server:", err)
			return // Exit on error
		}
		clientInfo := string(buf[:n])
		fmt.Printf("Received from discovery server: %s\n", clientInfo)
	}
}

// collectMetrics gathers metrics based on the specified type.
func collectMetrics(metricType string) ([]byte, error) {
	switch metricType {
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

	default:
		return nil, fmt.Errorf("unknown metric type: %s", metricType)
	}
}
