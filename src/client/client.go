package client

import (
	"bufio"
	"fmt"
	"github.com/gucardona/ga-redes-udp/src/server"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

var ClientAddr *net.UDPAddr

func StartClient(serverPort int, clientPort int, messageInterval time.Duration) error {
	discoveryAddr := net.UDPAddr{
		Port: server.DiscoveryPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	connDiscovery, err := net.DialUDP("udp", nil, &discoveryAddr)
	if err != nil {
		return fmt.Errorf("error connecting to discovery server: %s", err)
	}
	defer connDiscovery.Close()

	registerMessage := []byte(fmt.Sprintf("register-%d", serverPort))

	_, err = connDiscovery.Write(registerMessage)
	if err != nil {
		return fmt.Errorf("error sending register message: %s", err)
	}

	go func() {
		for {
			connDiscovery.Write(registerMessage)
			time.Sleep(time.Second * 3)
		}
	}()

	serverAddr := net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	ClientAddr = &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: clientPort,
	}

	conn, err := net.DialUDP("udp", ClientAddr, &serverAddr)
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
