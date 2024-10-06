package client

import (
	"bufio"
	"fmt"
	"github.com/gucardona/ga-redes-udp/src/ebpf/cpu"
	"net"
	"os"
	"strings"
	"time"
)

func StartClient(serverPort int, messageInterval time.Duration) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select a metric to send (cpu/mem/gpu):")
	metricType, _ := reader.ReadString('\n')
	metricType = strings.TrimSpace(metricType)

	var collectMetric func() (string, error)

	switch metricType {
	case "cpu":
		// Initialize CPU metrics collection
		if err := cpu.InitCPUMetricsCollection(); err != nil {
			return fmt.Errorf("error initializing CPU metrics: %s", err)
		}
		collectMetric = cpu.CollectCPUMetrics
	//case "mem":
	//	collectMetric = metrics.CollectMemoryMetrics
	//case "gpu":
	//	collectMetric = metrics.CollectGPUMetrics
	default:
		fmt.Println("Invalid metric type, defaulting to CPU metrics")
		if err := cpu.InitCPUMetricsCollection(); err != nil {
			return fmt.Errorf("error initializing CPU metrics: %s", err)
		}
		collectMetric = cpu.CollectCPUMetrics
	}

	serverAddr := net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &serverAddr)
	if err != nil {
		return fmt.Errorf("error starting client: %s", err)
	}
	defer conn.Close()

	for {
		// Collect the metric (this can be CPU, memory, GPU)
		message, err := collectMetric()
		if err != nil {
			fmt.Println("Error collecting metric:", err)
			continue
		}

		// Send the collected metric over UDP
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending data:", err)
			continue
		}
		fmt.Println("Metrics sent:", message)

		// Sleep for the interval before sending the next metric
		time.Sleep(messageInterval)
	}
}
