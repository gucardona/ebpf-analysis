package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

// StartClient starts the UDP client to send metrics to the server
func StartClient(serverPort int, messageInterval time.Duration) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select a metric to send (cpu/mem/gpu):")
	fmt.Println(" - cpu: The command is continuously counting how many times different processes are scheduled by the Linux kernel.")
	metricType, _ := reader.ReadString('\n')
	metricType = strings.TrimSpace(metricType)

	var cmd *exec.Cmd

	switch metricType {
	case "cpu":
		cmd = exec.Command(
			"sudo",
			"bpftrace",
			"-e",
			"kprobe:schedule { @[comm] = count(); } interval:s:1 { print(@); clear(@); exit(); }")

	default:
		log.Fatal("Invalid metric type...")
	}

	// Get the output pipe for real-time metric collection
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %s", err)
	}

	// Start the BPFtrace command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start BPFtrace: %s", err)
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

	// Read all lines from bpftrace output
	scanner := bufio.NewScanner(stdout)
	var allMetrics strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		allMetrics.WriteString(line + "\n")
	}

	// Send the accumulated metrics over UDP
	message := allMetrics.String()
	if _, err = conn.Write([]byte(message)); err != nil {
		fmt.Println("Error sending data:", err)
		return err
	}
	fmt.Println("Metrics sent:\n", message)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for command completion: %s", err)
	}

	return nil
}
