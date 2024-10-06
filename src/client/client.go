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

func StartClient(serverPort int, messageInterval time.Duration) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select a metric to send (cpu/mem/gpu):")
	fmt.Println(" - cpu: The command is continuously counting how many times different processes are scheduled by the Linux kernel.")
	metricType, _ := reader.ReadString('\n')
	metricType = strings.TrimSpace(metricType)

	serverAddr := net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &serverAddr)
	if err != nil {
		return fmt.Errorf("error starting client: %s", err)
	}
	defer conn.Close()

	switch metricType {
	case "cpu":
		go func() {
			for {
				out, err := exec.Command(
					"sudo",
					"bpftrace",
					"-e",
					"kprobe:schedule { @[comm] = count(); } interval:s:1 { print(@); clear(@); exit(); }").Output()
				if err != nil {
					fmt.Printf("failed to exec command: %s", err)
					continue
				}

				if err := sendUDP(conn, out); err != nil {
					fmt.Printf("failed to send udp data: %s", err)
					continue
				}

				time.Sleep(messageInterval)
			}
		}()

	default:
		log.Fatal("Invalid metric type...")
	}

	select {}
}

func sendUDP(conn net.Conn, metrics []byte) error {
	if _, err := conn.Write(metrics); err != nil {
		return fmt.Errorf("error sending data: %s", err)
	}

	return nil
}
