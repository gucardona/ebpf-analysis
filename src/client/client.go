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

var ServerRegisteredClients []int

func StartClient(serverPort int, clientPort int, messageInterval time.Duration) error {
	discoveryAddr := net.UDPAddr{
		Port: 9999,
		IP:   net.ParseIP("127.0.0.1"),
	}

	serverAddr := &net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("127.0.0.1"),
	}

	clientAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: clientPort,
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Select a metric to send:")
	fmt.Println(" - schp: Tracks the number of times different processes are scheduled by the Linux kernel.")
	fmt.Println(" - packet: Monitors how many packets are sent/received at the network interface level.")
	fmt.Println(" - data: Tracks the total amount of data (in bytes) transmitted by network devices.")
	fmt.Println(" - rtime: This metric provides insights into how much runtime each process is utilizing, in ns.")
	fmt.Println(" - read: This metric tracks the number of times the read system call is invoked by different processes running on the system.")
	fmt.Println(" - write: This metric tracks the number of times the write system call is invoked by different processes running on the system.")
	metricType, _ := reader.ReadString('\n')
	metricType = strings.TrimSpace(metricType)

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

	conn, err := net.DialUDP("udp", clientAddr, serverAddr)
	if err != nil {
		return fmt.Errorf("error connecting to server: %s", err)
	}
	defer conn.Close()

	for {
		metrics, err := collectMetrics(metricType)
		if err != nil {
			fmt.Println("Error collecting metrics:", err)
		}

		_, err = conn.Write(metrics)
		if err != nil {
			return fmt.Errorf("error sending data: %s", err)
		}

		for _, registeredServerPort := range ServerRegisteredClients {
			if registeredServerPort != serverPort {
				forwardAddr := &net.UDPAddr{
					Port: registeredServerPort,
					IP:   net.ParseIP("127.0.0.1"),
				}

				_, err := conn.WriteToUDP(metrics, forwardAddr)
				if err != nil {
					fmt.Printf("Error sending data to client %d: %s\n", registeredServerPort, err)
				}
			}
		}

		time.Sleep(messageInterval)
	}
}

func collectMetrics(metricType string) ([]byte, error) {
	switch metricType {
	case "schp":
		out, err := exec.Command(
			"sudo",
			"bpftrace",
			"-e",
			"kprobe:schedule { @[comm] = count(); } interval:s:1 { print(@); clear(@); exit(); }").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to exec command: %s", err)
		}

		metricTypeStr := ":T:SCHEDULE_METRIC"
		metricTypeBytes := []byte(metricTypeStr)

		appendedResult := append(out, metricTypeBytes...)
		return appendedResult, nil

	case "packet":
		out, err := exec.Command(
			"sudo",
			"bpftrace",
			"-e",
			"tracepoint:net:netif_receive_skb { @[comm] = count(); } interval:s:1 { print(@); clear(@); exit(); }").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to exec command: %s", err)
		}

		metricTypeStr := ":T:PACKET_METRIC"
		metricTypeBytes := []byte(metricTypeStr)

		appendedResult := append(out, metricTypeBytes...)
		return appendedResult, nil

	case "data":
		out, err := exec.Command(
			"sudo",
			"bpftrace",
			"-e",
			"tracepoint:net:net_dev_xmit { @[comm] = sum(args->len); } interval:s:1 { print(@); clear(@); exit(); }").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to exec command: %s", err)
		}

		metricTypeStr := ":T:DATA_METRIC"
		metricTypeBytes := []byte(metricTypeStr)

		appendedResult := append(out, metricTypeBytes...)
		return appendedResult, nil

	case "rtime":
		out, err := exec.Command(
			"sudo",
			"bpftrace",
			"-e",
			"tracepoint:sched:sched_stat_runtime { @[comm] = sum(args->runtime); } interval:s:1 { print(@); clear(@); exit(); }").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to exec command: %s", err)
		}

		metricTypeStr := ":T:RTIME_METRIC"
		metricTypeBytes := []byte(metricTypeStr)

		appendedResult := append(out, metricTypeBytes...)
		return appendedResult, nil

	case "read":
		out, err := exec.Command(
			"sudo",
			"bpftrace",
			"-e",
			"tracepoint:syscalls:sys_enter_read { @[comm] = count(); } interval:s:1 { print(@); clear(@); exit(); }").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to exec command: %s", err)
		}

		metricTypeStr := ":T:READ_METRIC"
		metricTypeBytes := []byte(metricTypeStr)

		appendedResult := append(out, metricTypeBytes...)
		return appendedResult, nil

	case "write":
		out, err := exec.Command(
			"sudo",
			"bpftrace",
			"-e",
			"tracepoint:syscalls:sys_enter_write { @[comm] = count(); } interval:s:1 { print(@); clear(@); exit(); }").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to exec command: %s", err)
		}

		metricTypeStr := ":T:WRITE_METRIC"
		metricTypeBytes := []byte(metricTypeStr)

		appendedResult := append(out, metricTypeBytes...)
		return appendedResult, nil

	default:
		return nil, fmt.Errorf("unknown metric type: %s", metricType)
	}
}
