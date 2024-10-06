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
			"kprobe:schedule { @[comm] = count(); } interval:s:1 { print(@); clear(@); }")

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

	// Prepare a slice to accumulate metrics
	var allMetrics []string

	// Start a goroutine to read and send metrics periodically
	go func() {
		scanner := bufio.NewScanner(stdout)

		for {
			if scanner.Scan() {
				// Read the latest output from bpftrace
				metric := scanner.Text()
				allMetrics = append(allMetrics, metric) // Accumulate metrics

				// Format the accumulated metrics for sending
				metricsToSend := formatMetrics(allMetrics)

				// Send the accumulated metrics over UDP
				if _, err := conn.Write([]byte(metricsToSend)); err != nil {
					fmt.Println("Error sending data:", err)
					return
				}

				fmt.Println("Metrics sent:\n", metricsToSend)

				// Sleep for the message interval
				time.Sleep(messageInterval)
			} else {
				// Exit if there's an error reading from stdout
				if err := scanner.Err(); err != nil {
					fmt.Println("Error reading from stdout:", err)
					break
				}
			}
		}
	}()

	// Wait for the bpftrace command to complete (if it ever does)
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for command completion: %s", err)
	}

	return nil
}

// formatMetrics formats the accumulated metrics into a string
func formatMetrics(metrics []string) string {
	var formatted strings.Builder
	for _, metric := range metrics {
		formatted.WriteString(metric + "\n") // Add each metric on a new line
	}
	return formatted.String()
}
