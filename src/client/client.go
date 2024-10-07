package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gucardona/ga-redes-udp/src/server"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	HeartbeatInterval = 10 * time.Second
)

func StartClient(serverPort int, messageInterval time.Duration) error {
	discoveryAddr := net.UDPAddr{
		Port: server.DiscoveryPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	connDiscovery, err := net.DialUDP("udp", nil, &discoveryAddr)
	if err != nil {
		return fmt.Errorf("error connecting to discovery server: %s", err)
	}
	defer connDiscovery.Close()

	registerMessage := struct {
		Type string `json:"type"`
		Port int    `json:"port"`
	}{
		Type: "register",
		Port: serverPort,
	}

	data, err := json.Marshal(registerMessage)
	if err != nil {
		return fmt.Errorf("error marshaling register message: %s", err)
	}

	_, err = connDiscovery.Write(data)
	if err != nil {
		return fmt.Errorf("error sending register message: %s", err)
	}

	// Start heartbeat goroutine
	go func() {
		heartbeatMessage := struct {
			Type string `json:"type"`
			Port int    `json:"port"`
		}{
			Type: "heartbeat",
			Port: serverPort,
		}

		for {
			data, _ := json.Marshal(heartbeatMessage)
			connDiscovery.Write(data)
			time.Sleep(HeartbeatInterval)
		}
	}()

	serverAddr := net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &serverAddr)
	if err != nil {
		return fmt.Errorf("error connecting to server: %s", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select a metric to send (cpu/mem/gpu):")
	fmt.Println(" - cpu: The command is continuously counting how many times different processes are scheduled by the Linux kernel.")
	metricType, _ := reader.ReadString('\n')
	metricType = strings.TrimSpace(metricType)

	for {
		metrics, err := collectMetrics(metricType)
		if err != nil {
			fmt.Println("Error collecting metrics:", err)
		}

		_, err = conn.Write(metrics)
		if err != nil {
			return fmt.Errorf("error sending data: %s", err)
		}

		time.Sleep(messageInterval)
	}
}

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
