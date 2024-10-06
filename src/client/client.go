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

	// Start a goroutine to read and send metrics periodically
	go func() {
		scanner := bufio.NewScanner(stdout)
		var accumulatedMetrics strings.Builder

		for {
			if scanner.Scan() {
				line := scanner.Text()
				// Filter out "Attaching probes..." and empty lines
				if strings.Contains(line, "Attaching probes") || strings.TrimSpace(line) == "" {
					continue
				}

				// Format the line to be aligned
				// Assuming the format is @[name]: count
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					metricName := strings.TrimSpace(parts[0])
					metricCount := strings.TrimSpace(parts[1])
					message := fmt.Sprintf("%-25s: %s\n", metricName, metricCount)

					// Accumulate the metrics
					accumulatedMetrics.WriteString(message)
				}
			}

			// After collecting metrics for the interval, send the accumulated metrics
			if accumulatedMetrics.Len() > 0 {
				// Send the accumulated metrics over UDP
				if _, err := conn.Write([]byte(accumulatedMetrics.String())); err != nil {
					fmt.Println("Error sending data:", err)
					return
				}

				// Clear the accumulated metrics for the next interval
				accumulatedMetrics.Reset()
			}

			// Sleep for the message interval
			time.Sleep(messageInterval)
		}
	}()

	// Wait for the BPFtrace command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for command completion: %s", err)
	}

	return nil
}
