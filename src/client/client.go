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

const discoveryPort = 9999

func StartClient(serverPort int, messageInterval time.Duration) error {
	discoveryAddr := net.UDPAddr{
		Port: discoveryPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	discoveryConn, err := net.DialUDP("udp", nil, &discoveryAddr)
	if err != nil {
		return fmt.Errorf("error connecting to discovery server: %s", err)
	}
	defer discoveryConn.Close()

	portInfo := fmt.Sprintf("REGISTER: %d", serverPort)
	_, err = discoveryConn.Write([]byte(portInfo))
	if err != nil {
		return fmt.Errorf("error sending registration: %s", err)
	}

	clients := make(map[string]*net.UDPAddr)

	go func() {
		buf := make([]byte, 2048)
		for {
			n, addr, err := discoveryConn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error receiving from discovery server:", err)
				continue
			}

			clientInfo := string(buf[:n])
			if clientInfo != "" && clientInfo != portInfo {
				clients[addr.String()] = addr
				fmt.Printf("Discovered client at: %s\n", addr.String())
			}
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
